package classifier

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	exiftool "github.com/barasher/FileDateDispatcher/pkg"

	"github.com/sirupsen/logrus"
)

type moveAction struct {
	from string
	to   string
}

type Classifier struct {
	batchSize uint
}

func NewClassifier(opts ...func(*Classifier) error) (*Classifier, error) {
	c := Classifier{batchSize: 10}
	for _, opt := range opts {
		if err := opt(&c); err != nil {
			return nil, fmt.Errorf("error when configuring classifier: %v", err)
		}
	}
	return &c, nil
}

func OptBatchSize(size uint) func(*Classifier) error {
	return func(c *Classifier) error {
		c.batchSize = size
		return nil
	}
}

func (cl *Classifier) Classify(inputFolder string, outputFolder string) error {
	ctx, cancel := context.WithCancel(context.Background())
	filesChan := make(chan string, cl.batchSize*2)
	actionChan := make(chan moveAction, cl.batchSize)
	var wgGlobal sync.WaitGroup
	wgGlobal.Add(3)

	go cl.listFiles(ctx, cancel, inputFolder, filesChan, &wgGlobal)
	go cl.getMoveActions(ctx, cancel, filesChan, actionChan, &wgGlobal)
	go cl.moveFiles(ctx, cancel, outputFolder, actionChan, &wgGlobal)

	wgGlobal.Wait()
	return nil
}

func (cl *Classifier) listFiles(ctx context.Context, cancel context.CancelFunc, inputFolder string, filesChan chan string, wgGlobal *sync.WaitGroup) {
	defer wgGlobal.Done()
	defer close(filesChan)
	fileCount := 0
	var err2 error

	err2 = filepath.Walk(inputFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error when browsing file %v: %v", path, err)
		}
		if !info.IsDir() {
			select {
			case <-ctx.Done():
				return nil
			case filesChan <- path:
				fileCount++
				logrus.Debugf("New file to extract: %v", path)
			}
		}
		return nil
	})

	if err2 != nil {
		cancel()
		logrus.Errorf("%v", err2)
	}
	logrus.Infof("%v file(s) found", fileCount)
}

func (cl *Classifier) getMoveActions(ctx context.Context, cancel context.CancelFunc, filesChan chan string, actionChan chan moveAction, wgGlobal *sync.WaitGroup) {
	defer wgGlobal.Done()
	defer close(actionChan)
	files := make([]string, cl.batchSize)
	var i uint = 0
	actionCount := 0

	for f := range filesChan {
		select {
		case <-ctx.Done():
			logrus.Infof("getMoveAction canceled")
			return
		default:
			files[i] = f
			if i == cl.batchSize-1 {
				count, err2 := cl.buildActionsAndPush(ctx, files, actionChan)
				if err2 != nil {
					cancel()
					logrus.Errorf("error while pushing: %v", err2)
					return
				} else {
					actionCount += count
				}
				i = 0
			} else {
				i++
			}
		}
	}

	if i > 0 {
		count, err2 := cl.buildActionsAndPush(ctx, files[:i], actionChan)
		if err2 != nil {
			cancel()
			logrus.Errorf("error while pushing: %v", err2)
			return
		} else {
			actionCount += count
		}
	}
	logrus.Infof("%v move(s)", actionCount)
}

func (cl *Classifier) buildActionsAndPush(ctx context.Context, files []string, actionChan chan moveAction) (int, error) {
	logrus.Debugf("Build action batch: %v", files)
	e := exiftool.Exiftool{}
	fms, err := e.Load(files)
	if err != nil {
		return 0, fmt.Errorf("error while building actions: %v", err)
	}

	actionCount := 0
	for _, fm := range fms {
		select {
		case <-ctx.Done():
			return 0, fmt.Errorf("Canceled")
		default:
			if d, err := fm.GuessDate(); err != nil {
				if err != exiftool.NoDateFound {
					logrus.Errorf("error while generating moveAction for %v: %v", fm.File, err)
				}
			} else {
				actionChan <- moveAction{
					from: fm.File,
					to:   d.Format("2006_01"),
				}
				actionCount++
			}
		}
	}

	return actionCount, nil
}

func (cl *Classifier) moveFiles(ctx context.Context, cancel context.CancelFunc, outputFolder string, actionChan chan moveAction, wgGlobal *sync.WaitGroup) {
	defer wgGlobal.Done()
	moveCount := 0
	dirs := make(map[string]bool)
	for ma := range actionChan {
		select {
		case <-ctx.Done():
			logrus.Infof("moveFiles canceled")
		default:
			if _, found := dirs[ma.to]; !found {
				if err := os.MkdirAll(filepath.Join(outputFolder, ma.to), 0777); err != nil {
					logrus.Errorf("error when creating output folder: %v", err)
					continue
				}
				dirs[ma.to] = true
			}
			_, f := filepath.Split(ma.from)
			to := filepath.Join(outputFolder, ma.to, f)
			logrus.Debugf("Moving %v to %v", ma.from, to)
			if err := move(ma.from, to); err != nil {
				logrus.Errorf("error when moving %v to %v: %v", ma.from, to, err)
			} else {
				moveCount++
			}
		}
	}
	logrus.Infof("%v moved file(s)", moveCount)
}

func copy(from, to string) error {
	source, err := os.Open(from)
	if err != nil {
		return err
	}
	defer source.Close()
	destination, err := os.Create(to)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

func move(from, to string) error {
	if err := copy(from, to); err != nil {
		return err
	}
	return os.Remove(from)
}
