package chaff

type (
	NullGenerator struct {
	}
)

func ParseNull(node SchemaNode) (NullGenerator, error) {
	return NullGenerator{}, nil
}

func (g NullGenerator) Generate(opts *GeneratorOptions) interface{} {
	return nil
}

func (g NullGenerator) String() string {
	return "NullGenerator"
}
