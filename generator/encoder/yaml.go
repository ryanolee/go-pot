package encoder

import "gopkg.in/yaml.v2"

type YamlEncoder struct {}

func NewYamlEncoder() *YamlEncoder {
	return &YamlEncoder{}
}

func (e *YamlEncoder) Marshal(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)
}

func (*YamlEncoder) Start() string {
	return ""
}

func (*YamlEncoder) End() string {
	return ""
}

func (*YamlEncoder) Delimiter() string {
	return ""
}

func (*YamlEncoder) ContentType() string {
	return "application/x-yaml"
}

func (*YamlEncoder) GetSupportedGenerator() string {
	return "config"
}