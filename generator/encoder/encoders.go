package encoder

import "regexp"

type MatchingEncoder struct {
	encoder       Encoder
	generatorType  string
	regexp  *regexp.Regexp
}


var defaultEncoder = NewJsonEncoder()
var encoders = []MatchingEncoder{
	{
		encoder: NewYamlEncoder(),
		generatorType: "structured",
		regexp:  regexp.MustCompile(`\.ya?ml`),
	},
	{
		encoder: NewJsonEncoder(),
		generatorType: "structured",
		regexp:  regexp.MustCompile(`\.json5?`),
	},
	{
		encoder: NewXmlEncoder(),
		generatorType: "structured",
		regexp:  regexp.MustCompile(`\.xml`),
	},
	{
		encoder: NewTomlEncoder(),
		generatorType: "structured",
		regexp:  regexp.MustCompile(`\.toml`),
	},
	{
		encoder: NewHclEncoder(),
		generatorType: "structured",
		regexp:  regexp.MustCompile(`\.(hcl|tf|tfvars)`),
	},
	{
		encoder: NewIniEncoder(),
		generatorType: "structured",
		regexp:  regexp.MustCompile(`\.ini`),
	},
	{
		encoder: NewCsvEncoder(),
		generatorType: "tabular",
		regexp:  regexp.MustCompile(`\.csv`),
	},
	{
		encoder: NewSqlEncoder(),
		generatorType: "tabular",
		regexp:  regexp.MustCompile(`\.sql`),
	},
}

func GetEncoderForPath(path string) Encoder {
	for _, encoder := range encoders {
		if encoder.regexp.MatchString(path) {
			return  encoder.encoder
		}
	}

	return defaultEncoder
}