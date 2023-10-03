package chaff

import (
	"errors"
	"fmt"
	"strings"

	"github.com/thoas/go-funk"
)

type (
	CombinationGenerator struct {
		Generators []Generator
		Type string
	}
)

func ParseCombination(node SchemaNode, metadata ParserMetadata) (Generator, error) {
	ref := metadata.ReferenceHandler
	if len(node.OneOf) == 0 && len(node.AnyOf) == 0{
		return NullGenerator{}, errors.New("no items specified for oneOf / anyOf")
	}

	if len(node.OneOf) > 0 && len(node.AnyOf) > 0 {
		return NullGenerator{}, errors.New("only one of [oneOf / anyOf] can be specified")
	}

	target := node.OneOf
	nodeType := "oneOf"
	if len(node.AnyOf) > 0 {
		target = node.AnyOf
		nodeType = "anyOf"
	}

	generators := []Generator{}
	for i, node := range target {
		refPath := fmt.Sprintf("/%s/%d", nodeType, i)
		generator, err := ref.ParseNodeInScope(refPath, node, metadata)
		if err != nil {
			generators = append(generators, NullGenerator{})
		} else {
			generators = append(generators, generator)
		}
	}

	return CombinationGenerator{
		Generators: generators,
		Type: nodeType,
	}, nil
}

func (g CombinationGenerator) Generate(opts *GeneratorOptions) interface{} {
	// Select a random generator
	generator := g.Generators[opts.Rand.RandomInt(0,len(g.Generators))]
	return generator.Generate(opts)
}

func (g CombinationGenerator) String() string {
	formattedGenerators := funk.Map(g.Generators, func(generator Generator) string {
		return generator.String()
	}).([]string)
	return fmt.Sprintf("CombinationGenerator[%s]{%s}",g.Type, strings.Join(formattedGenerators, ","))
}