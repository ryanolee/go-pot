package driver

import (
	"crypto/tls"
	"fmt"
	"net"

	ftpserver "github.com/fclairamb/ftpserverlib"
	"github.com/ryanolee/go-pot/config"
	"github.com/ryanolee/go-pot/protocol/detect"
	"github.com/ryanolee/go-pot/protocol/ftp/logging"
	"github.com/ryanolee/go-pot/protocol/ftp/throttle"
	"go.uber.org/zap"
)

// FTP Server handles top level FTP Sever Connections
// [FTP Sever Driver] --> FTP Client Driver --> FTP File, FTP File Info

type FtpServerDriver struct {
	clientFactory *FtpClientDriverFactory
	settings      *ftpserver.Settings
	tlsConfig     *tls.Config
	throttle      *throttle.FtpThrottle
	logger        *logging.FtpCommandLogger
}

func NewFtpServerDriver(c *config.Config, cf *FtpClientDriverFactory, throttle *throttle.FtpThrottle, logger *logging.FtpCommandLogger, ls *detect.MultiProtocolListener) (*FtpServerDriver, error) {
	multiProtocolEnabled := ls.ProtocolEnabled("ftp")
	if !c.FtpServer.Enabled && !multiProtocolEnabled {
		return nil, nil
	}

	cert, err := getSelfSignedCert(c)
	if err != nil {
		return nil, err
	}

	lowerRange, upperRange, err := config.ParsePortRange(c.FtpServer.PassivePortRange)
	if err != nil {
		return nil, err
	}

	var listener net.Listener = nil
	listenerAddr := fmt.Sprintf("%s:%d", c.FtpServer.Host, c.FtpServer.Port)
	if multiProtocolEnabled {
		listener = ls.GetListenerForProtocol("ftp")
		listenerAddr = ""
	}

	return &FtpServerDriver{
		clientFactory: cf,
		throttle:      throttle,
		logger:        logger,
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

			// Listener set
			ListenAddr: listenerAddr,
			Listener:   listener,

			// Connection port range
			PassiveTransferPortRange: &ftpserver.PortRange{
				Start: lowerRange,
				End:   upperRange,
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
	return f.settings, nil
}

func (f *FtpServerDriver) ClientConnected(cc ftpserver.ClientContext) (string, error) {
	f.logger.LogWithContext(cc, "client_connected")
	return "Welcome to the FTP Server", nil
}

func (f *FtpServerDriver) ClientDisconnected(cc ftpserver.ClientContext) {
	f.logger.LogWithContext(cc, "client_disconnected")
	clientId := f.clientFactory.GetClientIdFromContent(cc)
	f.throttle.Unregister(clientId)
}

func (f *FtpServerDriver) AuthUser(cc ftpserver.ClientContext, user, pass string) (ftpserver.ClientDriver, error) {
	f.logger.LogWithContext(cc, "auth_user",
		zap.String("user", user),
		zap.String("pass", pass),
		zap.String("client_version", cc.GetClientVersion()),
		zap.String("client_ip", cc.RemoteAddr().String()),
	)

	return f.clientFactory.FromContext(cc), nil
}

func (f *FtpServerDriver) GetTLSConfig() (*tls.Config, error) {
	return f.tlsConfig, nil
}
