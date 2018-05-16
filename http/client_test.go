package http

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_NewWithCert(t *testing.T) {
	file, err := ioutil.TempFile("/tmp/", "ca.pem")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	c := Client{
		"ca.pem",
		"/tmp/",
	}
	c.NewWithCert()
	assert.Equal(t, c.CABundleFile, "ca.pem")
	assert.Equal(t, c.CABundleDir, "/tmp/")
}
