package encoder

import "encoding/json"


type JsonEncoder struct {}

func NewJsonEncoder() *JsonEncoder {
	return &JsonEncoder{}
}

func (e *JsonEncoder) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (*JsonEncoder) Start() string {
	return "["
}

func (*JsonEncoder) End() string {
	return "]"
}

func (*JsonEncoder) Delimiter() string {
	return ","
}

func (*JsonEncoder) ContentType() string {
	return "application/json"
}