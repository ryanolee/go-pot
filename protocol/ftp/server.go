package ftp

import (
	ftpserver "github.com/fclairamb/ftpserverlib"
	"github.com/ryanolee/ryan-pot/config"
	"github.com/ryanolee/ryan-pot/protocol/ftp/driver"
)

func NewServer(driver *driver.FtpServerDriver, conf *config.Config) *ftpserver.FtpServer {
	if !conf.FtpServer.Enabled {
		return nil
	}

	return ftpserver.NewFtpServer(driver)
}
