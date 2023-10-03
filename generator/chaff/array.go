package chaff

import (
	"fmt"
)

type (
	ArrayGenerator struct {
		TupleGenerators []Generator
		ItemGenerator Generator
		
		MinItems int
		MaxItems int

		DisallowAdditional bool
	}
)

const (
	ItemsPathPart = "/items"
	ContainsPathPart = "/contains"
	TuplePathPart = "/tuple"
)

func ParseArray(node SchemaNode, metadata ParserMetadata) (Generator, error) {

	// Validate Bounds
	if node.MinItems > node.MaxItems {
		return NullGenerator{}, fmt.Errorf("minItems must be less than or equal to maxItems (minItems: %d, maxItems: %d)", node.MinItems, node.MaxItems)
	}

	if node.MinContains > node.MaxContains {
		return NullGenerator{}, fmt.Errorf("minContains must be less than or equal to maxContains (minContains: %d, maxContains: %d)", node.MinContains, node.MaxContains)
	}

	// Validate if tuple makes sense in this context
	tupleLength := len(node.PrefixItems)
	if tupleLength > node.MaxItems {
		return NullGenerator{}, fmt.Errorf("tuple length must be less than or equal to maxItems (tupleLength: %d, maxItems: %d)", tupleLength, node.MaxItems)
	}

	min := getInt(node.MinItems, node.MinContains)
	max := getInt(node.MaxItems, node.MaxContains)
	
	// Force the generator to use only the tuple in the event that additional items
	// are not allowed
	if (node.Items.DisallowAdditional){
		min = tupleLength
		max = tupleLength
	}

	return ArrayGenerator{
		TupleGenerators: parseTupleGenerator(node.PrefixItems, metadata),
		ItemGenerator: parseItemGenerator(node.Items, metadata),

		MinItems: min,
		MaxItems: max,
		DisallowAdditional: node.Items.DisallowAdditional,
	}, nil
}

func parseTupleGenerator(nodes []SchemaNode, metadata ParserMetadata) []Generator {
	if len(nodes) == 0 {
		return nil
	}

	generators := []Generator{}
	for i, item := range nodes {
		refPath := fmt.Sprintf("/prefixItems/%d", i)
		generator, err := metadata.ReferenceHandler.ParseNodeInScope(refPath, item, metadata)
		if err != nil {
			generators = append(generators, NullGenerator{})
		} else {
			generators = append(generators, generator)
		}
	}

	return generators
}

func parseItemGenerator(additionalData AdditionalData, metadata ParserMetadata) Generator {
	if additionalData.DisallowAdditional || additionalData.Schema == nil {
		return nil
	}

	generator, err := metadata.ReferenceHandler.ParseNodeInScope("/items", *additionalData.Schema, metadata)
	if err != nil {
		return nil
	}

	return generator
}

func (g ArrayGenerator) Generate(opts *GeneratorOptions) interface{} {
	tupleLength := len(g.TupleGenerators)
	arrayData := make([]interface{}, 0)
	
	if(tupleLength != 0){
		for _, generator := range g.TupleGenerators {
			arrayData = append(arrayData, generator.Generate(opts))
		}
	}

	if g.ItemGenerator == nil || g.DisallowAdditional {
		return arrayData
	}

	minItems := getInt(g.MinItems, opts.DefaultArrayMinItems)
	maxItems := getInt(g.MaxItems, opts.DefaultArrayMaxItems)	
	remainingItemsToGenerate := maxInt(0, maxItems - tupleLength)

	itemsToGenerate := opts.Rand.RandomInt(0, remainingItemsToGenerate)
	
	// Generate the remaining items up to a random number
	// (This might skew the distribution of the length of the array)
	for i := 0; i < itemsToGenerate || minItems > len(arrayData); i++ {
		arrayData = append(arrayData, g.ItemGenerator.Generate(opts))
	}

	return arrayData
}

func (g ArrayGenerator) String() string {
	tupleString := ""
	for _, generator := range g.TupleGenerators {
		tupleString += fmt.Sprintf("%s,", generator)
	}

	return fmt.Sprintf("ArrayGenerator{items: %s, tuple: [%s] }", g.ItemGenerator, tupleString)
} 
