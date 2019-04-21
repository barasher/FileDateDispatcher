package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	classifier "github.com/barasher/FileDateDispatcher/internal"

	_ "github.com/barasher/FileDateDispatcher/pkg"
	"github.com/sirupsen/logrus"
)

const (
	retOk          int = 0
	retConfFailure int = 1
	retExecFailure int = 2

	defaultLoggingLevel string = "info"
	defaultBatchSize    uint   = uint(10)
)

var loggingLevels = map[string]logrus.Level{
	"debug": logrus.DebugLevel,
	"info":  logrus.InfoLevel,
	"warn":  logrus.WarnLevel,
	"error": logrus.ErrorLevel,
	"fatal": logrus.FatalLevel,
	"panic": logrus.PanicLevel,
}

type dateField struct {
	Field   string `json:"field"`
	Pattern string `json:"pattern"`
}

type dispatcherConf struct {
	LoggingLevel string      `json:"loggingLevel"`
	BatchSize    uint        `json:"batchSize"`
	DateFields   []dateField `json:"dateFields"`
}

func main() {
	os.Exit(doMain(os.Args))
}

func doMain(args []string) int {
	cmd := flag.NewFlagSet("Classifier", flag.ContinueOnError)
	from := cmd.String("s", "", "Source folder")
	to := cmd.String("d", "", "Destination folder")
	confFile := cmd.String("c", "", "Configuration file")

	err := cmd.Parse(args[1:])
	if err != nil {
		if err != flag.ErrHelp {
			logrus.Errorf("error while parsing command line arguments: %v", err)
		}
		return retConfFailure
	}

	if *confFile == "" {
		logrus.Errorf("No configuration file provided (-c)")
		return retConfFailure
	}
	conf, err := loadConf(*confFile)
	if err != nil {
		logrus.Errorf("Error during configuration file validation: %v", err)
		return retConfFailure
	}

	if logLvl, found := loggingLevels[conf.LoggingLevel]; !found {
		logrus.Errorf("Unknown logging level specified (%v)", conf.LoggingLevel)
		return retConfFailure
	} else {
		logrus.SetLevel(logLvl)
	}

	var classifierOpts []func(*classifier.Classifier) error
	classifierOpts = append(classifierOpts, classifier.OptBatchSize(conf.BatchSize))
	dfs := map[string]string{}
	for _, v := range conf.DateFields {
		dfs[v.Field] = v.Pattern
	}
	classifierOpts = append(classifierOpts, classifier.OptDateFields(dfs))

	if *from == "" {
		logrus.Errorf("No source provided (-s)")
		return retConfFailure
	}

	if *to == "" {
		logrus.Errorf("No destination provided (-s)")
		return retConfFailure
	}

	c, err := classifier.NewClassifier(classifierOpts...)
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

func loadConf(confFile string) (dispatcherConf, error) {
	c := dispatcherConf{}

	r, err := os.Open(confFile)
	if err != nil {
		return c, fmt.Errorf("Error while opening configuration file %v :%v", confFile, err)
	}
	err = json.NewDecoder(r).Decode(&c)
	if err != nil {
		return c, fmt.Errorf("Error while unmarshaling configuration file %v :%v", confFile, err)
	}

	if c.BatchSize < 1 {
		c.BatchSize = defaultBatchSize
		logrus.Warnf("No batch size specified (or 0), using default (%v)", c.BatchSize)
	}

	if c.LoggingLevel == "" {
		c.LoggingLevel = defaultLoggingLevel
		logrus.Warnf("No logging level specified, using default (%v)", c.LoggingLevel)
	}

	if len(c.DateFields) == 0 {
		return c, fmt.Errorf("No date fields specified in the configuration file")
	}

	return c, nil
}
