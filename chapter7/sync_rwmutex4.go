package main

import (
	"sync"
	"time"
)

var m1 *sync.RWMutex

func main() {
	m1 = new(sync.RWMutex)
	go Writing(1)
	go Read(2)
	go Writing(3)
	time.Sleep(2 * time.Second)
}

func Read(i int) {
	println(i, "reading start")
	m1.RLock()
	println(i, "reading")
	time.Sleep(1 * time.Second)
	m1.RUnlock()
	println(i, "reading over")
}

func Writing(i int) {
	println(i, "writing start")
	m1.Lock()
	println(i, "writing")
	time.Sleep(1 * time.Second)
	m1.Unlock()
	println(i, "writing over")
}
