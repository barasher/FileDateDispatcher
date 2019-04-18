package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoMainFailure(t *testing.T) {
	var tcs = []struct {
		tcId    string
		params  []string
		expCode int
	}{
		{"parsing error", []string{"-a"}, retConfFailure},
		{"help", []string{"-h"}, retConfFailure},
		{"unparsable batch size", []string{"-b", "t"}, retConfFailure},
		{"incorrect batch size", []string{"-b", "-1"}, retConfFailure},
		{"no source", []string{"-d", "/tmp"}, retConfFailure},
		{"no destination", []string{"-s", "/tmp"}, retConfFailure},
		{"wrong logging level", []string{"-l", "b"}, retConfFailure},
	}

	for _, tc := range tcs {
		t.Run(tc.tcId, func(t *testing.T) {
			ret := doMain(tc.params)
			assert.Equal(t, tc.expCode, ret)

		})
	}
}
