package chaff

import "fmt"

type (
	EnumGenerator struct {
		Values []interface{}
	}
)

func ParseEnum(node SchemaNode) (EnumGenerator, error) {
	return EnumGenerator{
		Values: node.Enum,
	}, nil
}

func (g EnumGenerator) Generate(opts *GeneratorOptions) interface{} {
	return opts.Rand.Choice(g.Values)
}

func (g EnumGenerator) String() string {
	numberOfItemsInEnum := len(g.Values)
	return fmt.Sprintf("EnumGenerator[items: %d]", numberOfItemsInEnum)
}
