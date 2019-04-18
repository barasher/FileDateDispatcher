package main

import (
	"flag"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	classifier "github.com/barasher/FileDateDispatcher/internal"
)

const (
	retOk          int = 0
	retConfFailure int = 1
	retExecFailure int = 2
)

var loggingLevels = map[string]logrus.Level{
	"debug": logrus.DebugLevel,
	"info":  logrus.InfoLevel,
	"warn":  logrus.WarnLevel,
	"error": logrus.ErrorLevel,
	"fatal": logrus.FatalLevel,
	"panic": logrus.PanicLevel,
}

func main() {
	os.Exit(doMain(os.Args))
}

func doMain(args []string) int {
	cmd := flag.NewFlagSet("Classifier", flag.ContinueOnError)
	from := cmd.String("s", "", "Source folder")
	to := cmd.String("d", "", "Destination folder")
	batchSize := cmd.String("b", "10", "Batch size")
	logLevelLists := make([]string, 0, len(loggingLevels))
	for lvl := range loggingLevels {
		logLevelLists = append(logLevelLists, lvl)
	}
	logLevel := cmd.String("l", "", "Logging level ("+strings.Join(logLevelLists, "|")+")")
	err := cmd.Parse(args[1:])
	if err != nil {
		if err != flag.ErrHelp {
			logrus.Errorf("error while parsing command line arguments: %v", err)
		}
		return retConfFailure
	}

	var opts []func(*classifier.Classifier) error
	if *batchSize == "" {
		logrus.Errorf("No batchSize provided (-b)")
		return retConfFailure
	}
	batchSizeInt, err := strconv.Atoi(*batchSize)
	if err != nil {
		logrus.Errorf("Error during batchSize conversion (%v): %v", *batchSize, err)
		return retConfFailure
	}
	if batchSizeInt <= 0 {
		logrus.Errorf("Wrong batchSize value (%v), must be > 0", batchSizeInt)
		return retConfFailure
	}
	logrus.Infof("BatchSize: %v", batchSizeInt)
	opts = append(opts, classifier.OptBatchSize(uint(batchSizeInt)))

	if *from == "" {
		logrus.Errorf("No source provided (-s)")
		return retConfFailure
	}

	if *to == "" {
		logrus.Errorf("No destination provided (-s)")
		return retConfFailure
	}

	logrus.SetLevel(logrus.InfoLevel)
	if *logLevel != "" {
		lvl, found := loggingLevels[*logLevel]
		if !found {
			logrus.Errorf("Logging level unknown (%v)", *logLevel)
			return retConfFailure
		}
		logrus.SetLevel(lvl)
	}

	c, err := classifier.NewClassifier(opts...)
	if err != nil {
		logrus.Errorf("Error while initializing classifier: %v", err)
		return retExecFailure
	}

	if err := c.Classify(*from, *to); err != nil {
		logrus.Errorf("Error while classifying: %v", err)
		return retExecFailure
	}

	return retOk
}
