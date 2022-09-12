package database

import (
	"crypto/tls"

	"github.com/moov-io/base/log"
)

func LoadTLSClientCertsFromConfig(logger log.Logger, config *MySQLConfig) ([]tls.Certificate, error) {
	var clientCerts []tls.Certificate

	for _, clientCert := range config.TLSClientCerts {
		cert, err := LoadTLSClientCertFromFile(logger, clientCert.CertFilePath, clientCert.KeyFilePath)
		if err != nil {
			return []tls.Certificate{}, err
		}
		clientCerts = append(clientCerts, cert)
	}

	return clientCerts, nil
}

func LoadTLSClientCertFromFile(logger log.Logger, certFile, keyFile string) (tls.Certificate, error) {
	if certFile == "" || keyFile == "" {
		return tls.Certificate{}, logger.LogErrorf("cert path or key path not provided").Err()
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return tls.Certificate{}, logger.LogErrorf("error loading client cert/key from file: %v", err).Err()
	}
	return cert, nil
}
