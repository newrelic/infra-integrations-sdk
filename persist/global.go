package persist

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/newrelic/infra-integrations-sdk/log"
)

// DefaultDebug default debug mode for the Storer.
const (
	DefaultDebug    = false
	dirFilePerm     = 0755
	integrationsDir = "nr-integrations"
)

// TODO delete when removing these globals
var (
	GlobalLog = log.NewStdErr(DefaultDebug)
	instance  Storer
	err       error
)

// SetupLogging setup the global logger verbosity.
func SetupLogging(verbose bool) {
	GlobalLog.SetDebug(verbose)
}

// DefaultPath returns a default folder/filename path to a Storer for an integration from the given name. The name of
// the file will be the name of the integration with the .json extension.
func DefaultPath(integrationName string) string {
	baseDir := filepath.Join(os.TempDir(), integrationsDir)
	// Create integrations Storer directory
	if os.MkdirAll(baseDir, dirFilePerm) != nil {
		baseDir = os.TempDir()
	}
	return filepath.Join(baseDir, fmt.Sprint(integrationName, ".json"))
}

func globalPath() string {
	storePath := os.Getenv(pathEnvVar)
	if storePath == "" {
		_, fname := filepath.Split(os.Args[0])
		storePath = filepath.Join(os.TempDir(), fmt.Sprintf("%s-global.json", fname))
	}
	return storePath
}

func checkInstance() {
	if instance == nil {
		instance, err = NewStorer(globalPath(), GlobalLog)
	}
}

// Save marshalls and stores the data the Storer is holding into disk as a JSON file.
// Deprecated: don't use global Storer. Create your own or let `IntegrationBuilder` create a default Storer for you.
func Save() error {
	checkInstance()
	return instance.Save()
}

// Get looks for a key in the Storer and returns its value together with the timestamp
// of when it was last set. The third returned value indicates whether
// the key has been found or not.
// Deprecated: don't use global Storer. Create your own or let `IntegrationBuilder` create a default Storer for you.
func Get(name string) (float64, int64, bool) {
	checkInstance()
	return instance.Get(name)
}

// Set adds a value into the Storer together with the current timestamp.
// Deprecated: don't use global Storer. Create your own or let `IntegrationBuilder` create a default Storer for you.
func Set(name string, value float64) int64 {
	checkInstance()
	return instance.Set(name, value)
}

// Status will return an error if any was found during global Storer creation.
// Deprecated: don't use global Storer. Create your own or let `IntegrationBuilder` create a default Storer for you.
func Status() error {
	checkInstance()
	return err
}
