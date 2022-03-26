package main

import "fmt"

func SendNotification(user string) chan string {
	notifications := make(chan string, 500)
	go func() {
		notifications <- fmt.Sprintf("Hi %s, welcome to our site!", user)
	}()
	return notifications
}

func main() {
	barry := SendNotification("barry")
	shirdon := SendNotification("shirdon")

	fmt.Println(<-barry)
	fmt.Println(<-shirdon)
}
