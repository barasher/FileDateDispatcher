package exiftool

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewExiftoolNominal(t *testing.T) {
	invocation1 := false
	invocation2 := false
	f1 := func(*Exiftool) error {
		invocation1 = true
		return nil
	}
	f2 := func(*Exiftool) error {
		invocation2 = true
		return nil
	}
	_, err := NewExiftool(f1, f2)
	assert.Nil(t, err)
	assert.True(t, invocation1)
	assert.True(t, invocation2)
}

func TestNewExiftoolError(t *testing.T) {
	f := func(*Exiftool) error {
		return fmt.Errorf("error")
	}
	_, err := NewExiftool(f)
	assert.NotNil(t, err)
}

func TestGetLoadNominal(t *testing.T) {
	var tcs = []struct {
		tcId     string
		files    []string
		expFiles []string
	}{
		{
			tcId:     "single",
			files:    []string{"../testdata/input/20190404_131804.jpg"},
			expFiles: []string{"../testdata/input/20190404_131804.jpg"},
		}, {
			tcId:     "multiple",
			files:    []string{"../testdata/input/20190404_131804.jpg", "../testdata/input/subFolder/20190404_131805.jpg"},
			expFiles: []string{"../testdata/input/20190404_131804.jpg", "../testdata/input/subFolder/20190404_131805.jpg"},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.tcId, func(t *testing.T) {
			e := Exiftool{}
			res, err := e.Load(tc.files)
			assert.Nil(t, err)
			for i := range res {
				assert.Subset(t, tc.expFiles, []string{res[i].File})
			}
		})
	}
}

func TestLoadUnexistingFile(t *testing.T) {
	e := Exiftool{}
	_, err := e.Load([]string{"../testdata/shouldNotExist"})
	assert.NotNil(t, err)
}
