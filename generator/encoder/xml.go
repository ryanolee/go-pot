package encoder

import (
	"encoding/xml"
	"fmt"
)

type XmlEncoder struct{}

type UnknownMap map[string]interface{}

func NewUnknownMap(data interface{}) UnknownMap {
	returnData := make(UnknownMap)
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil
	}

	for key, value := range dataMap {
		switch value.(type) {
		case map[string]interface{}:
			returnData[key] = NewUnknownMap(value)
		case []interface{}:
			for index, item := range value.([]interface{}) {
				returnData[fmt.Sprintf("%s.%d", key, index)] = NewUnknownMap(item)
			}
		default:
			returnData[key] = value
		}
	}

	return returnData
}
func (s UnknownMap) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	tokens := []xml.Token{start}

	for key, value := range s {
		t := xml.StartElement{Name: xml.Name{"", key}}
		if data, ok := value.(UnknownMap); ok {
			if err := data.MarshalXML(e, t); err != nil {
				return err
			}
		} else {
			tokens = append(tokens, t, xml.CharData(fmt.Sprintf("%v", value)), xml.EndElement{t.Name})
		}
	}

	tokens = append(tokens, xml.EndElement{start.Name})

	for _, t := range tokens {
		err := e.EncodeToken(t)
		if err != nil {
			return err
		}
	}

	// flush to ensure tokens are written
	return e.Flush()
}

func NewXmlEncoder() *XmlEncoder {
	return &XmlEncoder{}
}

func (e *XmlEncoder) Marshal(v interface{}) ([]byte, error) {
	data := NewUnknownMap(v)
	if data == nil {
		return nil, fmt.Errorf("failed to convert data to unknown map")
	}

	xml, err := xml.MarshalIndent(data, "", "  ")
	return xml, err
}

func (*XmlEncoder) Start() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes" ?><root>`
}

func (*XmlEncoder) End() string {
	return "</root>"
}

func (*XmlEncoder) Delimiter() string {
	return ""
}

func (*XmlEncoder) ContentType() string {
	return "application/xml"
}

func (*XmlEncoder) GetSupportedGenerator() string {
	return "config"
}
