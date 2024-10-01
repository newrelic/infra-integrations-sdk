package persist

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/newrelic/infra-integrations-sdk/v3/log"
)

const (
	storeFileTemplate = "%s-%s.json"
)

// StorePath will handle the location for the persistence.
type StorePath interface {
	GetFilePath() string
	CleanOldFiles()
}

// storePath will handle the location for the persistence.
type storePath struct {
	dir             string
	integrationName string
	integrationID   string
	ilog            log.Logger
	ttl             time.Duration
}

// NewStorePath create a new instance of StorePath
func NewStorePath(integrationName, integrationID, customTempDir string, ilog log.Logger, ttl time.Duration) (StorePath, error) {
	if integrationName == "" {
		return nil, fmt.Errorf("integration name not specified")
	}

	if integrationID == "" {
		return nil, fmt.Errorf("integration id not specified")
	}

	if ttl <= 0 {
		return nil, fmt.Errorf("invalid TTL: %d", ttl)
	}

	return &storePath{
		dir:             tmpIntegrationDir(customTempDir),
		integrationName: integrationName,
		integrationID:   integrationID,
		ilog:            ilog,
		ttl:             ttl,
	}, nil
}

// GetFilePath will return the file for storing integration state.
func (t *storePath) GetFilePath() string {
	return filepath.Join(t.dir, fmt.Sprintf(storeFileTemplate, t.integrationName, t.integrationID))
}

// CleanOldFiles will remove all old files created by this integration.
func (t *storePath) CleanOldFiles() {
	files, err := t.findOldFiles()
	if err != nil {
		t.ilog.Debugf("failed to cleanup old files: %v", err)
		return
	}

	for _, file := range files {
		t.ilog.Debugf("removing store file (%s)", file)
		err := os.Remove(file)
		if err != nil {
			t.ilog.Debugf("failed to remove store file (%s): %v", file, err)
			continue
		}
	}
}

// glob returns the pattern for finding all files for the same integration name.
func (t *storePath) glob() string {
	return filepath.Join(t.dir, fmt.Sprintf(storeFileTemplate, t.integrationName, "*"))
}

func (t *storePath) findOldFiles() ([]string, error) {
	var result []string
	// List all files by pattern: /tmp/nr-integrations/com.newrelic.nginx-*.json
	files, err := filepath.Glob(t.glob())
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if file == t.GetFilePath() {
			continue
		}

		fileStat, err := os.Stat(file)
		if err != nil {
			continue
		}

		if now().Sub(fileStat.ModTime()) > t.ttl {
			t.ilog.Debugf("store file (%s) is older than %v", fileStat.Name(), t.ttl)
			result = append(result, file)
		}
	}
	return result, nil
}
