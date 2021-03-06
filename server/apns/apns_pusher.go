package apns

import (
	"crypto/tls"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/certificate"
)

type Pusher interface {
	Push(*apns2.Notification) (*apns2.Response, error)
}

func newPusher(c Config) (Pusher, error) {
	var (
		cert    tls.Certificate
		errCert error
	)
	if c.CertificateFileName != nil && *c.CertificateFileName != "" {
		cert, errCert = certificate.FromP12File(*c.CertificateFileName, *c.CertificatePassword)
	} else {
		cert, errCert = certificate.FromP12Bytes(*c.CertificateBytes, *c.CertificatePassword)
	}
	if errCert != nil {
		return nil, errCert
	}
	if *c.Production {
		logger.Debug("APNS Pusher in Production mode")
		return apns2.NewClient(cert).Production(), nil
	}
	logger.Debug("APNS Pusher in Development mode")
	return apns2.NewClient(cert).Development(), nil
}
