package auth

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
)

func newHttpClient(cacertfile string) (*http.Client, error) {

	// if there is no ca certificate, use default RootCAs
	if len(cacertfile) == 0 {
		tr := &http.Transport{
			ForceAttemptHTTP2: true,
		}
		return &http.Client{Transport: tr}, nil
	}

	// load the ca cert file
	cacert, err := os.ReadFile(cacertfile)
	if err != nil {
		return nil, fmt.Errorf("auth/client: failed to read ca certificate file: %w", err)
	}
	certs := x509.NewCertPool()
	certs.AppendCertsFromPEM(cacert)

	tlsConfig := tls.Config{
		RootCAs: certs,
	}

	tr := &http.Transport{
		TLSClientConfig:   &tlsConfig,
		ForceAttemptHTTP2: true,
	}

	return &http.Client{Transport: tr}, nil
}
