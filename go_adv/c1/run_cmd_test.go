package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestRunCmd(t *testing.T) {
	// TODO 插入之前定义的tests[]切片
	tests := []struct {
		c      config
		input  string
		output string
		err    error
	}{
		{
			c:      config{printUsage: true},
			output: usageString,
		},
		{
			c:      config{numTimes: 5},
			input:  "",
			output: strings.Repeat("Your name please? Press the Enter key when done.\n", 1),
			err:    errors.New("You didn't enter your name"),
		},
		{
			c:      config{numTimes: 5},
			input:  "Bill Bryson",
			output: "Your name please? Press the Enter key when done.\n" + strings.Repeat("Nice to meet you Bill Bryson!\n", 5),
		},
	}
	byteBuf := new(bytes.Buffer)
	for _, tc1 := range tests {
		rd := strings.NewReader(tc1.input)
		err := runCmd(rd, byteBuf, tc1.c)
		if err != nil && tc1.err == nil {
			t.Fatalf("Expected nil error, got: %v\n", err)
		}
		if tc1.err != nil && err.Error() != tc1.err.Error() {
			t.Fatalf("Expected error: %v, Got error: %v\n", tc1.err.Error(), err.Error())
		}
		gotMsg := byteBuf.String()
		if gotMsg != tc1.output {
			t.Errorf("Expected stdout message to be: %v, Got: %v\n", tc1.output, gotMsg)
		}
		byteBuf.Reset()
	}
}
