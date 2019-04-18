package exiftool

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
			for i, _ := range res {
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

func TestGuessDateNominal(t *testing.T) {
	fields := map[string]interface{}{
		"a":          "b",
		"CreateDate": "2018:01:02 03:04:05",
	}
	fm := FileMetadata{File: "a", Fields: fields}
	got, err := fm.GuessDate()
	assert.Nil(t, err)
	assert.Equal(t, 2018, got.Year())
	assert.Equal(t, time.January, got.Month())
	assert.Equal(t, 2, got.Day())
	assert.Equal(t, 3, got.Hour())
	assert.Equal(t, 4, got.Minute())
	assert.Equal(t, 5, got.Second())
}

func TestGuessDateWithoutDateField(t *testing.T) {
	fields := map[string]interface{}{
		"a": "b",
	}
	fm := FileMetadata{File: "a", Fields: fields}
	_, err := fm.GuessDate()
	assert.Equal(t, NoDateFound, err)
}

func TestGuessDateUnparsableDate(t *testing.T) {
	fields := map[string]interface{}{
		"a":          "b",
		"CreateDate": "unparsableDate",
	}
	fm := FileMetadata{File: "a", Fields: fields}
	_, err := fm.GuessDate()
	assert.NotNil(t, err)
	assert.NotEqual(t, NoDateFound, err)
}
