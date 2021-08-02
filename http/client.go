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

type ClientOption func(*http.Client) error

// New creates a new http.Client with the Passed Transport Options
// that will have a custom certificate if it is loaded from the passed CA Bundle file and/or
// directory. If both CABundleFile and CABundleDir are empty arguments, it creates an unsecure HTTP client.
func New(opts ...ClientOption) (*http.Client, error) {
	t := &http.Transport{
		TLSClientConfig: &tls.Config{},
	}

	client := &http.Client{
		Transport: t,
	}

	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, err
		}
	}

	return client, nil
}

// WithCABundleFile adds the CABundleFile cert to the the client's certPool,
// if a CABundleDir is passed it will be joined to the CABundleFile to try to get the file.
// If the file is a full path, even it the CABundleDir is passed, it will be detected and the full
// path will be used instead
func WithCABundleFile(CABundleFile, CABundleDir string) ClientOption {
	return func(c *http.Client) error {
		if CABundleFile != "" {
			certPool := getClientCertPool(c)
			if err := addCert(filepath.Join(CABundleDir, CABundleFile), certPool); err != nil {
				if os.IsNotExist(err) {
					if err = addCert(CABundleFile, certPool); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}
}

// WithCABundleFile adds the CABundleDir looks for pem certs in the
// CABundleDir and adds them to the the client's certPool.
func WithCABundleDir(CABundleDir string) ClientOption {
	return func(c *http.Client) error {
		if CABundleDir != "" {
			certPool := getClientCertPool(c)
			files, err := ioutil.ReadDir(CABundleDir)
			if err != nil {
				return err
			}

			for _, f := range files {
				if strings.Contains(f.Name(), ".pem") {
					caCertFilePath := filepath.Join(CABundleDir, f.Name())
					if err := addCert(caCertFilePath, certPool); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}
}

// WithAcceptInvalidHostname allows the client to call the acceptInvalidHostname host
// instead of the host from the certificates.
func WithAcceptInvalidHostname(acceptInvalidHostname string) ClientOption {
	return func(c *http.Client) error {
		if acceptInvalidHostname != "" {
			transport := c.Transport.(*http.Transport)
			// Default validation is replaced with VerifyPeerCertificate.
			// Note that when InsecureSkipVerify and VerifyPeerCertificate are in use,
			// ConnectionState.VerifiedChains will be nil.
			transport.TLSClientConfig.InsecureSkipVerify = true
			// While packages like net/http will implicitly set ServerName, the
			// VerifyPeerCertificate callback can't access that value, so it has to be set
			// explicitly here or in VerifyPeerCertificate on the client side. If in
			// an http.Transport DialTLS callback, this can be obtained by passing
			// the addr argument to net.SplitHostPort.
			transport.TLSClientConfig.ServerName = acceptInvalidHostname
			// Approximately equivalent to what crypto/tls does normally:
			// https://github.com/golang/go/commit/29cfb4d3c3a97b6f426d1b899234da905be699aa
			transport.TLSClientConfig.VerifyPeerCertificate = func(certificates [][]byte, _ [][]*x509.Certificate) error {
				certs := make([]*x509.Certificate, len(certificates))
				for i, asn1Data := range certificates {
					cert, err := x509.ParseCertificate(asn1Data)
					if err != nil {
						return errors.New("tls: failed to parse certificate from server: " + err.Error())
					}
					certs[i] = cert
				}

				opts := x509.VerifyOptions{
					Roots:         transport.TLSClientConfig.RootCAs, // On the server side, use config.ClientCAs.
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
		return nil
	}
}

// WithInsecureSkipVerify allows the client to call any host without checking the certificates.
func WithInsecureSkipVerify() ClientOption {
	return func(c *http.Client) error {
		c.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify = true
		return nil
	}
}

// WithTimeout sets the timeout for the client, if not called the timeout will be 0 (no timeout).
func WithTimeout(httpTimeout time.Duration) ClientOption {
	return func(c *http.Client) error {
		c.Timeout = httpTimeout
		return nil
	}
}

func getClientCertPool(c *http.Client) *x509.CertPool {
	transport := c.Transport.(*http.Transport)
	if transport.TLSClientConfig.RootCAs == nil {
		transport.TLSClientConfig.RootCAs = x509.NewCertPool()
	}
	return transport.TLSClientConfig.RootCAs
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
