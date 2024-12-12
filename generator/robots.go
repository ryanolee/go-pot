package generator

import (
	"embed"
	"fmt"
	"strings"

	"github.com/ryanolee/go-pot/rand"
	"github.com/thoas/go-funk"
)

var (
	//go:embed data/robots.wordlist.txt
	rulesFile embed.FS
)

const robotsTxtFile = "data/robots.wordlist.txt"

type RobotsTxtGenerator struct {
	rand         *rand.SeededRand
	robotEntries *[]string
}

func NewRobotsTxtGenerator(rand *rand.SeededRand) (*RobotsTxtGenerator, error) {
	rulesFile, err := rulesFile.ReadFile(robotsTxtFile)
	if err != nil {
		return nil, err
	}

	robotEntries := strings.Split(string(rulesFile), "\n")

	return &RobotsTxtGenerator{
		rand:         rand,
		robotEntries: &robotEntries,
	}, nil
}

func (g *RobotsTxtGenerator) Generate() string {
	entries := g.rand.StringChoiceMultiple(g.robotEntries, g.rand.RandomInt(1, 10))
	entries = append([]string{"/"}, entries...)
	entries = funk.Map(entries, g.formatEntry).([]string)
	return strings.Join(entries, "")
}

func (g *RobotsTxtGenerator) GenerateChunk() string {
	return g.formatEntry(g.rand.StringChoice(g.robotEntries))
}

func (g *RobotsTxtGenerator) formatEntry(entry string) string {
	return fmt.Sprintf("User-agent: *\nDisallow: %s\n\n", entry)
}
