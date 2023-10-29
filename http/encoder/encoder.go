package encoder

type Encoder interface {
	ContentType() string
	Start() string
	Marshal(interface{}) ([]byte, error)
	Delimiter() string
	End() string
}

