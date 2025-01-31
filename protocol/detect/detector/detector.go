package detector

type (
	ProtocolDetector interface {
		ProtocolName() string
		IsMatch([]byte) bool
		GetProbe() []byte
	}
)
