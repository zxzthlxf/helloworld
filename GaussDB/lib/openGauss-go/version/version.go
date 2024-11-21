package main

import (
	"fmt"
)

var (
	version     = "v5.0.0"
	productline = "htrunk1"
	versionid   = ""
)

func main() {
	fmt.Printf("%s-%s.%s\n", version, productline, versionid)
}
