package encoder

import "regexp"

type MatchingEncoder struct {
	encoder Encoder
	regexp  *regexp.Regexp
}


var defaultEncoder = NewJsonEncoder()
var encoders = []MatchingEncoder{
	{
		encoder: NewYamlEncoder(),
		regexp:  regexp.MustCompile(`\.ya?ml`),
	},
	{
		encoder: NewJsonEncoder(),
		regexp:  regexp.MustCompile(`\.json5?`),
	},
	{
		encoder: NewXmlEncoder(),
		regexp:  regexp.MustCompile(`\.xml`),
	},
	{
		encoder: NewTomlEncoder(),
		regexp:  regexp.MustCompile(`\.toml`),
	},
	{
		encoder: NewHclEncoder(),
		regexp:  regexp.MustCompile(`\.(hcl|tf|tfvars)`),
	},
	{
		encoder: NewIniEncoder(),
		regexp:  regexp.MustCompile(`\.ini`),
	},
}

func GetEncoderForPath(path string) Encoder {
	for _, encoder := range encoders {
		if encoder.regexp.MatchString(path) {
			return encoder.encoder
		}
	}

	return defaultEncoder
}