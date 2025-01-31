package config

import (
	"errors"
	"regexp"
	"strconv"

	"github.com/go-playground/validator/v10"
)

var portRangeRegex = regexp.MustCompile(`^(\d+)-(\d+)$`)

type (
	duplicatePortValidatorMeta struct {
		name  string
		ports []int
	}

	// Stateful validator use to assert all ports with the "no_duplicate_port" and "no_duplicate_port_range" tags are unique to stop later bind conflicts
	duplicatePortValidator struct {
		registeredPorts      map[int]*duplicatePortValidatorMeta
		registeredPortRanges []*duplicatePortValidatorMeta
	}
)

func NewDuplicatePortValidator() *duplicatePortValidator {
	return &duplicatePortValidator{
		registeredPorts:      make(map[int]*duplicatePortValidatorMeta),
		registeredPortRanges: make([]*duplicatePortValidatorMeta, 0),
	}
}

func (v *duplicatePortValidator) ValidatePort(fl validator.FieldLevel) bool {
	port := int(fl.Field().Int())
	if _, ok := v.registeredPorts[port]; ok {
		return false
	}

	for _, portRange := range v.registeredPortRanges {
		if port >= portRange.ports[0] && port <= portRange.ports[1] {
			return false
		}
	}

	// Register the port
	v.registeredPorts[port] = &duplicatePortValidatorMeta{
		name:  fl.FieldName(),
		ports: []int{port},
	}

	return true
}

func (v *duplicatePortValidator) ValidatePortRange(fl validator.FieldLevel) bool {
	minPort, maxPort, err := ParsePortRange(fl.Field().String())
	if err != nil {
		return false
	}

	// Assert there are no conflicts with existing port ranges
	for _, port := range v.registeredPortRanges {
		if max(port.ports[0], minPort) <= min(port.ports[1], maxPort) {
			return false
		}
	}

	// Assert there are no conflicts with existing ports
	for port, _ := range v.registeredPorts {
		if port >= minPort && port <= maxPort {
			return false
		}
	}

	// Register the port range
	v.registeredPortRanges = append(v.registeredPortRanges, &duplicatePortValidatorMeta{
		name:  fl.Param(),
		ports: []int{minPort, maxPort},
	})

	return true
}

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

	dpv := NewDuplicatePortValidator()
	if err := v.RegisterValidation("no_duplicate_port", dpv.ValidatePort); err != nil {
		return nil, err
	}

	if err := v.RegisterValidation("no_duplicate_port_range", dpv.ValidatePortRange); err != nil {
		return nil, err
	}

	return v, nil
}
