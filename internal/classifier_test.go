package classifier

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/barasher/go-exiftool"
)

func checkExist(t *testing.T, path string, shouldExist bool) {
	_, err := os.Stat(path)
	if shouldExist {
		assert.Nil(t, err)
	} else {
		assert.True(t, os.IsNotExist(err))
	}
}

func TestNewClassifierNominal(t *testing.T) {
	invocation1 := false
	invocation2 := false
	f1 := func(*Classifier) error {
		invocation1 = true
		return nil
	}
	f2 := func(*Classifier) error {
		invocation2 = true
		return nil
	}
	_, err := NewClassifier(f1, f2)
	assert.Nil(t, err)
	assert.True(t, invocation1)
	assert.True(t, invocation2)
}

func TestNewClassifierError(t *testing.T) {
	f := func(*Classifier) error {
		return fmt.Errorf("error")
	}
	_, err := NewClassifier(f)
	assert.NotNil(t, err)
}

func TestListFilesNominal(t *testing.T) {
	var tcs = []struct {
		tcID        string
		folder      string
		expFiles    []string
		expCanceled bool
	}{
		{
			tcID:        "nominal",
			folder:      "../testdata/input/",
			expFiles:    []string{"../testdata/input/20190404_131804.jpg", "../testdata/input/subFolder/20190404_131805.jpg"},
			expCanceled: false,
		},
		{
			tcID:        "nonExistingFolder",
			folder:      "../nonExistingFolder/",
			expFiles:    []string{},
			expCanceled: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.tcID, func(t *testing.T) {

			ctx, cancel := context.WithCancel(context.TODO())
			filesChan := make(chan string, 10)
			var wgGlobal sync.WaitGroup
			wgGlobal.Add(1)

			c := buildDefaultClassifier(t, 1)
			c.listFiles(ctx, cancel, tc.folder, filesChan, &wgGlobal)

			files := make([]string, 10)
			for f := range filesChan {
				files = append(files, f)
			}

			assert.Subset(t, files, tc.expFiles)
			select {
			case <-ctx.Done():
				assert.True(t, tc.expCanceled)
			default:
				assert.False(t, tc.expCanceled)
			}

		})
	}
}

func TestBuildActionsAndPushCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	actionChan := make(chan moveAction, 10)
	defer close(actionChan)

	cancel()
	c := buildDefaultClassifier(t, 2)
	_, err := c.buildActionsAndPush(ctx, []string{"../testdata/input/20190404_131804.jpg"}, actionChan)
	assert.NotNil(t, err)
}

func TestBuildActionsAndPush(t *testing.T) {
	var tcs = []struct {
		tcID       string
		files      []string
		expActions []moveAction
	}{
		{
			tcID:  "nominal",
			files: []string{"../testdata/input/20190404_131804.jpg"},
			expActions: []moveAction{
				{from: "../testdata/input/20190404_131804.jpg", to: "2019_04"},
			},
		}, {
			tcID:       "fileWithoutDate",
			files:      []string{"../testdata/input/subFolder/noDate.txt"},
			expActions: []moveAction{},
		}, {
			tcID: "multiple",
			files: []string{"../testdata/input/20190404_131804.jpg",
				"../testdata/input/subFolder/20190404_131805.jpg",
				"../testdata/input/subFolder/noDate.txt"},
			expActions: []moveAction{
				{from: "../testdata/input/20190404_131804.jpg", to: "2019_04"},
				{from: "../testdata/input/subFolder/20190404_131805.jpg", to: "2019_04"},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.tcID, func(t *testing.T) {
			func() {
				ctx := context.TODO()
				actionChan := make(chan moveAction, 10)

				c := buildDefaultClassifier(t, 2)
				count, err := c.buildActionsAndPush(ctx, tc.files, actionChan)
				close(actionChan)
				assert.Nil(t, err)
				assert.Equal(t, len(tc.expActions), count)

				actions := []moveAction{}
				for ma := range actionChan {
					actions = append(actions, ma)
				}
				assert.Subset(t, actions, tc.expActions)

			}()

		})
	}
}

func TestGetMoveActionsCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	fileChan := make(chan string, 10)
	fileChan <- "../testdata/input/20190404_131804.jpg"
	close(fileChan)
	actionChan := make(chan moveAction, 10)
	var wgGlobal sync.WaitGroup
	wgGlobal.Add(1)

	cancel()
	c := buildDefaultClassifier(t, 2)
	c.getMoveActions(ctx, cancel, fileChan, actionChan, &wgGlobal)

	actionCount := 0
	for range actionChan {
		actionCount++
	}
	assert.Equal(t, 0, actionCount)
}

func TestGetMoveActions(t *testing.T) {
	var tcs = []struct {
		tcID       string
		files      []string
		expActions []moveAction
	}{
		{
			tcID: "nominal",
			files: []string{
				"../testdata/input/20190404_131804.jpg",
				"../testdata/input/subFolder/20190404_131805.jpg",
				"../testdata/input/subFolder/20190404_131806.jpg",
			},
			expActions: []moveAction{
				{from: "../testdata/input/20190404_131804.jpg", to: "2019_04"},
				{from: "../testdata/input/subFolder/20190404_131805.jpg", to: "2019_04"},
				{from: "../testdata/input/subFolder/20190404_131806.jpg", to: "2019_04"},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.tcID, func(t *testing.T) {
			func() {
				ctx, cancel := context.WithCancel(context.TODO())
				fileChan := make(chan string, 10)
				actionChan := make(chan moveAction, 10)
				var wgGlobal sync.WaitGroup
				wgGlobal.Add(1)

				for _, s := range tc.files {
					fileChan <- s
				}
				close(fileChan)

				c := buildDefaultClassifier(t, 2)
				c.getMoveActions(ctx, cancel, fileChan, actionChan, &wgGlobal)

				actions := []moveAction{}
				for ma := range actionChan {
					actions = append(actions, ma)
				}
				assert.Subset(t, actions, tc.expActions)

			}()

		})
	}
}

func TestMoveFiles(t *testing.T) {
	assert.Nil(t, os.MkdirAll("../testdata/tmp/batch/TestMoveFilesNominal/in", 0777))
	assert.Nil(t, copy("../testdata/input/20190404_131804.jpg", "../testdata/tmp/batch/TestMoveFilesNominal/in/20190404_131804.jpg"))

	ctx, cancel := context.WithCancel(context.TODO())
	moveChan := make(chan moveAction, 2)
	moveChan <- moveAction{from: "../testdata/tmp/batch/TestMoveFilesNominal/in/20190404_131804.jpg", to: "2019_04"}
	close(moveChan)
	var wgGlobal sync.WaitGroup
	wgGlobal.Add(1)

	c := buildDefaultClassifier(t, 2)
	c.moveFiles(ctx, cancel, "../testdata/tmp/batch/TestMoveFilesNominal/out", moveChan, &wgGlobal)

	checkExist(t, "../testdata/tmp/batch/TestMoveFilesNominal/in/20190404_131804.jpg", false)
	checkExist(t, "../testdata/tmp/batch/TestMoveFilesNominal/out/2019_04/20190404_131804.jpg", true)
}

func TestClassify(t *testing.T) {
	assert.Nil(t, os.MkdirAll("../testdata/tmp/batch/TestClassify/in/subFolder", 0777))
	assert.Nil(t, copy("../testdata/input/20190404_131804.jpg", "../testdata/tmp/batch/TestClassify/in/subFolder/20190404_131805.jpg"))
	assert.Nil(t, copy("../testdata/input/20190404_131804.jpg", "../testdata/tmp/batch/TestClassify/in/subFolder/20190404_131806.jpg"))
	assert.Nil(t, copy("../testdata/input/subFolder/noDate.txt", "../testdata/tmp/batch/TestClassify/in/subFolder/noDate.txt"))
	assert.Nil(t, copy("../testdata/input/20190404_131804.jpg", "../testdata/tmp/batch/TestClassify/in/20190404_131804.jpg"))

	c := buildDefaultClassifier(t, 2)
	c.Classify("../testdata/tmp/batch/TestClassify/in/", "../testdata/tmp/batch/TestClassify/out/")

	checkExist(t, "../testdata/tmp/batch/TestClassify/in/subFolder/noDate.txt", true)
	checkExist(t, "../testdata/tmp/batch/TestClassify/in/subFolder/20190404_131805.jpg", false)
	checkExist(t, "../testdata/tmp/batch/TestClassify/in/subFolder/20190404_131806.jpg", false)
	checkExist(t, "../testdata/tmp/batch/TestClassify/in/20190404_131804.jpg", false)
	checkExist(t, "../testdata/tmp/batch/TestClassify/out/2019_04/20190404_131804.jpg", true)
	checkExist(t, "../testdata/tmp/batch/TestClassify/out/2019_04/20190404_131805.jpg", true)
	checkExist(t, "../testdata/tmp/batch/TestClassify/out/2019_04/20190404_131806.jpg", true)
}

func TestGuessDateNominal(t *testing.T) {
	fields := map[string]interface{}{
		"a":          "b",
		"CreateDate": "2018:01:02 03:04:05",
	}
	fm := exiftool.FileMetadata{File: "a", Fields: fields}
	c := buildDefaultClassifier(t, 2)
	got, err := c.guessDate(fm)
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
	fm := exiftool.FileMetadata{File: "a", Fields: fields}
	c := buildDefaultClassifier(t, 2)
	_, err := c.guessDate(fm)
	assert.Equal(t, errNoDateFount, err)
}

func TestGuessDateUnparsableDate(t *testing.T) {
	fields := map[string]interface{}{
		"a":          "b",
		"CreateDate": "unparsableDate",
	}
	fm := exiftool.FileMetadata{File: "a", Fields: fields}
	c := buildDefaultClassifier(t, 2)
	_, err := c.guessDate(fm)
	assert.NotNil(t, err)
	assert.NotEqual(t, errNoDateFount, err)
}

func buildDefaultClassifier(t *testing.T, batchSize uint) *Classifier {
	c, err := NewClassifier(
		OptBatchSize(batchSize),
		OptDateFields(map[string]string{"CreateDate": "2006:01:02 15:04:05"}),
	)
	assert.Nil(t, err)
	return c
}
