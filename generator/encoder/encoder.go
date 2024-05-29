package encoder

type Encoder interface {
	GetSupportedGenerator() string
	ContentType() string
	Start() string
	Marshal(interface{}) ([]byte, error)
	Delimiter() string
	End() string
}

