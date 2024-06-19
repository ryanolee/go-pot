package driver

import (
	"crypto/tls"
	"fmt"

	ftpserver "github.com/fclairamb/ftpserverlib"
	"github.com/ryanolee/ryan-pot/config"
	"go.uber.org/zap"
)

type FtpServerDriver struct {
	clientFactory *FtpClientDriverFactory
	settings      *ftpserver.Settings
	tlsConfig     *tls.Config
}

func NewFtpServerDriver(c *config.Config, cf *FtpClientDriverFactory) (*FtpServerDriver, error) {
	cert, err := getSelfSignedCert(c)
	if err != nil {
		return nil, err
	}

	return &FtpServerDriver{
		clientFactory: cf,
		tlsConfig: &tls.Config{
			Certificates: []tls.Certificate{
				cert,
			},

			InsecureSkipVerify: true,
			ClientAuth:         tls.NoClientCert,
		},
		settings: &ftpserver.Settings{
			// Feature toggles we want to keep enabled
			DisableMLSD: false,
			DisableMLST: false,
			DisableMFMT: false,

			// Connection port range
			ListenAddr: fmt.Sprintf("%s:%d", c.FtpServer.Host, c.FtpServer.Port),
			PassiveTransferPortRange: &ftpserver.PortRange{
				Start: 50000,
				End:   50100,
			},

			// Disable active mode
			DisableActiveMode: true,

			// Aim to be as open as possible
			TLSRequired:            ftpserver.ClearOrEncrypted,
			ActiveConnectionsCheck: ftpserver.IPMatchDisabled,
			PasvConnectionsCheck:   ftpserver.IPMatchDisabled,
		},
	}, nil
}

func (f *FtpServerDriver) GetSettings() (*ftpserver.Settings, error) {
	zap.L().Sugar().Info("__STUB__  FtpServerDriver.GetSettings")
	return f.settings, nil
}

func (f *FtpServerDriver) ClientConnected(cc ftpserver.ClientContext) (string, error) {
	zap.L().Sugar().Info("__STUB__  FtpServerDriver.ClientConnected", cc)
	return "Welcome to the FTP Server", nil
}

func (f *FtpServerDriver) ClientDisconnected(cc ftpserver.ClientContext) {
	zap.L().Sugar().Info("__STUB__  FtpServerDriver.ClientDisconnected", cc)
	return
}

func (f *FtpServerDriver) AuthUser(cc ftpserver.ClientContext, user, pass string) (ftpserver.ClientDriver, error) {
	zap.L().Sugar().Info("__STUB__  FtpServerDriver.AuthUser", cc, user, pass)
	return f.clientFactory.FromContext(cc), nil
}

func (f *FtpServerDriver) GetTLSConfig() (*tls.Config, error) {
	zap.L().Sugar().Info("__STUB__  FtpServerDriver.GetTLSConfig")
	return f.tlsConfig, nil
}
