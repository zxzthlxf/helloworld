package main

import (
	"fmt"
	"helloworld/person"
)

func main() {
	s := new(person.Student)
	s.SetName("Shirdon")
	fmt.Println(s.GetName())
}
