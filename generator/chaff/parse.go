package chaff

import (
	"encoding/json"
	"os"

	"github.com/ryanolee/ryan-pot/internal/regen"
)

type (
	// Struct containing metadata for parse operations within the JSON Schema
	ParserMetadata struct {
		// Used to keep track of every referenceable route
		ReferenceHandler *ReferenceHandler
		ParserOptions ParserOptions
		Errors map[string]error
		
		// Generators that need to have their structures Re-Parsed once all references have been resolved
		UnprocessedGenerators []Generator
	}

	ParserOptions struct {
		RegexStringOptions regen.GeneratorArgs
		RegexPatternPropertyOptions regen.GeneratorArgs
	}

	SchemaNode struct {
		// Shared Properties
		Type MultipleType `json:"type"`
		Length int `json:"length"` // Shared by String and Array

		// Object Properties
		Properties map[string]SchemaNode `json:"properties"`
		AdditionalProperties AdditionalData `json:"additionalProperties"`
		PatternProperties map[string]SchemaNode `json:"patternProperties"`
		MinProperties int `json:"minProperties"`
		MaxProperties int `json:"maxProperties"`
		Required []string `json:"required"`

		// String Properties
		Pattern string `json:"pattern"`
		Format string `json:"format"`

		// Number Properties
		Minimum float64 `json:"minimum"`
		Maximum float64 `json:"maximum"`
		ExclusiveMinimum float64 `json:"exclusiveMinimum"`
		ExclusiveMaximum float64 `json:"exclusiveMaximum"`
		MultipleOf float64 `json:"multipleOf"`

		// Array Properties
		Items AdditionalData `json:"items"`
		MinItems int `json:"minItems"`
		MaxItems int `json:"maxItems"`

		Contains *SchemaNode `json:"contains"`
		MinContains int `json:"minContains"`
		MaxContains int `json:"maxContains"`

		PrefixItems []SchemaNode `json:"prefixItems"`

		// Enum Properties
		Enum []interface{} `json:"enum"`

		// Constant Properties
		Const interface{} `json:"const"`

		// Combination Properties
		// TODO: Implement these
		//Not *SchemaNode `json:"not"`
		AllOf []SchemaNode `json:"allOf"`
		AnyOf []SchemaNode `json:"anyOf"`
		OneOf []SchemaNode `json:"oneOf"`
		
		// Reference Operator
		Ref string `json:"$ref"`
		Id string `json:"$id"`
		Defs map[string]SchemaNode `json:"$defs"`
		Definitions map[string]SchemaNode `json:"definitions"`
	}
)

type Op int64



const (
	// Null Operations
	Noop Op = iota

	// Data Type Operations
	Object = "object"
	Array = "array"
	Number = "number"
	Integer = "integer"
	String = "string"
	Boolean = "boolean"
)

func ParseSchemaFile(path string) (RootGenerator, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return RootGenerator{
			Generator: NullGenerator{},
		}, err
	}

	return ParseSchema(data)
}

func ParseSchema(schema []byte) (RootGenerator, error) {
	var node SchemaNode
	err := json.Unmarshal(schema, &node)
	if err != nil {
		return RootGenerator{
			Generator: NullGenerator{},
		}, err
	}
	
	refHandler := NewReferenceHandler()
	generator, err := ParseRoot(node, ParserMetadata{
		ReferenceHandler: &refHandler,
		Errors: make(map[string]error),
	})

	return generator, err
}

func ParseNode(node SchemaNode, metadata ParserMetadata)(Generator, error){
	refHandler := metadata.ReferenceHandler
	gen, err := ParseSchemaNode(node, metadata)

	if err != nil {
		metadata.Errors[refHandler.CurrentPath] = err
	}

	if node.Id != "" {
		refHandler.AddIdReference(node.Id, node, gen)
	}

	refHandler.AddReference(node, gen)
	return gen, err
		
}

func ParseSchemaNode(node SchemaNode, metadata ParserMetadata)(Generator, error){
	// Handle reference nodes
	if node.Ref != "" {
		return ParseReference(node, metadata)
	}

	// Handle combination nodes
	if node.OneOf != nil || node.AnyOf != nil {
		return ParseCombination(node, metadata)
	}

	// Handle enum nodes
	if node.Enum != nil {
		return ParseEnum(node)
	}

	// Handle constant nodes
	if node.Const != nil {
		return ParseConst(node)
	}

	// Handle multiple type nodes
	if node.Type.MultipleTypes != nil {
		return ParseMultipleType(node, metadata)
	}

	return ParseType(node.Type.SingleType, node, metadata)
}

func ParseType(nodeType string, node SchemaNode, metadata ParserMetadata) (Generator, error) {
	// Handle object nodes
	switch nodeType {
		case Object: return ParseObject(node, metadata)
		case Array: return ParseArray(node, metadata)
		case Number: return ParseNumber(node, TypeNumber)
		case Integer: return ParseNumber(node, TypeInteger)
		case String: return ParseString(node, metadata)
		case Boolean: return ParseBoolean(node)
		default: return NullGenerator{}, nil
	}
}
