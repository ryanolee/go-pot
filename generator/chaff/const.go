package chaff

import "fmt"

type (
	ConstGenerator struct {
		Value interface{}
	}
)

func ParseConst(node SchemaNode) (ConstGenerator, error) {
	return ConstGenerator{
		Value: node.Const,
	}, nil
}

func (g ConstGenerator) Generate(opts *GeneratorOptions) interface{} {
	return g.Value
}

func (g ConstGenerator) String() string {
	return fmt.Sprintf("ConstGenerator[%s]", g.Value)
}
