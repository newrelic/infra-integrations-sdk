// Package http provides an easy way to construct an http client with custom certificates and customizable timeout.
// If you need to customize other attributes you can use the golang http package. https://golang.org/pkg/net/http/
package http

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// New creates a new http.Client with a custom certificate, which can be loaded from the passed CA Bundle file and/or
// directory. If both CABundleFile and CABundleDir are empty arguments, it creates an unsecure HTTP client.
func New(CABundleFile, CABundleDir string, httpTimeout time.Duration) (*http.Client, error) {
	return _new(CABundleFile, CABundleDir, httpTimeout, "")
}

// NewAcceptInvalidHostname new http.Client with ability to accept HTTPS certificates that don't
// match the hostname of the server they are connecting to.
func NewAcceptInvalidHostname(CABundleFile, CABundleDir string, httpTimeout time.Duration, hostname string) (*http.Client, error) {
	return _new(CABundleFile, CABundleDir, httpTimeout, hostname)
}

func _new(CABundleFile, CABundleDir string, httpTimeout time.Duration, acceptInvalidHostname string) (*http.Client, error) {
	// go default http transport settings
	t := &http.Transport{}

	if CABundleFile != "" || CABundleDir != "" {
		certs, err := getCertPool(CABundleFile, CABundleDir)
		if err != nil {
			return nil, err
		}

		t.TLSClientConfig = &tls.Config{
			RootCAs: certs,
		}

		if acceptInvalidHostname != "" {
			// Setting ServerName tells Go's TLS verification to check the
			// certificate against this hostname instead of the connection address.
			// This allows connecting to a host whose certificate is issued to a
			// different name while still performing full certificate chain validation.
			t.TLSClientConfig.ServerName = acceptInvalidHostname
		}
	}

	return &http.Client{
		Timeout:   httpTimeout,
		Transport: t,
	}, nil
}

func getCertPool(certFile string, certDirectory string) (*x509.CertPool, error) {
	caCertPool := x509.NewCertPool()

	if certFile != "" {
		if err := addCert(filepath.Join(certDirectory, certFile), caCertPool); err != nil {
			if os.IsNotExist(err) {
				if err = addCert(certFile, caCertPool); err != nil {
					return nil, err
				}
			}
		}
	}

	if certDirectory != "" {
		files, err := ioutil.ReadDir(certDirectory)
		if err != nil {
			return nil, err
		}

		for _, f := range files {
			if strings.Contains(f.Name(), ".pem") {
				caCertFilePath := filepath.Join(certDirectory, f.Name())
				if err := addCert(caCertFilePath, caCertPool); err != nil {
					return nil, err
				}
			}
		}
	}
	return caCertPool, nil
}

func addCert(certFile string, caCertPool *x509.CertPool) error {
	caCert, err := ioutil.ReadFile(certFile)
	if err != nil {
		return err
	}

	ok := caCertPool.AppendCertsFromPEM(caCert)
	if !ok {
		return errors.New("can't parse certificate")
	}
	return nil
}
