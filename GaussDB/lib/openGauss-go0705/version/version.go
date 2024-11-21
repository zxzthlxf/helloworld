package main

import (
	"fmt"
)

var (
	version     = "v3.0.0"
	productline = "htrunk7"
	versionid   = ""
)

func main() {
	fmt.Printf("%s-%s.%s\n", version, productline, versionid)
}
