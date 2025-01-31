package detect

import (
	"net"

	"github.com/ryanolee/go-pot/config"
)

type (
	// Listener that binds itself to a multi protocol "detector"
	ConditionalListener struct {
		receiverChannel chan net.Conn
		shutdownChannel chan bool
		address         net.Addr
	}
)

func NewConditionalListenerFromConfig(config *config.Config) *ConditionalListener {
	return NewConditionalListener(
		&net.TCPAddr{
			IP:   net.ParseIP(config.MultiProtocol.Host),
			Port: config.MultiProtocol.Port,
		},
	)
}

func NewConditionalListener(address net.Addr) *ConditionalListener {
	return &ConditionalListener{
		receiverChannel: make(chan net.Conn),
		shutdownChannel: make(chan bool, 1),
		address:         address,
	}
}

func (l *ConditionalListener) Close() error {
	l.shutdownChannel <- true
	return nil
}

func (l *ConditionalListener) Accept() (net.Conn, error) {
	select {
	case <-l.shutdownChannel:
		return nil, net.ErrClosed
	case connection := <-l.receiverChannel:
		return connection, nil
	}
}

func (l *ConditionalListener) Addr() net.Addr {
	return l.address
}

func (l *ConditionalListener) Dispatch(conn net.Conn) {
	l.receiverChannel <- conn
}
