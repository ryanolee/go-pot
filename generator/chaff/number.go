package chaff

import (
	"errors"
	"math"

	"github.com/ryanolee/ryan-pot/rand"
)

type (
	NumberGeneratorType string
	NumberGenerator struct {
		Type NumberGeneratorType
		Min float64
		Max float64
		MultipleOf float64
	}
)

const (
	Infinitesimal = math.SmallestNonzeroFloat64
	TypeInteger NumberGeneratorType = "integer"
	TypeNumber NumberGeneratorType = "number"

	DefaultOffset = 10
)

func ParseNumber(node SchemaNode, genType NumberGeneratorType) (Generator, error) {
	var min float64
	var max float64
	
	// Initial Validation
	if node.Minimum != 0 && node.ExclusiveMinimum != 0 {
		return NullGenerator{}, errors.New("cannot have both minimum and exclusive minimum")
	}

	if node.Maximum != 0 && node.ExclusiveMaximum != 0 {
		return NullGenerator{}, errors.New("cannot have both maximum and exclusive maximum")
	}

	// Set min and max
	if node.Minimum != 0 {
		min = float64(node.Minimum)
	} else if node.ExclusiveMinimum != 0 {
		min = float64(node.ExclusiveMinimum) + Infinitesimal
	}

	if node.Maximum != 0 {
		max = float64(node.Maximum)
	} else if node.ExclusiveMaximum != 0 {
		max = float64(node.ExclusiveMaximum) - Infinitesimal
	} else if min != 0 {
		max = min + DefaultOffset
	} else {
		max = DefaultOffset
	}

	// Validate min and max
	if min > max {
		return NullGenerator{}, errors.New("minimum cannot be greater than maximum")
	}

	// Validate multipleOf
	if node.MultipleOf != 0 {
		if node.MultipleOf <= 0 {
			return NullGenerator{}, errors.New("multipleOf cannot be negative or zero")
		}

		multiplesInRange := countMultiplesInRange(min, max, node.MultipleOf)

		if multiplesInRange == 0 {
			return NullGenerator{}, errors.New("minimum and maximum do not allow for any multiples of multipleOf")
		}
	}
	
	return &NumberGenerator{
		Type: genType,
		Min: min,
		Max: max,
		MultipleOf: node.MultipleOf,
	}, nil
}

func countMultiplesInRange(min float64, max float64, multiple float64) int {
	if min == 0 {
		return int(math.Floor(max / multiple))
	}

	return int(math.Floor(max / multiple)) - int(math.Floor(min / multiple))
}

func generateMultipleOf(rand rand.SeededRand, min float64, max float64, multiple float64) float64{
	multiplesInRange := countMultiplesInRange(min, max, multiple)

	if multiplesInRange == 0 {
		return 0
	}

	lowerBound := math.Floor(min / multiple) * multiple
	randomMultiple := float64(rand.RandomInt(1, multiplesInRange)) * multiple
	return  lowerBound + randomMultiple


}
func (g *NumberGenerator) Generate(opts *GeneratorOptions) interface{} {
	if g.Type == TypeInteger && g.MultipleOf != 0 {
		return int(generateMultipleOf(*opts.Rand, g.Min, g.Max, g.MultipleOf))
	} else if g.Type == TypeInteger && g.MultipleOf == 0 {
		return int(math.Round(opts.Rand.RandomFloat(g.Min, g.Max)))
	} else if g.Type == TypeNumber && g.MultipleOf != 0 {
		return generateMultipleOf(*opts.Rand, g.Min, g.Max, g.MultipleOf)
	} else if g.Type == TypeNumber && g.MultipleOf == 0 {
		return opts.Rand.RandomFloat(g.Min, g.Max)
	}

	return 0
}

func (g *NumberGenerator) String() string {
	return "NumberGenerator"
}
