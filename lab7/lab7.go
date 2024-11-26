package main

import (
	"fmt"
	"sync"
	"time"
)

var doorStatus string
var handStatus string
var mutex sync.Mutex

func hand() {
	mutex.Lock()
	handStatus = "in"
	time.Sleep(time.Millisecond * 200)
	handStatus = "out"
	mutex.Unlock()
	wg.Done()
}

func door() {
	mutex.Lock()
	doorStatus = "close"
	time.Sleep(time.Millisecond * 200)
	if handStatus == "in" {
		fmt.Println("夾到手了啦！")
	} else {
		fmt.Println("沒夾到喔！")
	}
	doorStatus = "open"
	mutex.Unlock()
	wg.Done()
}

var wg sync.WaitGroup

func main() {
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go door()
		go hand()
		wg.Wait()
		time.Sleep(time.Millisecond * 200)
	}
}
