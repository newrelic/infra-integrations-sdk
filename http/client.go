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
			// Default validation is replaced with VerifyPeerCertificate.
			// Note that when InsecureSkipVerify and VerifyPeerCertificate are in use,
			// ConnectionState.VerifiedChains will be nil.
			t.TLSClientConfig.InsecureSkipVerify = true
			// While packages like net/http will implicitly set ServerName, the
			// VerifyPeerCertificate callback can't access that value, so it has to be set
			// explicitly here or in VerifyPeerCertificate on the client side. If in
			// an http.Transport DialTLS callback, this can be obtained by passing
			// the addr argument to net.SplitHostPort.
			t.TLSClientConfig.ServerName = acceptInvalidHostname
			// Approximately equivalent to what crypto/tls does normally:
			// https://github.com/golang/go/commit/29cfb4d3c3a97b6f426d1b899234da905be699aa
			t.TLSClientConfig.VerifyPeerCertificate = func(certificates [][]byte, _ [][]*x509.Certificate) error {
				certs := make([]*x509.Certificate, len(certificates))
				for i, asn1Data := range certificates {
					cert, err := x509.ParseCertificate(asn1Data)
					if err != nil {
						return errors.New("tls: failed to parse certificate from server: " + err.Error())
					}
					certs[i] = cert
				}

				opts := x509.VerifyOptions{
					Roots:         t.TLSClientConfig.RootCAs, // On the server side, use config.ClientCAs.
					DNSName:       acceptInvalidHostname,
					Intermediates: x509.NewCertPool(),
					// On the server side, set KeyUsages to ExtKeyUsageClientAuth. The
					// default value is appropriate for clients side verification.
					// KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
				}
				for _, cert := range certs[1:] {
					opts.Intermediates.AddCert(cert)
				}
				_, err := certs[0].Verify(opts)
				return err
			}

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
