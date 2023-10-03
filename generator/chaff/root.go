package chaff

import (
	"fmt"
)

type (
	RootGenerator struct {
		Generator Generator
		// For any "$defs"
		Defs map[string]Generator
		// For any "definitions"
		Definitions map[string]Generator
		Metadata ParserMetadata
	}
)

func ParseRoot(node SchemaNode, metadata ParserMetadata) (RootGenerator, error) {
	
	def := ParseDefinitions("$defs", metadata, node.Defs)
	definitions := ParseDefinitions("definitions", metadata, node.Definitions)

	generator, err := ParseNode(node, metadata)

	return RootGenerator{
		Generator: generator,
		Defs: def,
		Definitions: definitions,
		Metadata: metadata,
	}, err
}

func ParseDefinitions(path string, metadata ParserMetadata, definitions map[string]SchemaNode) map[string]Generator {
	ref := metadata.ReferenceHandler
	generators := make(map[string]Generator)
	for key, value := range definitions {
		refPath := fmt.Sprintf("/%s/%s", path, key)
		generator, _ := ref.ParseNodeInScope(refPath, value, metadata)
		
		generators[key] = generator
	}

	return generators
}

func (g RootGenerator) Generate(opts *GeneratorOptions) interface{} {
	opts = WithGeneratorOptionsDefaults(*opts)
	return g.Generator.Generate(opts)
}

func (g RootGenerator) String() string {
	formattedString := ""
	for name, prop := range g.Definitions {
		formattedString += fmt.Sprintf("%s: %s,", name, prop)
	}

	formattedString += "$defs:"
	for name, prop := range g.Defs {
		formattedString += fmt.Sprintf("%s: %s,", name, prop)
	}

	return fmt.Sprintf("RootGenerator{Generator: %s Definitions: %s}", g.Generator, formattedString)
}