package chaff

func getString(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}

	return ""
}

func getInt(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}

	return 0
}

func getFloat(values ...float64) float64 {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}

	return 0
}

func getBool(value bool, defaultValue bool) bool {
	if !value {
		return defaultValue
	}

	return value
}

func maxInt(a ...int) int {
	max := a[0]
	for _, v := range a {
		if v > max {
			max = v
		}
	}

	return max
}

func minInt(a ...int) int {
	min := a[0]
	for _, v := range a {
		if v < min {
			min = v
		}
	}

	return min
}