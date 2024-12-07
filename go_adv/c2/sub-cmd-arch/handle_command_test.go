package main

import (
	"bytes"
	"testing"
)

func TestHandleCommand(t *testing.T) {
	usageMessage := `Usage: mync [http|grpc] -h
http: A HTTP client.

http: <options> server

Options:
  -verb string
		HTTP method (default "GET")

grpc: A gRPC client.

grpc: <options> server

Options:
  -body string
		  Body of request
  -method string
          Method to call
`
	// TODO 插入之前的testCofigs
	testConfigs := []struct {
		args   []string
		err    error
		output string
	}{
		// 测试没有为应用程序指定参数时的行为
		{
			args:   []string{},
			err:    errInvalidSubCommand,
			output: "Invalid sub-command specified\n" + usageMessage,
		},
		// 当"-h"被指定为应用程序的参数时测试行为
		{
			args:   []string{"-h"},
			err:    nil,
			output: usageMessage,
		},
		// 当无法识别的子命令发送到应用程序时测试行为
		{
			args:   []string{"foo"},
			err:    errInvalidSubCommand,
			output: "Invalid sub-command specified\n" + usageMessage,
		},
	}

	byteBuf := new(bytes.Buffer)
	for _, tc := range testConfigs {
		err := handleCommand(byteBuf, tc.args)
		if tc.err == nil && err != nil {
			t.Fatalf("Expected nil error, got %v", err)
		}

		if tc.err != nil && err.Error() != tc.err.Error() {
			t.Fatalf("Expected error %v, got %v", tc.err, err)
		}

		if len(tc.output) != 0 {
			gotOutput := byteBuf.String()
			if tc.output != gotOutput {
				t.Errorf("Expected output to be: %#v, Got: %#v", tc.output, gotOutput)
			}
		}
		byteBuf.Reset()
	}
}
