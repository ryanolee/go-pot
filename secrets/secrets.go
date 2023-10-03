package secrets

import (
	"embed"
	"log"
	"regexp/syntax"

	"github.com/ryanolee/ryan-pot/internal/regen"
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
        Name string `yaml:"name"`
        Weight int `yaml:"weight"`
        NameRegex string `yaml:"name_regex"`
        SecretRegex string `yaml:"secret_regex"`
    }

    SecretGenerator struct {
        Name string
        NameGenerator regen.Generator
        SecretGenerator regen.Generator
    }

)

func NewGenerator(rule SecretGeneratorRule) *SecretGenerator {
    args := &regen.GeneratorArgs{
        Flags: syntax.PerlX,
        MinUnboundedRepeatCount: 10,
    }
    nameGenerator, err := regen.NewGenerator(rule.NameRegex, args)
    if err != nil {
        log.Fatalf("Failed to parse name generator for %s error given as: %s", rule.Name, err)
    }

    secretGenerator, err := regen.NewGenerator(rule.SecretRegex, args)
    if err != nil {
        log.Fatalf("Failed to parse secret generator for %s error given as: %s", rule.Name, err)
    }

    return &SecretGenerator{
        Name: rule.Name,
        NameGenerator: nameGenerator,
        SecretGenerator: secretGenerator,
    }
}

func GetGenerators() []*SecretGenerator{
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

    