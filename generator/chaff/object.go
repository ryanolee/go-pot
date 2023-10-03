package chaff

import (
	"fmt"

	"github.com/ryanolee/ryan-pot/internal/regen"
	"github.com/thoas/go-funk"
)

type (
	ObjectGenerator struct {
		Properties map[string]Generator
		
		// Pattern Properties Regex -> Generator mapping
		PatternProperties map[string]Generator
		PatternPropertiesRegex map[string] regen.Generator

		DisallowAdditionalProperties bool
		AdditionalProperties Generator

		FallbackGenerator Generator

		MinProperties int
		MaxProperties int
		Required []string
	}
)

func ParseObject(node SchemaNode, metadata ParserMetadata) (Generator, error) {
	// Validator Max and Min Properties
	if node.MinProperties < 0 {
		return NullGenerator{}, fmt.Errorf("minProperties must be greater than or equal to 0")
	}

	if node.MaxProperties < 0 {
		return NullGenerator{}, fmt.Errorf("maxProperties must be greater than or equal to 0")
	}

	if node.MinProperties > node.MaxProperties {
		return NullGenerator{}, fmt.Errorf("minProperties must be less than or equal to MaxProperties")
	}

	// Validate Required Properties
	if node.MaxProperties != 0 && len(node.Required) > node.MaxProperties {
		return NullGenerator{}, fmt.Errorf("required properties must have a length of less than or equal to MaxProperties")
	}

	for _, requiredProperty := range node.Required {
		if _, ok := node.Properties[requiredProperty]; !ok {
			return NullGenerator{}, fmt.Errorf("required property %s does not exist in properties", requiredProperty)
		}
	}

	// Validate additionalProperties
	if node.AdditionalProperties.DisallowAdditional && node.MinProperties > len(node.Properties) {
		return NullGenerator{}, fmt.Errorf("minProperties must be less than or equal to the number of properties")
	}

	patternProperties, patternPropertiesRegex := parsePatternProperties(node, metadata)

	objectGenerator := ObjectGenerator{
		Required: node.Required,
		MinProperties: node.MinProperties,
		MaxProperties: node.MaxProperties,
		
		Properties: parseProperties(node, metadata),
		PatternProperties: patternProperties,
		PatternPropertiesRegex: patternPropertiesRegex,

		DisallowAdditionalProperties: node.AdditionalProperties.DisallowAdditional,
		AdditionalProperties: parseAdditionalProperties(node, metadata),
		FallbackGenerator: NullGenerator{},
	}

	return objectGenerator, nil
}

func parseProperties(node SchemaNode, metadata ParserMetadata) map[string]Generator {
	properties := make(map[string]Generator)
	ref := metadata.ReferenceHandler
	for	name, prop := range node.Properties {
		refPath := fmt.Sprintf("/properties/%s", name)
		propGenerator, err := ref.ParseNodeInScope(refPath, prop, metadata)
		if err != nil {
			propGenerator = NullGenerator{}
		}

		properties[name] = propGenerator
	}

	return properties
}

func parseAdditionalProperties(node SchemaNode, metadata ParserMetadata) Generator {
	if node.AdditionalProperties.DisallowAdditional || node.AdditionalProperties.Schema == nil {
		return nil
	}
	ref := metadata.ReferenceHandler
	refPath := "/additionalProperties"
	additionalProperties, err := ref.ParseNodeInScope(refPath, *node.AdditionalProperties.Schema, metadata)
	
	if err != nil {
		return NullGenerator{}
	}

	return additionalProperties
}

func parsePatternProperties(node SchemaNode, metadata ParserMetadata) (map[string]Generator, map[string] regen.Generator) {
	if node.PatternProperties == nil {
		return nil, nil
	}

	propertiesRegex := make(map[string]regen.Generator)
	properties := make(map[string]Generator)
	ref := metadata.ReferenceHandler

	for regex, property := range node.PatternProperties {
		refPath := fmt.Sprintf("/patternProperties/%s", regex)

		// Parse the schema node
		propGenerator, err := ref.ParseNodeInScope(refPath, property, metadata)
		if err != nil {
			propGenerator = NullGenerator{}
		}

		regexGenerator, err := regen.NewGenerator(regex, &metadata.ParserOptions.RegexPatternPropertyOptions)
		if err != nil {
			errPath := fmt.Sprintf("%s/regex/%s", ref.CurrentPath, regex)
			metadata.Errors[errPath] = fmt.Errorf("failed to create regex generator for %s. Error given: %s", regex, err) 
			regexGenerator = nil
		}

		propertiesRegex[regex] = regexGenerator
		properties[regex] = propGenerator
	}

	return properties, propertiesRegex
}

func (g ObjectGenerator) Generate(opts *GeneratorOptions) interface{} {
	// Generate Required Properties
	generatedValues := make(map[string]interface{})
	for _, key := range g.Required {
		generatedValues[key] = g.Properties[key].Generate(opts)
	}

	// Generate A random distribution of optional properties, pattern properties, and additional properties
	// (Using a fallback generator if none are available)
	optionalKeys := funk.UniqString(append(g.Required, funk.Keys(g.Properties).([]string)...))
	
	min := getInt(g.MinProperties, opts.DefaultObjectMinProperties)
	max := getInt(g.MaxProperties, opts.DefaultObjectMaxProperties)
	minimumExtrasToGenerate := maxInt(0, min - len(g.Required))
	maximumExtrasToGenerate := maxInt(0, max - len(g.Required))

	generatorTarget := opts.Rand.RandomInt(minimumExtrasToGenerate, maximumExtrasToGenerate)

	numberOfOptionalKeysToGenerate := minInt(len(optionalKeys), generatorTarget)
	optionalKeysToGenerate := opts.Rand.StringChoiceMultiple(&optionalKeys, numberOfOptionalKeysToGenerate)

	// Generate any optional keys
	for _, key := range optionalKeysToGenerate {
		generatedValues[key] = g.Properties[key].Generate(opts)
	}

	generatorTarget -= len(optionalKeysToGenerate)

	// Generate any pattern properties
	// Failing that generate any additional properties 
	// Failing that generate any fallback properties
	if len(g.PatternProperties) > 0 {
		for i := 0; i < generatorTarget; i++ {
			regex, value := g.GeneratePatternProperty(opts)
			generatedValues[regex] = value
		}
	}  else if g.DisallowAdditionalProperties {
		return generatedValues
	} else if  g.AdditionalProperties != nil {
		for i := 0; i < generatorTarget; i++ {
			generatedValues[fmt.Sprintf("additional_%d", i)] = g.AdditionalProperties.Generate(opts)
		}
	} else {
		for i := 0; i < generatorTarget; i++ {
			generatedValues[fmt.Sprintf("fallback_%d", i)] = g.FallbackGenerator.Generate(opts)
		}
	}

	return generatedValues
}

func (g ObjectGenerator) GeneratePatternProperty(opts *GeneratorOptions) (string, interface{}) {
	if len(g.PatternProperties) == 0 {
		return "", nil
	}

	availableRegexes := funk.Keys(g.PatternProperties).([]string)
	targetRegex := opts.Rand.StringChoice(&availableRegexes)
	targetRegexGenerator := g.PatternPropertiesRegex[targetRegex]
	targetGenerator := g.PatternProperties[targetRegex]

	return targetRegexGenerator.Generate(), targetGenerator.Generate(opts)
}

func (g ObjectGenerator) String() string {
	formattedString := ""
	for name, prop := range g.Properties {
		formattedString += fmt.Sprintf("%s: %s,", name, prop)
	}

	regexString := ""
	for regex, prop := range g.PatternProperties {
		regexString += fmt.Sprintf("%s: %s,", regex, prop)
	}

	return fmt.Sprintf("ObjectGenerator{properties: %s, patternProperties: %s, additionalProperties: %s}", formattedString, regexString, g.AdditionalProperties)
}

