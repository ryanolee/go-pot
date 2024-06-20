package config

import (
	"regexp"
	"strconv"

	"github.com/go-playground/validator/v10"
)

var portRangeRegex = regexp.MustCompile(`^(\d+)-(\d+)$`)

func validatePortRange(portRange validator.FieldLevel) bool {
	// Extract the min and max port values
	matches := portRangeRegex.FindStringSubmatch(portRange.Field().String())

	// Check sting matches
	if len(matches) < 3 {
		return false
	}

	// Convert the port values to integers
	minPort, err := strconv.Atoi(matches[1])
	if err != nil {
		return false
	}

	maxPort, err := strconv.Atoi(matches[2])
	if err != nil {
		return false
	}

	// Validate the port values
	if minPort > maxPort ||
		minPort < 0 || minPort > 65535 ||
		maxPort < 0 || maxPort > 65535 {
		return false
	}

	return true
}

func newConfigValidator() *validator.Validate {
	v := validator.New()
	v.RegisterValidation("port_range", validatePortRange)
	return v
}