// Package http provides an easy way to construct an http client with custom certificates and customizable timeout.
// If you need to customize other attributes you can use the golang http package. https://golang.org/pkg/net/http/
package http

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// ClientArguments are meant to be used as flags from a custom integrations. With this you could
// send this arguments from the command line.
type ClientArguments struct {
	HTTPCaBundleFile string        `default: "" help: "Name of the certificate file"`
	HTTPCaBundleDir  string        `default: "" help: "Path where the certificate exists"`
	HTTPTimeout      time.Duration `default:30 help: "Client http timeout in seconds"`
}

// New creates a new http.Client with a custom certificate
func New(CABundleFile, CABundleDir string, httpTimeout time.Duration) (*http.Client, error) {
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
		Timeout:   httpTimeout * time.Second,
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

		ok := caCertPool.AppendCertsFromPEM(caCert)
		if !ok {
			return nil, errors.New("can't parse certificate")
		}

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
				ok := caCertPool.AppendCertsFromPEM(caCert)
				if !ok {
					return nil, errors.New("can't parse certificate")
				}
			}
		}
	}
	return caCertPool, nil
}