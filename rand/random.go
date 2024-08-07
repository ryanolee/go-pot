package rand

import (
	"hash/crc64"
	"math/rand"
	"time"

	"github.com/thoas/go-funk"
)

var (
	AlphabetLower = []rune("abcdefghijklmnopqrstuvwxyz")
	AlphabetUpper = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	Hex           = []rune("0123456789abcdef")
	Numbers       = []rune("0123456789")
)

type SeededRand struct {
	Source rand.Source
	Rand   *rand.Rand
}

var crc64Table = crc64.MakeTable(crc64.ISO)

func NewSeededRandFromString(stringSeed string) *SeededRand {
	hash := crc64.Checksum([]byte(stringSeed), crc64Table)
	return NewSeededRand(int64(hash))
}

func NewSeededRandFromTime() *SeededRand {
	return NewSeededRand(time.Now().UnixNano())
}

func NewSeededRand(seed int64) *SeededRand {
	source := rand.NewSource(seed)
	r := rand.New(source)
	return &SeededRand{
		Source: source,
		Rand:   r,
	}
}

// Getters and setters
func (sr *SeededRand) SetSource(source rand.Source) {
	sr.Rand = rand.New(source)
}

// Generic functions
func (sr *SeededRand) Choice(slice []interface{}) interface{} {
	return slice[sr.Rand.Intn(len(slice))]
}

// Array functions
func (sr *SeededRand) StringChoice(stringSlice *[]string) string {
	return (*stringSlice)[sr.Rand.Intn(len(*stringSlice))]
}

func (sr *SeededRand) StringChoiceMultiple(stringSlice *[]string, numChoices int) []string {
	// Pick NumChoices random choices from the string slice without duplicates
	choices := funk.Shuffle(*stringSlice).([]string)

	return choices[:numChoices]

}

// String functions
func (sr *SeededRand) RandomString(length int, runes ...[]rune) string {
	if len(runes) == 0 {
		runes = [][]rune{AlphabetLower, AlphabetUpper, Numbers}
	}

	// Flatten runes
	flatRunes := funk.Flatten(runes).([]rune)

	b := make([]rune, length)
	for i := range b {
		b[i] = flatRunes[sr.Rand.Intn(len(flatRunes))]
	}
	return string(b)
}

// Int functions
func (sr *SeededRand) RandomInt(min int, max int) int {
	// In the the case that min == max, return min
	if min == max {
		return min
	}

	// Random int supporting negative numbers
	return sr.Rand.Intn(max-min) + min
}

// Float functions
func (sr *SeededRand) RandomFloat(min float64, max float64) float64 {
	return sr.Rand.Float64()*(max-min) + min
}

// Bool functions
func (sr *SeededRand) RandomBool() bool {
	return sr.Rand.Intn(2) == 1
}
