package main

import "fmt"

type Message interface {
	sending()
}

type User struct {
	name  string
	phone string
}

func (u *User) sending() {
	fmt.Printf("Sending user phone to %s<%s>\n", u.name, u.phone)
}

type admin struct {
	name  string
	phone string
}

func (a *admin) sending() {
	fmt.Printf("Sending admin phone to %s<%s>\n", a.name, a.phone)
}

func main() {
	bill := User{"Barry", "barry@gmail.com"}
	sendMessage(&bill)

	lisa := admin{"Jim", "jim@gmail.com"}
	sendMessage(&lisa)
}

func sendMessage(n Message) {
	n.sending()
}
