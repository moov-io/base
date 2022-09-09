package database_test

import (
	"path/filepath"
	"testing"

	"github.com/moov-io/base/database"
	"github.com/moov-io/base/log"
	"github.com/stretchr/testify/require"
)

func Test_LoadClientCertsFromConfig(t *testing.T) {

	config := &database.MySQLConfig{
		TLSClientCerts: []database.TLSClientCertConfig{
			{
				CertFilePath: filepath.Join("testdata", "client_cert.pem"),
				KeyFilePath:  filepath.Join("testdata", "client_cert_private_key.pem"),
			},
		},
	}

	clientCerts, err := database.LoadTLSClientCertsFromConfig(log.NewNopLogger(), config)
	require.Nil(t, err)

	require.Len(t, clientCerts, 1)
	require.Len(t, clientCerts[0].Certificate, 1)
}
