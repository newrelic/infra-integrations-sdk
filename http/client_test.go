package http

import (
	"bytes"
	"encoding/pem"
	"fmt"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const caCert = `-----BEGIN CERTIFICATE-----
MIIDgjCCAmoCCQDtqmB4gHIHFTANBgkqhkiG9w0BAQsFADCBgjELMAkGA1UEBhMC
RVMxDDAKBgNVBAgMA0NBVDEMMAoGA1UEBwwDYmNuMRIwEAYDVQQKDAlOZXcgcmVs
aWMxDTALBgNVBAsMBG9oYWkxEjAQBgNVBAMMCWxvY2FsaG9zdDEgMB4GCSqGSIb3
DQEJARYRb2hhaUBuZXdyZWxpYy5jb20wHhcNMTgwNTE3MTAxMjUwWhcNMjgwNTE0
MTAxMjUwWjCBgjELMAkGA1UEBhMCRVMxDDAKBgNVBAgMA0NBVDEMMAoGA1UEBwwD
YmNuMRIwEAYDVQQKDAlOZXcgcmVsaWMxDTALBgNVBAsMBG9oYWkxEjAQBgNVBAMM
CWxvY2FsaG9zdDEgMB4GCSqGSIb3DQEJARYRb2hhaUBuZXdyZWxpYy5jb20wggEi
MA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQC8xxoKmMJAjPESMWvEaOn/A5HG
b6ZdwM0MNAQL6b2UpGd1oe8ARcrJkMxD0pttYJFKCLYiTVZISfF/xqJuhQeuaPpH
gU+lDoGNb/HF3Q8YlUfmuZktw45t3biZKRLUDals/EYZBrwPO+8up4/2Hp888gIt
5bxUCVv32eKOwuLjFREwtDDCIZl95ZlzDEyeB0TzvssWFtwj8do3WZ0O3OnmdiKn
C/AqURj6KZmKgWFzELjde+W261N26oCciscgqu565QHo9ZJcAa0IXkTxVgFT+1d5
aUhhFv4oVs64gyAsxGv9EoTdlc2COm5ISqzy6tjVtzsXqaXM0cl7VGTow03ZAgMB
AAEwDQYJKoZIhvcNAQELBQADggEBAIaDnxJwXKe4riMT19LygsVoYExX+tKC6Z/J
37iosZLzu6bzNhvsCSuqDdvCQQkuumlNQgd9XkxtieOMVyrt0MBY7aYdg+dXJXqv
1Ft40590w0Yg6HoAnA2eMvV7D9G1ss6q7VjOae/zxh9UJCsYrVdTU/xYrfyN5HEa
jH7a0BjznBqRSSYub49syKq4EL1oeCF0SMjxuACpriAJ/iAxYibVfO1O2x+AZb6Q
1iFUtU70nOEUrGM0EZ1wZF7atJVgsmdGpsh6kyfsSIZQ5aoNIZHmDVWTfiYcygQd
47Yd5b55SMXDYHGr9ZtRFGKj4IMXqs7R46arQpT4VCPeeSGJhdA=
-----END CERTIFICATE-----`

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
