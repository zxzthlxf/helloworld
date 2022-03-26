package main

import (
	"fmt"
	"regexp"
	"strings"
)

func main() {
	text := "Hello Gopher, Hello Go Web"
	reg := regexp.MustCompile(`\w+`)
	fmt.Println(reg.MatchString(text))
	fmt.Println(reg.FindAllString(text, -1))

	match, _ := regexp.MatchString("H(.*)d!", "Hello World!")
	fmt.Println(match)
	match, _ = regexp.Match("H(.*)d!", []byte("Hello World!"))
	fmt.Println(match)
	r, _ := regexp.Compile("H(.*)d!")
	fmt.Println(r.MatchString("Hello World!"))

	re := regexp.MustCompile(`who(o*)a(a|m)i`)
	fmt.Printf("%q\n", re.FindStringSubmatch("-whooooaai-whoooai"))
	fmt.Printf("%q\n", re.FindStringSubmatch("-whoami-whoami-"))

	re1 := regexp.MustCompile(`w(a*)i`)
	fmt.Printf("%q\n", re1.FindAllStringSubmatch("-wi-", -1))
	fmt.Printf("%q\n", re1.FindAllStringSubmatch("-waaai-", -1))
	fmt.Printf("%q\n", re1.FindAllStringSubmatch("-wi-wai-", -1))
	fmt.Printf("%q\n", re1.FindAllStringSubmatch("-waai-wi-", -1))

	text = "Hello Gopher, Hello Shirdon"
	reg = regexp.MustCompile("llo")
	fmt.Println(reg.FindStringIndex(text))
	fmt.Println(r.FindAllStringIndex("Hello World!", -1))
	fmt.Println(r.FindStringIndex("Hello World! world"))

	re = regexp.MustCompile(`Go(\w+)`)
	fmt.Println(re.ReplaceAllString("Hello Gopher, Hello GoLang", "Java$1"))

	re = regexp.MustCompile(`w(a*)i`)
	fmt.Printf("%s\n", re.ReplaceAll([]byte("-wi-waaaaai-"), []byte("T")))

	fmt.Printf("%s\n", re.ReplaceAll([]byte("-wi-waaaaai-"), []byte("$1")))

	fmt.Printf("%s\n", re.ReplaceAll([]byte("-wi-waaaaai-"), []byte("$1W")))

	fmt.Printf("%s\n", re.ReplaceAll([]byte("-wi-waaaaai-"), []byte("${1}W")))

	s := "I_Love_Go_Web"
	res := strings.Split(s, "_")
	fmt.Println(res)
	for value := range res {
		fmt.Println(value)
	}

	s1 := "a|b|c|d"
	result := strings.SplitN(s1, "|", 3)
	for v := range result {
		fmt.Println(result[v])
	}

}
