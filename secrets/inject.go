package secrets
const secretsToInject = 5

func InjectSecrets(generators *SecretGeneratorCollection, data interface{}) interface{} {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return data
	}

	for i := 0; i < secretsToInject; i++ {
		generator := generators.GetRandomGenerator()
		dataMap[generator.NameGenerator.Generate()] = generator.SecretGenerator.Generate() 
	}
	
	for key, value := range dataMap {
		dataMap[key] = InjectSecrets(generators, value)
	}
	
	return dataMap
}