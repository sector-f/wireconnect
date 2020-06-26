package server

import (
	"crypto/tls"
	"sync"
)

type ReloadableCert struct {
	certFile string
	keyFile  string
	tlsCert  *tls.Certificate
	mu       sync.Mutex
}

func NewReloadableCert(certFile string, keyFile string) (*ReloadableCert, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	rCert := ReloadableCert{
		certFile: certFile,
		keyFile:  keyFile,
		tlsCert:  &cert,
	}

	return &rCert, nil
}

func (cert *ReloadableCert) Reload() error {
	cert.mu.Lock()
	defer cert.mu.Unlock()

	newCert, err := tls.LoadX509KeyPair(cert.certFile, cert.keyFile)
	if err != nil {
		return err
	}

	cert.tlsCert = &newCert
	return nil
}

func (cert *ReloadableCert) GetCertificate(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	cert.mu.Lock()
	defer cert.mu.Unlock()

	return cert.tlsCert, nil
}
