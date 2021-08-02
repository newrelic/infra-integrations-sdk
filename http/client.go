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

var (
	// ErrInvalidTransportType defines the error to be returned if the transport defined for the client is incorrect
	ErrInvalidTransportType = errors.New("roundTripper transport invalid, should be http type")
	// ErrEmptyCABundleFile defines the error to be returned if the WithCABundleFile option is called with an empty value
	ErrEmptyCABundleFile = errors.New("caBundleFile can't be empty")
	// ErrEmptyCABundleDir defines the error to be returned if the WithCABundleDir option is called with an empty value
	ErrEmptyCABundleDir = errors.New("caBundleDir can't be empty")
	// ErrEmptyAcceptInvalidHostname defines the error to be returned if the WithAcceptInvalidHostname option is called with an empty value
	ErrEmptyAcceptInvalidHostname = errors.New("acceptInvalidHostname can't be empty")
)

// ClientOption defines the format of the client option functions
// Note that many options defined in this package rely on the Transport for the http.Client being of type *http.Transport.
// Using a different Transport will cause these options to fail.
type ClientOption func(*http.Client) error

// New creates a new http.Client with the Passed Client Options
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

// WithCABundleFile adds the CABundleFile cert to the the client's certPool
func WithCABundleFile(CABundleFile string) ClientOption {
	return func(c *http.Client) error {
		if CABundleFile == "" {
			return ErrEmptyCABundleFile
		}

		certPool, err := clientCertPool(c)
		if err != nil {
			return err
		}
		return addCert(CABundleFile, certPool)
	}
}

// WithCABundleDir adds the CABundleDir looks for pem certs in the
// CABundleDir and adds them to the the client's certPool.
func WithCABundleDir(CABundleDir string) ClientOption {
	return func(c *http.Client) error {
		if CABundleDir == "" {
			return ErrEmptyCABundleDir
		}

		certPool, err := clientCertPool(c)
		if err != nil {
			return err
		}

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
		return nil
	}
}

// WithAcceptInvalidHostname allows the client to call the acceptInvalidHostname host
// instead of the host from the certificates.
func WithAcceptInvalidHostname(acceptInvalidHostname string) ClientOption {
	return func(c *http.Client) error {
		if acceptInvalidHostname == "" {
			return ErrEmptyAcceptInvalidHostname
		}

		transport, ok := c.Transport.(*http.Transport)
		if !ok {
			return ErrInvalidTransportType
		}
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
		return nil
	}
}

// WithInsecureSkipVerify allows the client to call any host without checking the certificates.
func WithInsecureSkipVerify() ClientOption {
	return func(c *http.Client) error {
		transport, ok := c.Transport.(*http.Transport)
		if !ok {
			return ErrInvalidTransportType
		}

		transport.TLSClientConfig.InsecureSkipVerify = true
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

func clientCertPool(c *http.Client) (*x509.CertPool, error) {
	transport, ok := c.Transport.(*http.Transport)
	if !ok {
		return nil, ErrInvalidTransportType
	}

	if transport.TLSClientConfig.RootCAs == nil {
		transport.TLSClientConfig.RootCAs = x509.NewCertPool()
	}
	return transport.TLSClientConfig.RootCAs, nil
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
