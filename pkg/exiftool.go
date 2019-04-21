package exiftool

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/sirupsen/logrus"
)

// Exiftool is the exiftool utility wrapper
type Exiftool struct {
}

// FileMetadata is a structure that represents an exiftool extraction
type FileMetadata struct {
	File   string
	Fields map[string]interface{}
}

// NewExiftool instanciates a new Exiftool with configuration functions
func NewExiftool(opts ...func(*Exiftool) error) (*Exiftool, error) {
	e := Exiftool{}
	for _, opt := range opts {
		if err := opt(&e); err != nil {
			return nil, fmt.Errorf("error when configuring exiftool: %v", err)
		}
	}
	return &e, nil
}

// Load extracts metadata from files
func (e *Exiftool) Load(files []string) ([]FileMetadata, error) {
	args := append([]string{"-j", "-q", "-m"}, files...)
	cmd := exec.Command("exiftool", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			logrus.Errorf("Exiftool parameters: %v, stderr: %v, stdout: %v", args, stderr.String(), stdout.String())
			return []FileMetadata{}, fmt.Errorf("error when invoking exiftool: %v", err)
		}
	}

	if stderr.Len() > 0 {
		return []FileMetadata{}, fmt.Errorf("error during exiftool extraction: %v", stderr.String())
	}
	if stdout.Len() == 0 {
		return []FileMetadata{}, fmt.Errorf("error during exiftool extraction: no output on stdout")
	}

	return e.parseResult(stdout)
}

func (*Exiftool) parseResult(buf bytes.Buffer) ([]FileMetadata, error) {
	var data []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &data); err != nil {
		return []FileMetadata{}, fmt.Errorf("error during json unmarshaling: %v", err)
	}

	res := make([]FileMetadata, len(data))
	i := 0
	for _, curElt := range data {
		res[i] = FileMetadata{File: curElt["SourceFile"].(string), Fields: curElt}
		i++
	}
	return res, nil
}
