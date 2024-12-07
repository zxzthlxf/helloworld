package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"time"
)

var totalDuration time.Duration = 5

func getName(r io.Reader, w io.Writer) (string, error) {
	scanner := bufio.NewScanner(r)
	msg := "Your name please? Press the Enter key when done"
	fmt.Fprintln(w, msg)

	scanner.Scan()
	if err := scanner.Err(); err != nil {
		return "", err
	}
	name := scanner.Text()
	if len(name) == 0 {
		return "", errors.New("You entered an empty name")
	}
	return name, nil
}

// TODO 插入上面定义的getNameContext()
func getNameContext(ctx context.Context, r io.Reader, w io.Writer) (string, error) {
	var err error
	name := "Default Name"
	c:=make()

// TODO 插入上面定义的main()
