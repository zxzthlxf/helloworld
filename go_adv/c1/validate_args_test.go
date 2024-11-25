package main

import (
	"errors"
	"testing"
)

func TestValidateArgs(t *testing.T) {
	// TODO 插入上面定义的tests[]切片
	tests := []struct {
		c   config
		err error
	}{
		{
			c:   config{},
			err: errors.New("Must specify a number greater than 0"),
		},
		{
			c:   config{numTimes: -1},
			err: errors.New("Must specify a number greater than 0"),
		},
		{
			c:   config{numTimes: 10},
			err: nil,
		},
	}

	for _, tt := range tests {
		err := validateArgs(tt.c)
		if tt.err != nil && err.Error() != tt.err.Error() {
			t.Errorf("Excepted error to be: %v, got: %v\n", tt.err, err)
		}
		if tt.err == nil && err != nil {
			t.Errorf("Expected no error, got: %v\n", err)
		}
	}
}
