package encoder

import (
	"bytes"
	"encoding/json"

	"gopkg.in/ini.v1"
)

type IniEncoder struct{}

func NewIniEncoder() *IniEncoder {
	return &IniEncoder{}
}

func (*IniEncoder) Start() string {
	return ""
}

func (*IniEncoder) End() string {
	return ""
}

func (*IniEncoder) Delimiter() string {
	return ""
}

func (*IniEncoder) ContentType() string {
	return "text/plain"
}

func ContentType() string {
	return "text/plain"
}

func (*IniEncoder) GetSupportedGenerator() string {
	return "config"
}

func (e *IniEncoder) Marshal(v interface{}) ([]byte, error) {
	file := ini.Empty()
	if value, ok := v.(map[string]interface{}); ok {
		mapUnknownIniSections(file, value)
	}
	buffer := new(bytes.Buffer)
	_, err := file.WriteTo(buffer)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func mapUnknownIniSections(file *ini.File, value map[string]interface{}) {
	for sectionName, sectionValue := range value {
		section, err := file.NewSection(sectionName)
		if err != nil {
			continue
		}

		if data, ok := sectionValue.(map[string]interface{}); ok {
			mapUnknownValuesToIniSection(section, data)
		}
	}
}

func mapUnknownValuesToIniSection(section *ini.Section, value map[string]interface{}) {
	for key, value := range value {
		bytes, err := json.Marshal(value)

		if err != nil {
			continue
		}

		if _, err := section.NewKey(key, string(bytes)); err != nil {
			continue
		}
	}
}
