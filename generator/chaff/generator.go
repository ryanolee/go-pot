package chaff

import (
	"fmt"
	"strings"

	"github.com/ryanolee/ryan-pot/rand"
	"github.com/thoas/go-funk"
)

type (
	Generator interface {
		fmt.Stringer
		Generate(*GeneratorOptions) interface{}
	}

	GeneratorOptions struct {
		Rand *rand.SeededRand
		
		// The default minimum number value
		DefaultNumberMinimum int

		// The default maximum number value
		DefaultNumberMaximum int

		// The default minimum String length
		DefaultStringMinLength int

		// The default maximum String length
		DefaultStringMaxLength int

		// The default minimum array length
		DefaultArrayMinItems int

		// The default maximum array length
		DefaultArrayMaxItems int

		// The default minimum object properties (Will be ignored if there are fewer properties available)
		DefaultObjectMinProperties int

		// The default maximum object properties (Will be ignored if there are fewer properties available)
		DefaultObjectMaxProperties int

		// The maximum number of references to resolve at once (Default: 10)
		MaximumReferenceDepth int

		// In the event that schemas are recursive there is a good chance the generator
		// will run forever. This option will bypass the check for cyclic references
		// Please defer to the MaximumReferenceDepth option if possible when using this
		BypassCyclicReferenceCheck bool

		resolutions []string
	}
)

func (g *GeneratorOptions) PushRefResolution(reference string) {
	g.resolutions = append(g.resolutions, reference)
}

func (g *GeneratorOptions) PopRefResolution() {
	g.resolutions = g.resolutions[:len(g.resolutions)-1]
}

func (g *GeneratorOptions) HasResolved(reference string) bool {
	return funk.ContainsString(g.resolutions, reference)
}

func (g *GeneratorOptions) GetResolutions() []string {
	return g.resolutions
}

func (g *GeneratorOptions) GetFormattedResolutions() string {
	return strings.Join(g.resolutions, " -> \n")
}

func WithGeneratorOptionsDefaults(options GeneratorOptions) *GeneratorOptions {
	return &GeneratorOptions{
		Rand: options.Rand,
		// Number
		DefaultNumberMinimum: getInt(options.DefaultNumberMinimum, 0),
		DefaultNumberMaximum: getInt(options.DefaultNumberMaximum, 100),

		// String
		DefaultStringMinLength: getInt(options.DefaultStringMinLength, 0),
		DefaultStringMaxLength: getInt(options.DefaultStringMaxLength, 100),

		// Array
		DefaultArrayMinItems: getInt(options.DefaultArrayMinItems, 0),
		DefaultArrayMaxItems: getInt(options.DefaultArrayMaxItems, 10),

		// Object
		DefaultObjectMinProperties: getInt(options.DefaultObjectMinProperties, 0),
		DefaultObjectMaxProperties: getInt(options.DefaultObjectMaxProperties, 10),

		// References
		BypassCyclicReferenceCheck: getBool(options.BypassCyclicReferenceCheck, false),
		MaximumReferenceDepth: getInt(options.MaximumReferenceDepth, 10),
		resolutions: []string{},
	}
}



