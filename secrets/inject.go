package secrets

const secretsToInject = 4

func InjectSecrets(generators *SecretGeneratorCollection, data interface{}) interface{} {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return data
	}

	for i := 0; i < secretsToInject; i++ {
		generators.onGenerate()
		generator := generators.GetRandomGenerator()
		dataMap[generator.NameGenerator.Generate()] = generator.SecretGenerator.Generate()
	}

	for key, value := range dataMap {
		dataMap[key] = InjectSecrets(generators, value)
	}

	return dataMap
}
