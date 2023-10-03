package chaff

type (
	BooleanGenerator struct {}
)

func ParseBoolean(node SchemaNode) (BooleanGenerator, error) {
	return BooleanGenerator{}, nil
}

func (g BooleanGenerator) Generate(opts *GeneratorOptions) interface{} {
	return opts.Rand.RandomBool()
}

func (g BooleanGenerator) String() string {
	return "BooleanGenerator"
}
