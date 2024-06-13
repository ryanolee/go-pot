package driver

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)
func getSelfSignedCert() (tls.Certificate, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println("Failed to generate key")
		return tls.Certificate{}, err
	}
	

	certTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"ryan-pot"},
		},
		NotBefore: time.Now(),
		NotAfter: time.Now().AddDate(1, 0, 0),
		KeyUsage: x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA: true,
	}
		
	x509Cert, err := x509.CreateCertificate(rand.Reader, certTemplate, certTemplate, key.Public(), key);

	pemCert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: x509Cert})
	pemKey := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	return tls.X509KeyPair(pemCert, pemKey)
}