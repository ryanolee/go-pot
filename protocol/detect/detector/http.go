package detector

import "regexp"

type (
	HttpDetector struct {
	}
)

var httpDetectionRegex = regexp.MustCompile(`(?i)^(GET|POST|HEAD|PUT|DELETE|OPTIONS|TRACE|CONNECT)`)

func NewHttpDetector() *HttpDetector {
	return &HttpDetector{}
}

func (d *HttpDetector) ProtocolName() string {
	return "http"
}

func (d *HttpDetector) IsMatch(buffer []byte) bool {
	return httpDetectionRegex.Match(buffer)
}

func (d *HttpDetector) GetProbe() []byte {
	return nil
}
