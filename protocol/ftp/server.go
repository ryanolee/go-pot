package ftp

import (
	ftpserver "github.com/fclairamb/ftpserverlib"
	"github.com/ryanolee/ryan-pot/protocol/ftp/driver"
)

func NewServer(driver *driver.FtpServerDriver) *ftpserver.FtpServer {
	return ftpserver.NewFtpServer(driver)
}
