package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
)

func Get(url string) (result string, err error) {
	resp, err1 := http.Get(url)
	if err != nil {
		err = err1
		return
	}
	defer resp.Body.Close()
	buf := make([]byte, 4*1024)
	for true {
		n, err := resp.Body.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Println("文件读取完毕")
				break
			} else {
				fmt.Println("resp.Body.Read err = ", err)
				break
			}
		}
		result += string(buf[:n])
	}
	return
}

func SpiderPage(i int, page chan<- int) {
	url := "http://zhannei.baidu.com/cse/site?q=go&cc=jb51.net&p=0" + strconv.Itoa((i-1)*10)
	fmt.Printf("正在爬取第%d个网页\n", i)
	result, err := Get(url)
	if err != nil {
		fmt.Println("http.Get err = ", err)
		return
	}
	filename := "page" + strconv.Itoa(i) + ".html"
	f, err1 := os.Create(filename)
	if err != nil {
		fmt.Println("os.Create err = ", err1)
		return
	}
	f.WriteString(result)
	f.Close()
	page <- i
}

func Run(start, end int) {
	fmt.Printf("正在爬取第%d页到第%d页\n", start, end)
	page := make(chan int)
	for i := start; i <= end; i++ {
		go SpiderPage(i, page)
	}
	for i := start; i <= end; i++ {
		fmt.Printf("第%d个页面爬取完成\n", <-page)
	}
}

func main() {
	var start, end int
	fmt.Printf("请输入起始页数字>=1:>")
	fmt.Scan(&start)
	fmt.Printf("请输入结束页数字:> ")
	fmt.Scan(&end)
	Run(start, end)
}
