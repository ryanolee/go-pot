package encoder

import (
	"bytes"
	"encoding/json"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type HclEncoder struct{}

func NewHclEncoder() *HclEncoder {
	return &HclEncoder{}
}

func (e *HclEncoder) Marshal(v interface{}) ([]byte, error) {
	f := hclwrite.NewEmptyFile()
	body := f.Body()

	if value, ok := v.(map[string]interface{}); ok {
		mapUnknownHclBlocks(body, value)
	}

	buffer := new(bytes.Buffer)

	if _, err := f.WriteTo(buffer); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (*HclEncoder) Start() string {
	return ""
}

func (*HclEncoder) End() string {
	return ""
}

func (*HclEncoder) Delimiter() string {
	return ""
}

func (*HclEncoder) ContentType() string {
	return "application/hcl"
}

func (*HclEncoder) GetSupportedGenerator() string {
	return "config"
}

func mapUnknownHclBlocks(file *hclwrite.Body, value map[string]interface{}) {
	for blockName, sectionValue := range value {
		block := file.AppendNewBlock(blockName, make([]string, 0))

		if data, ok := sectionValue.(map[string]interface{}); ok {
			mapUnknownValuesToHclBlock(block, data)
		}
	}
}

func mapUnknownValuesToHclBlock(block *hclwrite.Block, value map[string]interface{}) {
	body := block.Body()
	for key, value := range value {
		bytes, err := json.Marshal(value)

		if err != nil {
			continue
		}

		body.SetAttributeValue(key, cty.StringVal(string(bytes)))
	}
}
