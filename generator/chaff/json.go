package chaff

import (
	"encoding/json"
)

type (
	// Additional properties can be a schema node or a boolean value.
	// This handles both cases.
	AdditionalData struct {
		Schema *SchemaNode
		DisallowAdditional bool
	}

	MultipleType struct {
		SingleType string
		MultipleTypes []string
	}
)

func (a *AdditionalData) UnmarshalJSON(data []byte) error {
	if(len(data) == 0) {
		return nil
	}

	if string(data) == "false" {
		a.DisallowAdditional = true
		return nil
	}
	
	var schema SchemaNode
	err := json.Unmarshal(data, &schema)
	if err != nil {
		return err
	}

	a.Schema = &schema
	return nil
}

func (m *MultipleType) UnmarshalJSON(data []byte) error {
	if(len(data) == 0) {
		return nil
	}

	var multipleTypes []string
	var singleType string

	// Try to parse an array of types
	multipleTypesError := json.Unmarshal(data, &multipleTypes)
	singleTypeError := json.Unmarshal(data, &singleType)

	if multipleTypesError != nil && singleTypeError != nil {
		return singleTypeError
	}

	

	m.MultipleTypes = multipleTypes
	m.SingleType = singleType

	return nil
}