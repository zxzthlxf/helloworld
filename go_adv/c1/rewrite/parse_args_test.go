package main

import (
	"bytes"
	"errors"
	"testing"
)

type testConfig struct {
	args     []string
	err      error
	numTimes int
	result   struct {
		err      error
		numTimes int
	}
}

func TestParseArgs(t *testing.T) {
	// TODO 插入此前定义的test[]切片
	tests := []testConfig{
		{
			args:     []string{"-h"},
			err:      errors.New("flag: help requested"),
			numTimes: 0,
		},
		{
			args:     []string{"10"},
			err:      nil,
			numTimes: 10,
		},
		{
			args:     []string{"-n", "abc"},
			err:      errors.New("invalid value \"abc\" for flag -n: parse error"),
			numTimes: 0,
		},
		{
			args:     []string{"-n", "1", "foo"},
			err:      errors.New("Positional arguments specified"),
			numTimes: 1,
		},
	}

	byteBuf := new(bytes.Buffer)
	for _, tc := range tests {
		c, err := parseArgs(byteBuf, tc.args)
		if tc.result.err == nil && err != nil {
			t.Errorf("Expected nil error, got: %v\n", err)
		}
		if tc.result.err == nil && err.Error() != tc.result.err.Error() {
			t.Errorf("Expected error to be: %v, got: %v\n", tc.result.err, err)
		}
		if c.numTimes != tc.result.numTimes {
			t.Errorf("Expected numTimes to be: %v, got: %v\n", tc.result.numTimes, c.numTimes)
		}
		byteBuf.Reset()
	}
}
