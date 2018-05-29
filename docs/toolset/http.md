# HTTP package

The `GoSDK v3` integrations should rely on the standard Go `http` package for doing HTTP petitions. However, `GoSDK v3`
provides a helper [HTTP package](https://godoc.org/github.com/newrelic/infra-integrations-sdk/http) to create secure
HTTPS clients that require loading credentials from a Certificate Authority Bundle (stored in a file or in a directory).

The `http.New` method accepts both a bundle directory and a bundle file, but you if you don't want to provide both,
you can omit the parameter by passing an empty string `""`.

Examples:

```go
// Creating an HTTPS client which takes the certificates from a bundle dir,
// without specifying CA bundle file (timeout: 5 seconds)
client1, err := http.New("","/etc/ssl/crt", 5 * time.Second)

// Creating an HTTPS client which takes the certificates from a bundle file,
// without specifying CA bundle directory  (timeout: 10 seconds)
client2, err := http.New("/etc/ssl/crt/myserver.ca-bundle", "", 10 * time.Second)

// Creating an HTTPS client which takes the certificates from both a bundle file
// and a CA bundle directory
client3, err := http.New(
    "/etc/ssl/crt/myserver.ca-bundle",
    "/etc/ssl/crt",
    10 * time.Second)

// Creating a simple HTTP (unsecure) client
client4, err := http.New("", "", 5 * time.Second)
```

For more details, check the [http.New GoDoc](https://godoc.org/github.com/newrelic/infra-integrations-sdk/http#New)
page.
