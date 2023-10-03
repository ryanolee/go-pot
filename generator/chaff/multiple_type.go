package chaff

import (
	"fmt"
	"strings"

	"github.com/thoas/go-funk"
)

type (
	MultipleTypeGenerator struct {
		generators []Generator
	}
)

func ParseMultipleType(node SchemaNode, metadata ParserMetadata) (MultipleTypeGenerator, error) {
	generators := []Generator{}
	for _, nodeType := range node.Type.MultipleTypes {
		generator, err := ParseType(nodeType, node, metadata)
		if err != nil {
			generators = append(generators, NullGenerator{})
		} else {
			generators = append(generators, generator)
		}
	}

	return MultipleTypeGenerator{
		generators: generators,
	}, nil
}

func (g MultipleTypeGenerator) Generate(opts *GeneratorOptions) interface{} {
	generator := g.generators[opts.Rand.RandomInt(0,len(g.generators))]
	return generator.Generate(opts)
}

func (g MultipleTypeGenerator) String() string {
	formattedGenerators := funk.Map(g.generators, func(generator Generator) string {
		return generator.String()
	}).([]string)

	return fmt.Sprintf("MultiGenerator{%s}", strings.Join(formattedGenerators, ","))
}
