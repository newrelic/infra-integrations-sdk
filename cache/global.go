package cache

import "github.com/newrelic/infra-integrations-sdk/log"

// DefaultDebug default debug mode for the cache.
const DefaultDebug = false

// TODO delete when removing these globals
var GlobalLog = log.NewStdErr(DefaultDebug)

var instance, err = NewCache(GlobalLog)

func SetupLogging(verbose bool) {
	GlobalLog.SetDebug(verbose)
}

// Save marshalls and stores the data the cache is holding into disk as a JSON file.
func Save() error {
	return instance.Save()
}

// Get looks for a key in the cache and returns its value together with the timestamp
// of when it was last set. The third returned value indicates whether
// the key has been found or not.
func Get(name string) (float64, int64, bool) {
	return instance.Get(name)
}

// Set adds a value into the cache together with the current timestamp.
func Set(name string, value float64) int64 {
	return instance.Set(name, value)
}

// Status will return an error if any was found during global Cache creation.
func Status() error {
	return err
}
