package http

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/newrelic/infra-integrations-sdk/log"
)

// Client will create new HTTP client based on its configuration
type Client struct {
	CABundleFile string
	CABundleDir  string
	args         interface{}
}

// NewWithCert creates a new http.Client with a custom certificate
func (i Client) NewWithCert() *http.Client {
	return httpClient(i.CABundleFile, i.CABundleDir)
}

func httpClient(certFile string, certDirectory string) *http.Client {
	// go default http transport settings
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	if certFile != "" || certDirectory != "" {
		transport.TLSClientConfig = &tls.Config{RootCAs: getCertPool(certFile, certDirectory)}
	}

	return &http.Client{
		Transport: transport,
	}
}

func getCertPool(certFile string, certDirectory string) *x509.CertPool {
	caCertPool := x509.NewCertPool()
	var writer bytes.Buffer
	log := log.New(false, &writer)
	if certFile != "" {
		caCert, err := ioutil.ReadFile(certFile)
		if err != nil {
			log.Errorf("Error: %s", err)
		}

		ok := caCertPool.AppendCertsFromPEM(caCert)
		if !ok {
			log.Errorf("Cert %q could not be appended", certFile)
		}
	}
	if certDirectory != "" {
		files, err := ioutil.ReadDir(certDirectory)
		if err != nil {
			log.Errorf("Error: %s", err)
		}

		for _, f := range files {
			if strings.Contains(f.Name(), ".pem") {
				caCertFilePath := filepath.Join(certDirectory + "/" + f.Name())
				caCert, err := ioutil.ReadFile(caCertFilePath)
				if err != nil {
					log.Errorf("Error: %s", err)
				}
				ok := caCertPool.AppendCertsFromPEM(caCert)
				if !ok {
					log.Debugf("Cert %q could not be appended", caCertFilePath)
				}
			}
		}
	}
	return caCertPool
}
