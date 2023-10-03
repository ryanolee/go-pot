package chaff_test

import (
	"fmt"
	"math"
	"testing"

	. "github.com/ryanolee/ryan-pot/generator/chaff"
	"github.com/ryanolee/ryan-pot/rand"
	"github.com/stretchr/testify/assert"
)



func TestNumberGenerate(t *testing.T) {
	
	t.Run("Test int generation", func(t *testing.T){
		testCases := [][] float64 {
			{0, 100},
			{-100, 100},
			{-100, 0},
			{0, 0},
			{0, 1},
			{1, 1},
		}
		
		for _, testCase := range testCases {
			t.Run(fmt.Sprintf("TestCase Min: %f max: %f", testCase[0], testCase[1]), func(t *testing.T){
				generator := &NumberGenerator{
					Min: testCase[0],
					Max: testCase[1],
					Type: TypeInteger,
				}

				result := generator.Generate(&GeneratorOptions{
					Rand: rand.NewSeededRandFromTime(),
				})

				assert.GreaterOrEqual(t, result, int(testCase[0]))
				assert.LessOrEqual(t, result, int(testCase[1]))
				assert.IsType(t, int(0), result)
			})
		}
	})

	t.Run("Test float generation", func(t *testing.T){
		testCases := [][] float64 {
			{-7.2, 8.5},
			{-100, 100},
			{-100, 0},
			{0, 0.001},
		}
		
		for _, testCase := range testCases {
			t.Run(fmt.Sprintf("TestCase Min: %f max: %f", testCase[0], testCase[1]), func(t *testing.T){
				generator := &NumberGenerator{
					Min: testCase[0],
					Max: testCase[1],
					Type: TypeNumber,
				}

				result := generator.Generate(&GeneratorOptions{
					Rand: rand.NewSeededRandFromTime(),
				})

				assert.GreaterOrEqual(t, result, testCase[0])
				assert.LessOrEqual(t, result, testCase[1])
				assert.IsType(t, float64(6), result)
			})
		}
	})

	t.Run("Test Multiple Of Float", func(t *testing.T){
		testCases := [][] float64 {
			{0, 100, 10},
			{-100, 100, 10},
			{0, 100, 10},
			{0, 1, 0.1},
			{0.5, 1, 0.1},
			{-50, -25, 5},
			{612.324, 2342.234, 7},
			{612.324, 2342.234, 4.46},
		}
		
		for _, testCase := range testCases {
			t.Run(fmt.Sprintf("TestCase Min: %f max: %f, multiple: %f", testCase[0], testCase[1], testCase[2]), func(t *testing.T){
				generator := &NumberGenerator{
					Min: testCase[0],
					Max: testCase[1],
					Type: TypeNumber,
					MultipleOf: testCase[2],
				}

				result := generator.Generate(&GeneratorOptions{
					Rand: rand.NewSeededRandFromTime(),
				}).(float64)
				
				assert.GreaterOrEqual(t, result, testCase[0])
				assert.LessOrEqual(t, result, testCase[1])
				assert.IsType(t, float64(0), result)

				// Given the way floating point numbers are handled  sometimes math.Mod
				// will return a number very close to the multiple of the number or 0
				// This assertion handles that ... phun
				res := math.Abs(testCase[2] - math.Mod(result, testCase[2]))
				
				if(res > 0.00001) {
					res = math.Mod(result, testCase[2]);
				}
				
				assert.Less(t, res, 0.00001)	
			})
		}
	})
}

func TestNumberParse(t *testing.T) {
	t.Run("Test Invalid Min/Max", func(t *testing.T){})
}