package http_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	httpSDK "github.com/newrelic/infra-integrations-sdk/v4/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_New_with_CABundleFile(t *testing.T) {
	srv := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := fmt.Fprintln(w, "Test server")
			assert.NoError(t, err)
		}))
	defer srv.Close()

	// Given test server is working
	_, err := srv.Client().Get(srv.URL)
	require.NoError(t, err)

	// Then create temp dir
	tmpDir, err := ioutil.TempDir("", "test")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tmpDir)
		require.NoError(t, err)
	}()

	writeCApem(t, srv, tmpDir, "ca.pem")

	// New should return new client
	client, err := httpSDK.New(httpSDK.WithTimeout(time.Second), httpSDK.WithCABundleFile(tmpDir+"/ca.pem"))
	require.NoError(t, err)

	// And http get should work
	resp, err := client.Get(srv.URL)
	require.NoError(t, err)

	// Number of certs loaded should be 1
	require.Equal(t, 1, len(client.Transport.(*http.Transport).TLSClientConfig.RootCAs.Subjects()))

	_, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
}

func TestClient_New_with_CABundleDir(t *testing.T) {
	srv := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := fmt.Fprintln(w, "Test server")
			assert.NoError(t, err)
		}))
	defer srv.Close()

	// Given test server is working
	_, err := srv.Client().Get(srv.URL)
	require.NoError(t, err)

	// Then create temp dir
	tmpDir, err := ioutil.TempDir("", "test")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tmpDir)
		require.NoError(t, err)
	}()

	writeCApem(t, srv, tmpDir, "ca.pem")

	// New should return new client
	client, err := httpSDK.New(httpSDK.WithTimeout(time.Second), httpSDK.WithCABundleDir(tmpDir))
	require.NoError(t, err)

	// And http get should work
	resp, err := client.Get(srv.URL)
	require.NoError(t, err)

	// Number of certs loaded should be 1
	require.Equal(t, 1, len(client.Transport.(*http.Transport).TLSClientConfig.RootCAs.Subjects()))

	_, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
}

func TestClient_New_with_CABundleFile_and_CABundleDir(t *testing.T) {
	srv := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := fmt.Fprintln(w, "Test server")
			assert.NoError(t, err)
		}))
	defer srv.Close()

	// Given test server is working
	_, err := srv.Client().Get(srv.URL)
	require.NoError(t, err)

	// Then create temp dir
	tmpDir, err := ioutil.TempDir("", "test")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tmpDir)
		require.NoError(t, err)
	}()

	writeCApem(t, srv, tmpDir, "ca")
	writeAnotherCApem(t, tmpDir, "ca2.pem")

	// New should return new client
	client, err := httpSDK.New(httpSDK.WithTimeout(time.Second), httpSDK.WithCABundleFile(tmpDir+"/ca"), httpSDK.WithCABundleDir(tmpDir))
	require.NoError(t, err)

	// And http get should work
	resp, err := client.Get(srv.URL)
	require.NoError(t, err)

	// Number of certs loaded should be 2
	require.Equal(t, 2, len(client.Transport.(*http.Transport).TLSClientConfig.RootCAs.Subjects()))

	_, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
}

// Extract ca.pem from TLS server
func writeCApem(t *testing.T, srv *httptest.Server, tmpDir string, certName string) {
	caPEM := new(bytes.Buffer)
	err := pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: srv.Certificate().Raw,
	})
	require.NoError(t, err)

	// Then write the ca.pem to disk
	caPem, err := os.Create(filepath.Join(tmpDir, certName))
	require.NoError(t, err)
	_, err = caPem.Write(caPEM.Bytes())
	require.NoError(t, err)
}

func writeAnotherCApem(t *testing.T, tmpDir string, certName string) {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			Organization:  []string{"Another, INC."},
			Country:       []string{"MO"},
			Province:      []string{""},
			Locality:      []string{"Maputo"},
			StreetAddress: []string{"Maputo City"},
			PostalCode:    []string{"14016"},
			CommonName:    "invented",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// create our private and public key
	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	require.NoError(t, err)

	// create the CA
	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	require.NoError(t, err)

	// pem encode
	caPEM := new(bytes.Buffer)
	err = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	// Then write the ca.pem to disk
	caPem, err := os.Create(filepath.Join(tmpDir, certName))
	require.NoError(t, err)
	_, err = caPem.Write(caPEM.Bytes())
	require.NoError(t, err)
}

func Test_NewAcceptInvalidHostname(t *testing.T) {
	srv := httptest.NewUnstartedServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := fmt.Fprintln(w, "Test server")
			assert.NoError(t, err)
		}))
	defer srv.Close()

	// Given a server certificate accepting a certain IP and hostname
	var ip net.IP
	ip = net.IPv4(127, 0, 0, 111)
	serverTLSConf, err := certsetup("foo.bar", []net.IP{ip})
	require.NoError(t, err)
	srv.TLS = serverTLSConf

	// And server is running HTTPS
	srv.StartTLS()

	// And folder in client to contain a certificate
	tmpDir, err := ioutil.TempDir("", "test")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tmpDir)
		require.NoError(t, err)
	}()

	// And ca.pem exists for client
	writeCApem(t, srv, tmpDir, "ca.pem")

	// 2 assertions:

	// When HTTPS client is created accepting server certificated IP
	sameIP := ip.String()
	client, err := httpSDK.New(httpSDK.WithTimeout(time.Second), httpSDK.WithAcceptInvalidHostname(sameIP), httpSDK.WithCABundleFile(tmpDir+"/ca.pem"))
	require.NoError(t, err)

	// Then HTTPS should work even for different hostname and source IP (127.0.0.1)
	req, err := http.NewRequest("GET", srv.URL, nil)
	require.NoError(t, err)
	req.Host = "different.hostname"
	resp, err := client.Do(req)
	require.NoError(t, err)

	_, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	// When HTTPS client is created accepting server certificated hostname
	client, err = httpSDK.New(httpSDK.WithTimeout(time.Second), httpSDK.WithAcceptInvalidHostname("foo.bar"), httpSDK.WithCABundleFile(tmpDir+"/ca.pem"))
	require.NoError(t, err)

	// Then HTTPS should work
	req, err = http.NewRequest("GET", srv.URL, nil)
	require.NoError(t, err)
	req.Host = "different.hostname"
	resp, err = client.Do(req)
	require.NoError(t, err)

	_, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
}

func Test_WithInsecureSkipVerify(t *testing.T) {
	srv := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := fmt.Fprintln(w, "Test server")
			assert.NoError(t, err)
		}))
	defer srv.Close()

	// Given test server is working
	_, err := srv.Client().Get(srv.URL)
	require.NoError(t, err)

	// Then create temp dir
	tmpDir, err := ioutil.TempDir("", "test")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tmpDir)
		require.NoError(t, err)
	}()

	// Extract ca.pem from TLS server
	writeCApem(t, srv, tmpDir, "ca.pem")

	// New should return new client
	client, err := httpSDK.New(httpSDK.WithTimeout(time.Second), httpSDK.WithInsecureSkipVerify())
	require.NoError(t, err)

	// And http get should work
	resp, err := client.Get(srv.URL)
	require.NoError(t, err)

	_, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
}

func certsetup(hostname string, ips []net.IP) (serverTLSConf *tls.Config, err error) {
	// set up our CA certificate
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			Organization:  []string{"Company, INC."},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{"Golden Gate Bridge"},
			PostalCode:    []string{"94016"},
			CommonName:    hostname,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// create our private and public key
	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return
	}

	// create the CA
	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return
	}

	// pem encode
	caPEM := new(bytes.Buffer)
	err = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})
	if err != nil {
		return
	}

	caPrivKeyPEM := new(bytes.Buffer)
	err = pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	})
	if err != nil {
		return
	}

	// set up our server certificate
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			Organization:  []string{"Company, INC."},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{"Golden Gate Bridge"},
			PostalCode:    []string{"94016"},
		},
		DNSNames:     []string{hostname},
		IPAddresses:  ips,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return
	}

	certPEM := new(bytes.Buffer)
	err = pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	if err != nil {
		return
	}

	certPrivKeyPEM := new(bytes.Buffer)
	err = pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	})
	if err != nil {
		return
	}

	serverCert, err := tls.X509KeyPair(certPEM.Bytes(), certPrivKeyPEM.Bytes())
	if err != nil {
		return
	}

	serverTLSConf = &tls.Config{
		Certificates: []tls.Certificate{serverCert},
	}

	return
}
