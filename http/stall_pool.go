package http

type (
	HttpStallerPool struct {
		stallers []*HttpStaller
	}

	HttpStallerPoolOptions struct {
		maximumConnections int
	}
)

func NewHttpStallerPool() *HttpStallerPool {
	return &HttpStallerPool{}
}
