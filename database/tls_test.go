package database_test

import (
	"path/filepath"
	"testing"

	"github.com/madflojo/testcerts"
	"github.com/moov-io/base/database"
	"github.com/moov-io/base/log"
	"github.com/stretchr/testify/require"
)

func Test_LoadClientCertsFromConfig(t *testing.T) {
	certFilepath := filepath.Join("/", "tmp", "client_cert.pem")
	keyFilepath := filepath.Join("/", "tmp", "client_cert_private_key.pem")

	err := testcerts.GenerateCertsToFile(certFilepath, keyFilepath)
	require.Nil(t, err)

	config := &database.MySQLConfig{
		TLSClientCerts: []database.TLSClientCertConfig{
			{
				CertFilePath: certFilepath,
				KeyFilePath:  keyFilepath,
			},
		},
	}

	clientCerts, err := database.LoadTLSClientCertsFromConfig(log.NewNopLogger(), config)
	require.Nil(t, err)

	require.Len(t, clientCerts, 1)
	require.Len(t, clientCerts[0].Certificate, 1)
}
