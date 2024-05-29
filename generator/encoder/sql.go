package encoder

import (
	"strings"

	"github.com/ryanolee/ryan-pot/generator/source"
)

type SqlEncoder struct {}

func NewSqlEncoder() *SqlEncoder {
	return &SqlEncoder{}
}

func (*SqlEncoder) Start() string {
	return "INSERT INTO `UserRecords` (`" + strings.Join(source.GetTabularHeaderFields(), "`, `") + "`) VALUES (\n"
}

func (*SqlEncoder) End() string {
	return ")"
}

func (*SqlEncoder) Delimiter() string {
	return ",\n"
}

func (*SqlEncoder) ContentType() string {
	return "text/plain"
}

func (*SqlEncoder) GetSupportedGenerator() string {
	return "tabular"
}

func (e *SqlEncoder) Marshal(v interface{}) ([]byte, error) {
	return []byte("(`" + strings.Join(source.GetTabularFieldValues(), "`, `") + "`)"), nil
}

