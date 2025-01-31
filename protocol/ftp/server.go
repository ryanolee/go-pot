package ftp

import (
	ftpserver "github.com/fclairamb/ftpserverlib"
	"github.com/ryanolee/go-pot/config"
	"github.com/ryanolee/go-pot/protocol/detect"
	"github.com/ryanolee/go-pot/protocol/ftp/driver"
)

func NewServer(driver *driver.FtpServerDriver, conf *config.Config, ls *detect.MultiProtocolListener) *ftpserver.FtpServer {
	if !conf.FtpServer.Enabled && !ls.ProtocolEnabled("ftp") {
		return nil
	}

	server := ftpserver.NewFtpServer(driver)
	return server
}
