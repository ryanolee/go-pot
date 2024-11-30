package config

import (
	"errors"
	"regexp"
	"strconv"

	"github.com/go-playground/validator/v10"
)

var portRangeRegex = regexp.MustCompile(`^(\d+)-(\d+)$`)

func ParsePortRange(portRange string) (int, int, error) {
	matches := portRangeRegex.FindStringSubmatch(portRange)
	if len(matches) < 3 {
		return 0, 0, errors.New("invalid port range")
	}

	minPort, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, 0, err
	}

	maxPort, err := strconv.Atoi(matches[2])
	if err != nil {
		return 0, 0, err
	}

	return minPort, maxPort, nil
}

func validatePortRange(portRange validator.FieldLevel) bool {
	// Extract the min and max port values
	minPort, maxPort, err := ParsePortRange(portRange.Field().String())
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

func newConfigValidator() (*validator.Validate, error) {
	v := validator.New()
	if err := v.RegisterValidation("port_range", validatePortRange); err != nil {
		return nil, err
	}
	return v, nil
}
