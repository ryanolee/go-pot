package chaff

import (
	"fmt"
)

type (
	ReferenceGenerator struct {
		ReferenceStr string
		ReferenceHandler ReferenceHandler
	}
)

func ParseReference(node SchemaNode, metadata ParserMetadata) (ReferenceGenerator, error) {
	return ReferenceGenerator{
		ReferenceStr: node.Ref,
		ReferenceHandler: *metadata.ReferenceHandler,
	}, nil
}

func (g ReferenceGenerator) Generate(opts *GeneratorOptions) interface{} {
	reference, ok := g.ReferenceHandler.Lookup(g.ReferenceStr)
	if !ok {
		return nil
	}

	if len(opts.GetResolutions()) > opts.MaximumReferenceDepth {
		return fmt.Sprintf("Maximum reference depth exceeded: %d \n %s", opts.MaximumReferenceDepth, opts.GetFormattedResolutions())
	}

	if !opts.HasResolved(g.ReferenceStr) && !opts.BypassCyclicReferenceCheck{
		return fmt.Sprintf("Cyclic reference found: %s -> \n %s ", opts.GetFormattedResolutions(), g.ReferenceStr)
	}


	opts.PushRefResolution(g.ReferenceStr)
	defer opts.PopRefResolution()

	return reference.Generator.Generate(opts)
}

func (g ReferenceGenerator) String() string {
	return fmt.Sprintf("ReferenceGenerator{%s}", g.ReferenceStr)
}

