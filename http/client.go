package http

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
)

// NewWithCert creates a new http.Client with a custom certificate
func NewWithCert(CABundleFile, CABundleDir string) (*http.Client, error) {
	// go default http transport settings
	transport := &http.Transport{}

	if CABundleFile != "" || CABundleDir != "" {
		certs, err := getCertPool(CABundleFile, CABundleDir)
		if err != nil {
			return nil, err
		}
		transport.TLSClientConfig = &tls.Config{RootCAs: certs}
	}

	return &http.Client{
		Transport: transport,
	}, nil
}

func getCertPool(certFile string, certDirectory string) (*x509.CertPool, error) {
	caCertPool := x509.NewCertPool()
	if certFile != "" {
		caCert, err := ioutil.ReadFile(certFile)
		if err != nil {
			return nil, err
		}

		_ = caCertPool.AppendCertsFromPEM(caCert)

	}
	if certDirectory != "" {
		files, err := ioutil.ReadDir(certDirectory)
		if err != nil {
			return nil, err
		}

		for _, f := range files {
			if strings.Contains(f.Name(), ".pem") {
				caCertFilePath := filepath.Join(certDirectory + "/" + f.Name())
				caCert, err := ioutil.ReadFile(caCertFilePath)
				if err != nil {
					return nil, err
				}
				_ = caCertPool.AppendCertsFromPEM(caCert)
			}
		}
	}
	return caCertPool, nil
}
