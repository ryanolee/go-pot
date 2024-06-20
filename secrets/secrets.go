package secrets

import (
	"embed"
	"log"
	"regexp/syntax"
	"strings"

	"github.com/ryanolee/ryan-pot/core/metrics"
	"github.com/ryanolee/ryan-pot/internal/regen"
	"github.com/ryanolee/ryan-pot/rand"
	"gopkg.in/yaml.v3"

	"github.com/thoas/go-funk"
)

var (
	//go:embed secret-rules.yml
	rulesFile embed.FS
)

type (
	SecretGeneratorRules = map[string]SecretGeneratorRule

	SecretGeneratorRule struct {
		Name        string `yaml:"name"`
		Weight      int    `yaml:"weight"`
		NameRegex   string `yaml:"name_regex"`
		SecretRegex string `yaml:"secret_regex"`
	}

	SecretGenerator struct {
		Name            string
		NameGenerator   regen.Generator
		SecretGenerator regen.Generator
	}

	SecretGeneratorCollectionInput struct {
		OnGenerate func()
	}
	SecretGeneratorCollection struct {
		onGenerate func()
		Generators []*SecretGenerator
	}
)

func NewSecretGeneratorCollection(telemetry *metrics.Telemetry) *SecretGeneratorCollection {
	return &SecretGeneratorCollection{
		Generators: GetGenerators(),
		onGenerate: func(){
			if telemetry == nil {
				return
			}

			telemetry.TrackGeneratedSecrets(1)
		},
	}
}

func (c *SecretGeneratorCollection) GetRandomGenerator() *SecretGenerator {
	rnd := rand.NewSeededRandFromTime()
	return c.Generators[rnd.RandomInt(0, len(c.Generators))]
}

func NewGenerator(rule SecretGeneratorRule) *SecretGenerator {
	args := &regen.GeneratorArgs{
		Flags:                   syntax.PerlX,
		MinUnboundedRepeatCount: 30,
	}
	nameGenerator, err := newRegexGenerator(rule.NameRegex, args)
	if err != nil {
		log.Fatalf("Failed to parse name generator for %s error given as: %s", rule.Name, err)
	}

	secretGenerator, err := newRegexGenerator(rule.SecretRegex, args)
	if err != nil {
		log.Fatalf("Failed to parse secret generator for %s error given as: %s", rule.Name, err)
	}

	return &SecretGenerator{
		Name:            rule.Name,
		NameGenerator:   nameGenerator,
		SecretGenerator: secretGenerator,
	}
}

const fullStringLiteral = "[~{FULL_STOP_LITERAL}~]"

func newRegexGenerator(pattern string, opts *regen.GeneratorArgs) (regen.Generator, error) {
	pattern = strings.ReplaceAll(pattern, "\\.", fullStringLiteral)
	pattern = strings.ReplaceAll(pattern, ".", "\\w")
	pattern = strings.ReplaceAll(pattern, fullStringLiteral, "\\.")

	return regen.NewGenerator(pattern, opts)
}

func GetGenerators() []*SecretGenerator {
	rules := GetRules()
	return funk.Map(funk.Values(rules), NewGenerator).([]*SecretGenerator)
}

func GetRules() *SecretGeneratorRules {
	rules := &SecretGeneratorRules{}
	yamlFile, err := rulesFile.ReadFile("secret-rules.yml")
	if err != nil {
		panic(err)
	}

	if err := yaml.Unmarshal(yamlFile, &rules); err != nil {
		panic(err)
	}

	return rules
}
