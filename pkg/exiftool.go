package exiftool

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/sirupsen/logrus"
)

type Exiftool struct {
}

type FileMetadata struct {
	File   string
	Fields map[string]interface{}
}

var dateFields map[string]string = map[string]string{
	"CreateDate":        "2006:01:02 15:04:05",
	"Media Create Date": "2006:01:02 15:04:05",
}

var NoDateFound error = fmt.Errorf("No data found")

func (fm *FileMetadata) GuessDate() (time.Time, error) {
	for field, pattern := range dateFields {
		if val, found := fm.Fields[field]; found {
			t, err := time.Parse(pattern, val.(string))
			if err != nil {
				return time.Time{}, fmt.Errorf("error when parsing date %v: %v", val.(string), err)
			}
			return t, nil
		}
	}
	return time.Time{}, NoDateFound
}

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
