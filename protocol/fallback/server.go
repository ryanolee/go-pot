package fallback

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/ryanolee/go-pot/config"
	"github.com/ryanolee/go-pot/protocol/detect"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type (
	// Connection handler to use in the event no other protocols are matched
	fallbackProtocolServer struct {
		multiProtocolListener *detect.MultiProtocolListener
	}
)

const slowSendInterval = 2000 * time.Millisecond

var defaultMessage = []byte("‍")

func NewFallbackProtocolServer(config *config.Config, lf fx.Lifecycle, mpl *detect.MultiProtocolListener) (*fallbackProtocolServer, error) {
	if !config.MultiProtocol.Enabled {
		return nil, fmt.Errorf("MultiProtocol is not enabled")
	}

	cancelContext, cancel := context.WithCancel(context.Background())
	f := &fallbackProtocolServer{
		multiProtocolListener: mpl,
	}

	lf.Append(fx.Hook{
		OnStart: func(context context.Context) error {
			go func() {
				err := f.Start(cancelContext)
				if err != nil {
					zap.S().Errorf("Error starting fallback protocol server: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(context context.Context) error {
			cancel()
			return nil
		},
	})

	return f, nil
}

func (f *fallbackProtocolServer) Start(ctx context.Context) error {
	listener := f.multiProtocolListener.GetListenerForProtocol("fallback") // Send an infinite stream of Zero width joiners
	defer listener.Close()

	if listener == nil {
		return fmt.Errorf("Listener not found")
	}

	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	for {

		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		go f.handleConnection(ctx, conn)
	}
}

func (f *fallbackProtocolServer) handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	ticker := time.NewTicker(slowSendInterval)
	for {
		select {
		case <-ticker.C:
			_, err := conn.Write(defaultMessage)
			if err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
