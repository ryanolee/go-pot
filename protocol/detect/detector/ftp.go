package detector

import "regexp"

type (
	FtpDetector struct {
	}
)

var ftpDetectionRegex = regexp.MustCompile(`^(USER\s{1}[\w-_./]+|AUTH TLS)`)

func NewFtpDetector() *FtpDetector {
	return &FtpDetector{}
}

func (d *FtpDetector) ProtocolName() string {
	return "ftp"
}

func (d *FtpDetector) IsMatch(buffer []byte) bool {
	return ftpDetectionRegex.Match(buffer)
}

func (d *FtpDetector) GetProbe() []byte {
	return []byte("220 FTP Server\r\n")
}
