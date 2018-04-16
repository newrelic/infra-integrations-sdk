package cache

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestDiskCache(t *testing.T) {
	file, err := ioutil.TempFile("", "cache")
	if err != nil {
		t.Error("Can't create temporary cache file")
	}
	defer os.Remove(file.Name())

	// Create cache with existing file in env
	os.Setenv("NRIA_CACHE_PATH", file.Name())
	_, err = NewCache(GlobalLog)
	if err != nil {
		t.Error()
	}

	// Create cache with unexisting file in env
	os.Setenv("NRIA_CACHE_PATH", "newfile.json")
	_, err = NewCache(GlobalLog)
	defer os.Remove("newfile.json")
	if err != nil {
		t.Error()
	}

	// Create cache with default file
	os.Setenv("NRIA_CACHE_PATH", "")
	dc, err := NewCache(GlobalLog)
	if err != nil {
		t.Fail()
	}
	dc.Save()
	_, fname := filepath.Split(os.Args[0])
	expectedPath := filepath.Join(os.TempDir(), fmt.Sprintf("%s.json", fname))
	defer os.Remove(expectedPath)

	_, err = os.Stat(expectedPath)
	if err != nil {
		t.Error()
	}
}

func TestCacheSet(t *testing.T) {
	// Create cache with default file
	os.Setenv("NRIA_CACHE_PATH", "")
	curdir, _ := os.Getwd()

	dc, err := NewCache(GlobalLog)
	defer os.Remove(curdir + "/json")

	if err != nil {
		t.Fatal()
	}

	dc.Set("key", float64(100))
	if dc.Data["key"] != float64(100) {
		t.Error()
	}
	if dc.Timestamps["key"] == 0 {
		t.Error()
	}
}

func TestCacheGet(t *testing.T) {
	// Create cache with default file
	os.Setenv("NRIA_CACHE_PATH", "")
	curdir, _ := os.Getwd()

	dc, err := NewCache(GlobalLog)
	defer os.Remove(curdir + "/json")

	if err != nil {
		t.Fatal()
	}

	dc.Set("key1", float64(100))

	value, ts, exists := dc.Get("key1")

	if value != float64(100) {
		t.Error()
	}
	if ts == int64(0) {
		t.Error()
	}
	if !exists {
		t.Error()
	}

	value, ts, exists = dc.Get("key2")

	if value != float64(0) {
		t.Error()
	}
	if ts != int64(0) {
		t.Error()
	}
	if exists {
		t.Error()
	}
}

func TestCacheSave(t *testing.T) {
	// Create cache with default file
	os.Setenv("NRIA_CACHE_PATH", "json")
	curdir, _ := os.Getwd()

	dc, err := NewCache(GlobalLog)
	defer os.Remove(curdir + "/json")

	if err != nil {
		t.Fail()
	}

	dc.Set("key1", float64(100))
	dc.Set("key2", float64(200))

	err = dc.Save()
	if err != nil {
		t.Error()
	}

	dc, err = NewCache(GlobalLog)
	if err != nil {
		t.Fail()
	}

	value, ts, exists := dc.Get("key1")

	if value != float64(100) {
		t.Error()
	}
	if ts == int64(0) {
		t.Error()
	}
	if !exists {
		t.Error()
	}

	value, ts, exists = dc.Get("key2")

	if value != float64(200) {
		t.Error()
	}
	if ts == int64(0) {
		t.Error()
	}
	if !exists {
		t.Error()
	}
}
