package source

import (
	"github.com/go-faker/faker/v4"
	"github.com/thoas/go-funk"
)

type TabularGenerator struct {
	Name string
	Generate func() string
}

var tabularGeneratorFields = []TabularGenerator{
	{
		Name: "Id",
		Generate: func() string {
			return faker.UUIDHyphenated()
		},
	},
	{
		Name: "FirstName",
		Generate: func() string {
			return faker.FirstName()
		},
	},
	{
		Name: "LastName",
		Generate: func() string {
			return faker.LastName()
		},
	},
	{
		Name: "Email",
		Generate: func() string {
			return faker.Email()
		},
	},
	{
		Name: "Phone",
		Generate: func() string {
			return faker.E164PhoneNumber()
		},
	},
	{
		Name: "Postcode",
		Generate: func() string {
			return faker.GetRealAddress().PostalCode
		},
	},
	{
		Name: "City",
		Generate: func() string {
			return faker.GetRealAddress().City
		},
	},
	{
		Name: "Country",
		Generate: func() string {
			return "USA"
		},
	},
	{
		Name: "CcNumber",
		Generate: func() string {
			return faker.CCNumber()
		},
	},
	{
		Name: "CcType",
		Generate: func() string {
			return faker.CCType()
		},
	},
}

func GetTabularHeaderFields() []string {
	return funk.Map(tabularGeneratorFields, func(field TabularGenerator) string {
		return field.Name
	}).([]string)
}

func GetTabularFieldValues() []string {
	return funk.Map(tabularGeneratorFields, func(field TabularGenerator) string {
		return field.Generate()
	}).([]string)
}