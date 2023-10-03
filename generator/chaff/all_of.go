package chaff

type AllOfGenerator struct {
	UnparsedNodes []SchemaNode
	MergedNodes []SchemaNode
	Processed bool
}

func ParseAllOf(node SchemaNode, metadata ParserMetadata) (AllOfGenerator, error) {
	generator := AllOfGenerator{
		UnparsedNodes: node.AllOf,
	}
	return generator, nil
}



func (g *AllOfGenerator) Generate(opts *GeneratorOptions) interface{} {
	return nil
}

func mergeSchemaNode(metadata ParserMetadata, nodes ...SchemaNode) (SchemaNode, error) {
	if len(nodes) == 0 {
		return SchemaNode{}, nil
	}

	mergedNode := SchemaNode{
		Type: MultipleType{},
		Enum: make([]interface{}, 0),
	}

	for _, node := range nodes {
		if node == nil {
			continue
		}

		if node.Ref != "" {
			node, ok = metadata.ReferenceHandler.Lookup(node.Ref).SchemaNode
			if !ok {
				return SchemaNode{}, fmt.Errorf("Reference not found: %s", node.Ref)
			}
		}

		// Merge Type
		if node.Type.SingleType != "" {
			mergedNode.Type.MultipleTypes = append(node.Type.MultipleTypes, node.Type.SingleType)
		} else if len(node.Type.MultipleTypes) > 0 {
			mergedNode.Type.MultipleTypes = append(mergedNode.Type.MultipleTypes, node.Type.MultipleTypes...)
		}

		// Merge simple int properties
		mergedNode.Length = getInt(node.Length, mergedNode.Length)
		mergedNode.MinProperties = getInt(node.MinProperties, mergedNode.MinProperties)
		mergedNode.MaxProperties = getInt(node.MaxProperties, mergedNode.MaxProperties)
		mergedNode.MinItems = getInt(node.MinItems, mergedNode.MinItems)
		mergedNode.MaxItems = getInt(node.MaxItems, mergedNode.MaxItems)
		mergedNode.MinContains = getInt(node.MinContains, mergedNode.MinContains)
		mergedNode.MaxContains = getInt(node.MaxContains, mergedNode.MaxContains)
		
		// Merge simple float properties
		mergedNode.Minimum = getFloat(node.Minimum, mergedNode.Minimum)
		mergedNode.Maximum = getFloat(node.Maximum, mergedNode.Maximum)
		mergedNode.ExclusiveMinimum = getFloat(node.ExclusiveMinimum, mergedNode.ExclusiveMinimum)
		mergedNode.ExclusiveMaximum = getFloat(node.ExclusiveMaximum, mergedNode.ExclusiveMaximum)
		mergedNode.MultipleOf = getFloat(node.MultipleOf, mergedNode.MultipleOf)

		// Merge simple string properties
		mergedNode.Pattern = getString(node.Pattern, mergedNode.Pattern)
		mergedNode.Format = getString(node.Format, mergedNode.Format)

		
		if len(node.Enum) > 0 {
			mergedNode.Enum = append(mergedNode.Enum, node.Enum...)
		}

		// Merge properties
		for key, value := range node.Properties {
			mergedNode.Properties[key] = mergeSchemaNode(metadata, mergedNode.Properties[key], value)
		}

		for key, value := range node.PatternProperties {
			mergedNode.PatternProperties[key] = mergeSchemaNode(metadata, mergedNode.PatternProperties[key], value)
		}

		// Merge array items - @todo: Is this how the schema spec works?
		//                            for merging prefixItems?
		for i := 0; i < len(node.PrefixItems); i++ {
			mergedNode.PrefixItems[i] = mergeSchemaNode(metadata, mergedNode.PrefixItems[i], node.PrefixItems[i])
		}

		mergedNode.OneOf = append(mergedNode.OneOf, node.OneOf...)
		mergedNode.AnyOf = append(mergedNode.AnyOf, node.AnyOf...)

		mergedNode.AllOf = append(mergedNode.AllOf, node.AllOf...)
	}
	

	return SchemaNode{}, nil
}

func mergeSchemaNodeMap(metadata ParserMetadata, currentNode SchemaNode, nodeMap map[string]SchemaNode) (map[string]SchemaNode, error) {
	for key, node := range nodeMap {
}
