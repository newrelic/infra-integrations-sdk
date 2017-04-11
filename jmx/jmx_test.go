package jmx

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"
)

type nopReadCloser struct {
	io.Reader
}

func (nopReadCloser) Close() error { return nil }

type nopWriterCloser struct {
	io.Writer
}

func (nopWriterCloser) Close() error { return nil }

func TestQuery(t *testing.T) {
	resultMap := map[string]interface{}{
		"testmap": 1.0,
	}

	inBuffer := bytes.NewBufferString("")

	rMap, _ := json.Marshal(resultMap)
	cmdOut = nopReadCloser{bytes.NewBuffer(rMap)}
	cmdIn = nopWriterCloser{inBuffer}

	result, err := Query("testquery")
	if err != nil {
		t.Error()
	}
	if result["testmap"] != resultMap["testmap"] {
		t.Error()
	}
	input, err := inBuffer.ReadString('\n')
	if err != nil {
		t.Error()
	}
	if string(input) != "testquery\n" {
		t.Error()
	}
}

func TestQueryInvalidData(t *testing.T) {
	inBuffer := bytes.NewBufferString("")

	rMap := []byte("not a dict")
	cmdOut = nopReadCloser{bytes.NewBuffer(rMap)}
	cmdIn = nopWriterCloser{inBuffer}

	result, err := Query("testquery")
	if err == nil {
		t.Error()
	}
	if result != nil {
		t.Error()
	}
	input, err := inBuffer.ReadString('\n')
	if err != nil {
		t.Error()
	}
	if string(input) != "testquery\n" {
		t.Error()
	}
}

func TestQueryTimeout(t *testing.T) {
	inBuffer := bytes.NewBufferString("")

	cmdOut, _ = io.Pipe()
	cmdIn = nopWriterCloser{inBuffer}

	result, err := Query("testquery")
	if err == nil {
		t.Error()
	}
	if result != nil {
		t.Error()
	}
	input, err := inBuffer.ReadString('\n')
	if err != nil {
		t.Error()
	}
	if string(input) != "testquery\n" {
		t.Error()
	}
}
