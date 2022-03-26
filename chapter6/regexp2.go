package main

import (
	"fmt"
	"regexp"
)

func findPhoneNumber(str string) bool {
	reg := regexp.MustCompile("^1[0-9]{10}")
	res := reg.FindAllString(str, -1)
	if res == nil {
		return false
	}
	return true
}

func findEmail(str string) bool {
	reg := regexp.MustCompile("^[a-zA-Z0-9_]+@[a-zA-Z0-9]+\\.[a-zA-Z0-9]+")
	res := reg.FindAllString(str, -1)
	if res == nil {
		return false
	}
	return true
}

func main() {
	res := findPhoneNumber("13688888888")
	fmt.Println(res)
	res = findPhoneNumber("02888888888")
	fmt.Println(res)
	res = findPhoneNumber("123456")
	fmt.Println(res)

	res1 := findEmail("8888@qq.com")
	fmt.Println(res1)
	res1 = findEmail("shir?don@qq.com")
	fmt.Println(res1)
	res1 = findEmail("8888@qqcom")
	fmt.Println(res1)

}
