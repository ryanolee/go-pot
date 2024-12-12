package ftp

import (
	ftpserver "github.com/fclairamb/ftpserverlib"
	"github.com/ryanolee/go-pot/config"
	"github.com/ryanolee/go-pot/protocol/ftp/driver"
)

func NewServer(driver *driver.FtpServerDriver, conf *config.Config) *ftpserver.FtpServer {
	if !conf.FtpServer.Enabled {
		return nil
	}

	return ftpserver.NewFtpServer(driver)
}
