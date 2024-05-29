package encoder

import (
	"bytes"
	"encoding/csv"
	"strings"

	"github.com/ryanolee/ryan-pot/generator/source"
)

type CsvEncoder struct {}

func NewCsvEncoder() *CsvEncoder {
	return &CsvEncoder{}
}

func (e *CsvEncoder) Marshal(v interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	w := csv.NewWriter(buf)

	if err := w.Write(v.([]string)); err != nil {
		return nil, err
	}

	w.Flush()
	return buf.Bytes(), nil
}

func (*CsvEncoder) Start() string {
	return strings.Join(source.GetTabularHeaderFields(), ",") + "\n"
}

func (*CsvEncoder) End() string {
	return "\n"
}

func (*CsvEncoder) Delimiter() string {
	return "\n"
}

func (*CsvEncoder) ContentType() string {
	return "text/csv"
}

func (*CsvEncoder) GetSupportedGenerator() string {
	return "tabular"
}
