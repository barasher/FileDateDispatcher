package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConf(t *testing.T) {
	var tcs = []struct {
		tcId            string
		confFile        string
		expError        bool
		expLoggingLevel string
		expBatchSize    uint
	}{
		{"nominal", "../testdata/conf/nominal.json", false, "warning", 42},
		{"default", "../testdata/conf/default.json", false, defaultLoggingLevel, defaultBatchSize},
		{"unparsable", "../testdata/conf/unparsable.json", true, "", 0},
		{"nonExisting", "../testdata/conf/nonExisting.json", true, "", 0},
	}

	for _, tc := range tcs {
		t.Run(tc.tcId, func(t *testing.T) {
			c, err := loadConf(tc.confFile)
			assert.Equal(t, tc.expError, err != nil)
			assert.Equal(t, tc.expLoggingLevel, c.LoggingLevel)
			assert.Equal(t, tc.expBatchSize, c.BatchSize)
		})
	}
}

func TestDoMainFailure(t *testing.T) {
	var tcs = []struct {
		tcId    string
		params  []string
		expCode int
	}{
		{"parsing error", []string{"-a"}, retConfFailure},
		{"help", []string{"-h"}, retConfFailure},
		{"no confFile", []string{"-s", "/tmp", "-d", "/tmp"}, retConfFailure},
		{"no source", []string{"-c", "../testdata/conf/default.json", "-d", "/tmp"}, retConfFailure},
		{"no destination", []string{"-c", "../testdata/conf/default.json", "-s", "/tmp"}, retConfFailure},
	}

	for _, tc := range tcs {
		t.Run(tc.tcId, func(t *testing.T) {
			ret := doMain(tc.params)
			assert.Equal(t, tc.expCode, ret)
		})
	}
}
