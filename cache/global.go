package cache

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/newrelic/infra-integrations-sdk/log"
)

// DefaultDebug default debug mode for the cache.
const (
	DefaultDebug     = false
	cacheDirFilePerm = 0755
	integrationsDir  = "nr-integrations"
)

// TODO delete when removing these globals
var (
	GlobalLog = log.NewStdErr(DefaultDebug)
	instance  Cacher
	err       error
)

func SetupLogging(verbose bool) {
	GlobalLog.SetDebug(verbose)
}

// DefaultPath returns a default folder/filename path to a cache for an integration from the given name. The name of
// the file will be the name of the integration with the .json extension.
func DefaultPath(integrationName string) string {
	baseDir := filepath.Join(os.TempDir(), integrationsDir)
	// Create integrations cache directory
	if os.MkdirAll(baseDir, cacheDirFilePerm) != nil {
		baseDir = os.TempDir()
	}
	return filepath.Join(baseDir, fmt.Sprint(integrationName, ".json"))
}

func globalPath() string {
	cachePath := os.Getenv(pathEnvVar)
	if cachePath == "" {
		_, fname := filepath.Split(os.Args[0])
		cachePath = filepath.Join(os.TempDir(), fmt.Sprintf("%s-global.json", fname))
	}
	return cachePath
}

func checkInstance() {
	if instance == nil {
		instance, err = NewCache(globalPath(), GlobalLog)
	}
}

// Save marshalls and stores the data the cache is holding into disk as a JSON file.
// Deprecated: don't use global cache. Create your own or let `IntegrationBuilder` create a default cache for you.
func Save() error {
	checkInstance()
	return instance.Save()
}

// Get looks for a key in the cache and returns its value together with the timestamp
// of when it was last set. The third returned value indicates whether
// the key has been found or not.
// Deprecated: don't use global cache. Create your own or let `IntegrationBuilder` create a default cache for you.
func Get(name string) (float64, int64, bool) {
	checkInstance()
	return instance.Get(name)
}

// Set adds a value into the cache together with the current timestamp.
// Deprecated: don't use global cache. Create your own or let `IntegrationBuilder` create a default cache for you.
func Set(name string, value float64) int64 {
	checkInstance()
	return instance.Set(name, value)
}

// Status will return an error if any was found during global Cacher creation.
// Deprecated: don't use global cache. Create your own or let `IntegrationBuilder` create a default cache for you.
func Status() error {
	checkInstance()
	return err
}
