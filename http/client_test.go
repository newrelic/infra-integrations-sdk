package http

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"testing"

	"github.com/stretchr/testify/assert"
)

var caCert = `-----BEGIN CERTIFICATE-----
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

var privateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAvMcaCpjCQIzxEjFrxGjp/wORxm+mXcDNDDQEC+m9lKRndaHv
AEXKyZDMQ9KbbWCRSgi2Ik1WSEnxf8aiboUHrmj6R4FPpQ6BjW/xxd0PGJVH5rmZ
LcOObd24mSkS1A2pbPxGGQa8DzvvLqeP9h6fPPICLeW8VAlb99nijsLi4xURMLQw
wiGZfeWZcwxMngdE877LFhbcI/HaN1mdDtzp5nYipwvwKlEY+imZioFhcxC43Xvl
tutTduqAnIrHIKrueuUB6PWSXAGtCF5E8VYBU/tXeWlIYRb+KFbOuIMgLMRr/RKE
3ZXNgjpuSEqs8urY1bc7F6mlzNHJe1Rk6MNN2QIDAQABAoIBAQCd4nepPTHaAwbs
jGDxmD18h2O4b1DZQJM+DZME0603UHknLRRTSgvcoTn1z4Mm64kYPkj2T3BGbXGJ
yHu5q5FNEYehnkkaZxN7U5EGR2iEyvWjxr6SQ+gvgy0NDAkvSW3WNPf7nmJS63GT
t5jz45CSzGV+NZJZRqqglJ6jf+N6v4grEZEcIjVuOL5NppVCJE7mpOqAQlJG0kqu
sb41VoHEfy2DHuHX3fby3LaCl+EpHbZG+zbEksEO08mbcPKj8qfg81QyQFzuUIkH
+E2tcVKhoR834fQ+n2MzI/4yR4pkNFghdjCwK3nn+3UUt7agKzo/VmEF6sCt6y29
TnshnhIBAoGBAN7eLRpZ91znRNp6fmdbT0vQIpAtoGtvqyp6YZmR0ho8mxvBLROJ
QIhI5Mc/cwPeOCwX2tzRjUmQjfRfl275NYi3tWQVJpSTGZdR8jYn6zt1ZA5H/VJj
RJMrp4qksapCXPT0nv3Q7SRIhwT3V2TbJ0ssmDwahsSdtcDDguvjgR4xAoGBANjX
iSRHmvm5LfYIzbMm6ZZ78JfFX7rPfkIunBM5F6GZJkAJGmDZCXi4M5JPePwv8xAC
lOIc8HX8m53IvIC6MzBtU7k5/yhPsM7GRuQ4j6trrcs3IgkU87gwKEz1nrLIbWxI
S6/4hNG2P2QVnMlt3FoRSgmsbk46W1SVYt99tvgpAoGBAINxDbDI9rb4PweLzxku
JSpVas0V29MBXTYET6O++Oc4b1KDMA6hmEnIlAVfSnoxiXeX6iDqBiYo90/1QN7W
Y9hqYLTSNJrT1vgEAJIoIPhEV+qEUsdQfJU/3eRLFe2Qjjp6O3r+yZ3omJk5N3Xo
Oth/SJnKG0nCqfsyU/jDiNdBAoGAT+2auom+YUBV5bO3BstYHMUQmRECyVxEYObH
VvqbcFCAXeg9FefKavoS4GJ06RhPkt4wvOwH4qW7QrzEZvq7daVG0CbFm7lMJdvG
M8d5halKRXbMD+buMz1lDYEX/zSLyPcZFwMXCioQUbb5tPHO4FAxJ0Gs4x71nUb3
TAQN1okCgYA+Oi0KYss+6kVfmin27Loo08UrGQwAgRPcHPeKHVBt1yCQD2QKBEkw
r8q6iHv2b4jEsEWM2+V/ratUvN9ji/WKxhbQAARIW58n11kAJsE48hwpPBGPTnFX
x4W9xFO4kHze4qDxeIBh2OlyUqA9eUptrkkzie5CSlYE2A7JqkB43g==
-----END RSA PRIVATE KEY-----`

var quit = make(chan os.Signal, 1)
var done = make(chan bool)

func startHTTPServer(file, keyFile string) {
	srv := &http.Server{Addr: "localhost:8080"}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Test server\n")
	})
	go func() {
		<-quit
		if err := srv.ListenAndServeTLS(file, keyFile); err != nil {
			// cannot panic, because this probably is an intentional close
			log.Printf("Httpserver: ListenAndServe() error: %s", err)
		}
		close(done)
	}()

}

func TestClient_NewWithCert(t *testing.T) {

	file, err := ioutil.TempFile("/tmp/", "ca.pem")
	file.WriteString(caCert)
	assert.NoError(t, err)
	defer os.Remove(file.Name())
	keyFile, err := ioutil.TempFile("/tmp/", "key.pem")
	keyFile.WriteString(privateKey)
	assert.NoError(t, err)
	defer os.Remove(keyFile.Name())

	signal.Notify(quit, os.Interrupt)

	startHTTPServer(file.Name(), keyFile.Name())

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

	client, err := New("ca.pem", "/tmp/", 30)
	if err != nil {
		log.Println(err)
		return
	}

	resp, err := client.Get("https://localhost:8080")
	if err != nil {
		log.Println(err)
		return
	}

	_, err = ioutil.ReadAll(resp.Body)
	// Stop http server
	<-done
}

func TestClient_NewWithEmptyCert(t *testing.T) {
	_, err := New("", "/tmp/", 30)
	if err != nil {
		log.Println(err)
		return
	}
	assert.NoError(t, err)
}
