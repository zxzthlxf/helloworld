package main

import (
	"errors"
	"testing"
)

type testConfig struct {
	args   []string
	err    error
	result config
}

func TestParseArgs(t *testing.T) {
	// TODO 插入此前定义的test[]切片
	tests := []testConfig{
		{
			args: []string{"-h"},
			err:  nil,
			result: config{
				printUsage: true, numTimes: 0},
		},
		{
			args: []string{"10"},
			err:  nil,
			result: config{
				printUsage: false, numTimes: 10},
		},
		{
			args: []string{"abc"},
			err:  errors.New("strconv.Atoi: parsing \"abc\": invalid syntax"),
			result: config{
				printUsage: false, numTimes: 0},
		},
		{
			args: []string{"1", "foo"},
			err:  errors.New("Invalid number of arguments"),
			result: config{
				printUsage: false, numTimes: 0},
		},
	}

	for _, tc := range tests {
		c, err := parseArgs(tc.args)
		if tc.err != nil && err.Error() != tc.err.Error() {
			t.Fatalf("Expected error to be: %v, got: %v\n", tc.err, err)
		}
		if tc.err == nil && err != nil {
			t.Errorf("Expected nil error, got: %v\n", err)
		}
		if c.printUsage != tc.result.printUsage {
			t.Errorf("Expected printUsage to be: %v, got: %v\n", tc.result.printUsage, c.printUsage)
		}
		if c.numTimes != tc.result.numTimes {
			t.Errorf("Expected numTimes to be: %v, got: %v\n", tc.result.numTimes, c.numTimes)
		}
	}
}
