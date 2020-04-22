package http

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClient_New_with_CABundleFile(t *testing.T) {
	srv := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Test server")
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
	err = writeCApem(t, err, srv, tmpDir, "ca.pem")

	// New should return new client
	client, err := New(filepath.Join(tmpDir, "ca.pem"), "", time.Second)
	require.NoError(t, err)

	// And http get should work
	resp, err := client.Get(srv.URL)
	require.NoError(t, err)

	_, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
}

func TestClient_New_with_CABundleDir(t *testing.T) {
	srv := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Test server")
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
	err = writeCApem(t, err, srv, tmpDir, "ca.pem")

	// New should return new client
	client, err := New("", tmpDir, time.Second)
	require.NoError(t, err)

	// And http get should work
	resp, err := client.Get(srv.URL)
	require.NoError(t, err)

	_, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
}

func TestClient_New_with_CABundleFile_and_CABundleDir(t *testing.T) {
	srv := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Test server")
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
	err = writeCApem(t, err, srv, tmpDir, "ca")

	// New should return new client
	client, err := New("ca", tmpDir, time.Second)
	require.NoError(t, err)

	// And http get should work
	resp, err := client.Get(srv.URL)
	require.NoError(t, err)

	_, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
}

func TestClient_New_with_CABundleFile_full_path_and_CABundleDir(t *testing.T) {
	srv := httptest.NewTLSServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Test server")
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
	err = writeCApem(t, err, srv, tmpDir, "ca")

	// New should return new client
	client, err := New(filepath.Join(tmpDir, "ca"), tmpDir, time.Second)
	require.NoError(t, err)

	// And http get should work
	resp, err := client.Get(srv.URL)
	require.NoError(t, err)

	_, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
}

func writeCApem(t *testing.T, err error, srv *httptest.Server, tmpDir string, certName string) error {
	caPEM := new(bytes.Buffer)
	err = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: srv.Certificate().Raw,
	})
	require.NoError(t, err)

	// Then write the ca.pem to disk
	caPem, err := os.Create(filepath.Join(tmpDir, certName))
	require.NoError(t, err)
	_, err = caPem.Write(caPEM.Bytes())
	require.NoError(t, err)
	return err
}

func Test_NewAcceptInvalidHostname(t *testing.T) {
	srv := httptest.NewUnstartedServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Test server")
		}))

	serverTLSConf, _, err := certsetup()
	srv.TLS = serverTLSConf
	require.NoError(t, err)

	srv.StartTLS()
	defer srv.Close()
	// Given test server is working
	_, err = srv.Client().Get(srv.URL)
	require.NoError(t, err)

	// Then create temp dir
	tmpDir, err := ioutil.TempDir("", "test")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tmpDir)
		require.NoError(t, err)
	}()

	// Extract ca.pem from TLS server
	err = writeCApem(t, err, srv, tmpDir, "ca.pem")

	// New should return new client
	client, err := NewAcceptInvalidHostname(filepath.Join(tmpDir, "ca.pem"), "", time.Second, "127.0.0.1")
	require.NoError(t, err)

	// And http get should work
	resp, err := client.Get(srv.URL)
	require.NoError(t, err)
	_, err = ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
}

func certsetup() (serverTLSConf *tls.Config, clientTLSConf *tls.Config, err error) {
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
		return nil, nil, err
	}

	// create the CA
	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, nil, err
	}

	// pem encode
	caPEM := new(bytes.Buffer)
	err = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})
	if err != nil {
		return nil, nil, err
	}

	caPrivKeyPEM := new(bytes.Buffer)
	err = pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	})
	if err != nil {
		return nil, nil, err
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
		IPAddresses:  []net.IP{net.IPv4(127, 1, 1, 127), net.IPv6loopback},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, nil, err
	}

	certPEM := new(bytes.Buffer)
	err = pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	if err != nil {
		return nil, nil, err
	}

	certPrivKeyPEM := new(bytes.Buffer)
	err = pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	})
	if err != nil {
		return nil, nil, err
	}

	serverCert, err := tls.X509KeyPair(certPEM.Bytes(), certPrivKeyPEM.Bytes())
	if err != nil {
		return nil, nil, err
	}

	serverTLSConf = &tls.Config{
		Certificates: []tls.Certificate{serverCert},
	}

	certpool := x509.NewCertPool()
	certpool.AppendCertsFromPEM(caPEM.Bytes())
	clientTLSConf = &tls.Config{
		RootCAs: certpool,
	}
	return
}
