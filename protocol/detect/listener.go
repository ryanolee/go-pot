package detect

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/ryanolee/go-pot/config"
	"github.com/ryanolee/go-pot/protocol/detect/detector"
	"github.com/thoas/go-funk"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type (
	MultiProtocolListener struct {
		port              int
		host              string
		protocolDetectors map[string]detector.ProtocolDetector
		protocolListeners map[string]*ConditionalListener
		shutdownChannel   chan bool
		listenerContext   context.Context
		listenerCancel    context.CancelFunc
		enableAll         bool
		logger            *zap.Logger
	}
)

const (
	initialReadTimeout = 2 * time.Second
	detectReadTimeout  = 6 * time.Second
	probeInterval      = 500 * time.Millisecond
)

func NewMulitProtocolListener(lf fx.Lifecycle, detectors []detector.ProtocolDetector, config *config.Config, logger *zap.Logger) *MultiProtocolListener {
	if !config.MultiProtocol.Enabled {
		return &MultiProtocolListener{}
	}

	protocolDetectors := make(map[string]detector.ProtocolDetector)
	protocolListeners := make(map[string]*ConditionalListener)
	enableAll := funk.ContainsString(config.MultiProtocol.Protocols, "all")

	for _, detector := range detectors {
		if !funk.ContainsString(config.MultiProtocol.Protocols, detector.ProtocolName()) && !enableAll {
			continue
		}
		protocolDetectors[detector.ProtocolName()] = detector
		protocolListeners[detector.ProtocolName()] = NewConditionalListenerFromConfig(config)

	}

	protocolListeners["fallback"] = NewConditionalListenerFromConfig(config)

	shutdownChannel := make(chan bool, 1)
	protocolListener := &MultiProtocolListener{
		protocolDetectors: protocolDetectors,
		protocolListeners: protocolListeners,
		shutdownChannel:   shutdownChannel,
		port:              config.MultiProtocol.Port,
		host:              config.MultiProtocol.Host,
		logger:            logger,
		enableAll:         enableAll,
	}

	lf.Append(fx.StopHook(func(ctx context.Context) error {
		protocolListener.Shutdown()
		return nil
	}))

	return protocolListener
}

func (l *MultiProtocolListener) Shutdown() {
	l.listenerCancel()
	for _, listener := range l.protocolListeners {
		listener.Close()
	}
}

func (l *MultiProtocolListener) Listen() error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", l.host, l.port))
	if err != nil {
		return err
	}

	listenerContext, cancel := context.WithCancel(context.Background())
	l.listenerCancel = cancel

	go func() {
		<-listenerContext.Done()
		listener.Close()
	}()

	for {

		conn, err := listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				l.logger.Info("listener closed shutting down...")
				return nil
			}

			l.logger.Error("error accepting connection", zap.Error(err))
			continue
		}

		go l.HandleConnection(listenerContext, conn)
	}
}

func (l *MultiProtocolListener) HandleConnection(ctx context.Context, conn net.Conn) {

	// Wait for any data to be available on the connection
	rewindableConn := newRewindableConnFromConn(conn)
	context, cancel := context.WithCancel(ctx)
	defer cancel()

	successful, err := l.attemptHandoff(context, rewindableConn, initialReadTimeout)

	// Attempt to handoff the connection to the associated listener
	if err != nil {
		cancel()
		l.logger.Error("Failed to read data from connection", zap.Error(err))
		return
	}

	// If the connection was not successfully handed off begin a probe
	// to determine the protocol
	if successful {
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		l.probe(context, rewindableConn)
		if err != nil {
			cancel()
		}
		wg.Done()
	}()

	go func() {
		successful, err := l.attemptHandoff(context, rewindableConn, detectReadTimeout)

		if err != nil {
			l.logger.Error("Failed to read data from connection during probe attempt", zap.Error(err))
			cancel()
		}

		if !successful {
			l.logger.Info("No data read from connection during probe attempt")
			l.HandoffConnection([]byte{}, rewindableConn)
		}

		wg.Done()
	}()

	wg.Wait()
}

// Writes various probes down the pipe for connections that require banners to be sent
func (l *MultiProtocolListener) probe(context context.Context, conn net.Conn) error {
	ticker := time.NewTicker(probeInterval)
	for _, probe := range l.protocolDetectors {
		if probe.GetProbe() == nil {
			continue
		}

		select {
		case <-context.Done():
			return nil
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(probeInterval))
			_, err := conn.Write(probe.GetProbe())
			if err != nil {
				l.logger.Error("Failed to write probe", zap.Error(err))
				return err
			}
		}
	}
	return nil
}

func (l *MultiProtocolListener) attemptHandoff(ctx context.Context, conn *rewindableConn, timeout time.Duration) (bool, error) {
	data := make([]byte, rewindBufferSize)

	conn.SetReadDeadline(time.Now().Add(timeout))
	n, err := conn.ReadWithTimeout(ctx, data, timeout)

	// If the connection errored out
	// not because of a timeout, close the connection
	if err != nil && !os.IsTimeout(err) {
		conn.Close()
		return false, err
	}

	// If the context was cancelled for any reason close the overall connection
	if errors.Is(ctx.Err(), context.Canceled) {
		conn.Close()
		return false, fmt.Errorf("Context cancelled")
	}

	// No data was read from the connection
	if n == 0 {
		return false, nil
	}

	l.HandoffConnection(data, conn)
	return true, nil
}

func (l *MultiProtocolListener) FindMatchingProtocol(data []byte) (string, error) {
	for name, listener := range l.protocolDetectors {
		if listener.IsMatch(data) {
			return name, nil
		}
	}

	return "", fmt.Errorf("No Protocol found")
}

func (l *MultiProtocolListener) ProtocolEnabled(protocol string) bool {
	_, ok := l.protocolDetectors[protocol]
	return l.enableAll || ok
}

func (l *MultiProtocolListener) GetListenerForProtocol(protocol string) *ConditionalListener {
	listener, ok := l.protocolListeners[protocol]
	if !ok {
		return l.protocolListeners["fallback"]
	}

	return listener
}

// Once we have any data commit to a decision and handoff the connection the
func (l *MultiProtocolListener) HandoffConnection(data []byte, conn *rewindableConn) {
	protocol, err := l.FindMatchingProtocol(data)
	if err != nil {
		l.logger.Info("error finding matching protocol reverting to fallback handler", zap.String("data", string(data)))
		protocol = "fallback"
	} else {
		l.logger.Info("found protocol handler for sent data", zap.String("protocol", protocol), zap.String("data", string(data)))
	}

	protocolListener := l.GetListenerForProtocol(protocol)

	// Reset the connection
	conn.Rewind()

	// Perform the handoff
	protocolListener.Dispatch(conn)
}
