package cache

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/newrelic/infra-integrations-sdk/log"
)

// DefaultDebug default debug mode for the cache.
const DefaultDebug = false

// TODO delete when removing these globals
var GlobalLog = log.NewStdErr(DefaultDebug)

func SetupLogging(verbose bool) {
	GlobalLog.SetDebug(verbose)
}

const integrationsDir = "nr-integrations"
const cacheDirFilePerm = 0755

// DefaultPath returns a default folder/filename path to a cache for an integration from the given name. The name of
// the file will be the name of the integration with the .json extension.
func DefaultPath(integrationName string) string {
	baseDir := filepath.Join(os.TempDir(), integrationsDir)
	// Create integrations cache directory
	err := os.MkdirAll(baseDir, cacheDirFilePerm)
	if err != nil {
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

var instance Cache
var err error

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

// Status will return an error if any was found during global Cache creation.
// Deprecated: don't use global cache. Create your own or let `IntegrationBuilder` create a default cache for you.
func Status() error {
	checkInstance()
	return err
}
