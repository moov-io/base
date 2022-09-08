package database

import (
	"crypto/tls"
	"os"
	"strings"

	"github.com/moov-io/base/log"
)

const SQL_CLIENT_TLS_CERT = "SQL_CLIENT_TLS_CERT"
const SQL_CLIENT_TLS_PRIVATE_KEY = "SQL_CLIENT_TLS_PRIVATE_KEY"

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

func LoadTLSClientCertFromEnv(logger log.Logger) (tls.Certificate, error) {
	cert, certOk := os.LookupEnv(SQL_CLIENT_TLS_CERT)
	key, keyOk := os.LookupEnv(SQL_CLIENT_TLS_PRIVATE_KEY)

	if certOk && keyOk && strings.TrimSpace(cert) != "" && strings.TrimSpace(key) != "" {
		logger.Info().Log("loading client cert from environment")

		certPemBlock := []byte(cert)
		keyPemBlock := []byte(key)

		clientCert, err := tls.X509KeyPair(certPemBlock, keyPemBlock)
		if err != nil {
			return tls.Certificate{}, logger.LogErrorf("error loading client cert from environment: %v", err).Err()
		}

		return clientCert, nil
	}

	return tls.Certificate{}, logger.LogErrorf("missing client cert env vars").Err()
}
