package encoder

import (
	"bytes"

	"github.com/BurntSushi/toml"
)

type TomlEncoder struct {}

func NewTomlEncoder() *TomlEncoder {
	return &TomlEncoder{}
}

func (e *TomlEncoder) Marshal(v interface{}) ([]byte, error) {
	if v == nil {
		return nil, nil
	}

	buffer := new(bytes.Buffer)
	encoder := toml.NewEncoder(buffer)
	err := encoder.Encode(v)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}


func (*TomlEncoder) Start() string {
	return ""
}

func (*TomlEncoder) End() string {
	return ""
}

func (*TomlEncoder) Delimiter() string {
	return ""
}

func (*TomlEncoder) GetSupportedGenerator() string {
	return "config"
}

func (*TomlEncoder) ContentType() string {
	return "application/toml"
}