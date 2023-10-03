package chaff

import (
	"fmt"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/ryanolee/ryan-pot/internal/regen"
)

type (
	StringGenerator struct {
		Format StringFormat
		Pattern string
		PatternGenerator regen.Generator
	}
)

type StringFormat string

const (
	// Time
	FormatDateTime StringFormat = "date-time" // RFC3339
	FormatTime StringFormat = "time" //
	FormatDate StringFormat = "date"
	FormatDuration StringFormat = "duration"
	
	// Email
	FormatEmail StringFormat = "email"
	FormatIdnEmail StringFormat = "idn-email"

	// Hostname
	FormatHostname StringFormat = "hostname"
	FormatIdnHostname StringFormat = "idn-hostname"

	// IP
	FormatIpv4 StringFormat = "ipv4"
	FormatIpv6 StringFormat = "ipv6"

	// Rescource Identifier
	FormatUUID StringFormat = "uuid"
	FormatURI StringFormat = "uri"
	FormatURIReference StringFormat = "uri-reference"
	FormatIRI StringFormat = "iri"
	FormatIRIReference StringFormat = "iri-reference"

	// Uri Template
	FormatUriTemplate StringFormat = "uri-template"

	// JSON Pointer
	FormatJSONPointer StringFormat = "json-pointer"
	FormatRelativeJSONPointer StringFormat = "relative-json-pointer"

	// Regex
	FormatRegex StringFormat = "regex"
)

func ParseString(node SchemaNode, metadata ParserMetadata) (Generator, error) {
	if node.Format != "" && node.Pattern != "" {
		return NullGenerator{}, fmt.Errorf("cannot have both format and pattern on a string")
	}

	generator := StringGenerator{
		Format: StringFormat(node.Format),
		Pattern: node.Pattern,
	}

	if node.Pattern != "" {
		regenGenerator, err := regen.NewGenerator(node.Pattern, &metadata.ParserOptions.RegexStringOptions)
		if err != nil {
			return NullGenerator{}, fmt.Errorf("invalid regex pattern: %s", node.Pattern)
		}

		generator.PatternGenerator = regenGenerator
	}

	return generator, nil
}



func (g StringGenerator) Generate(opts *GeneratorOptions) interface{} {
	if(g.Pattern != "") {
		return g.PatternGenerator.Generate()
	}
	
	if(g.Format != "") {
		return GenerateFormat(g.Format, opts)
	}

	return faker.Sentence()
}

func GenerateFormat(format StringFormat, opts *GeneratorOptions) string {
	switch format {
	case FormatDateTime: 
		return time.Unix(faker.UnixTime(), 0).Format(time.RFC3339)
	case FormatTime:
		return time.Unix(faker.UnixTime(), 0).Format(time.TimeOnly)
	case FormatDate:
		return time.Unix(faker.UnixTime(), 0).Format(time.DateOnly)
	case FormatDuration:
		return fmt.Sprintf("P%dD", opts.Rand.RandomInt(0, 90))
	case FormatEmail, FormatIdnEmail:
		return faker.Email()
	case FormatHostname, FormatIdnHostname:
		return faker.DomainName()
	case FormatIpv4:
		return faker.IPv4()
	case FormatIpv6:
		return faker.IPv6()
	case FormatUUID:
		return faker.UUIDHyphenated()
	case FormatURI, FormatURIReference, FormatIRI, FormatIRIReference:
		return faker.URL()
	default:
		return fmt.Sprintf("Unsupported Format: %s", format)
	}
}

func (g StringGenerator) String() string {
	return "StringGenerator"
}