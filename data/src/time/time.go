package main

import (
	"fmt"
	"time"
)

func main() {
	for i := 0; i < 5; i++ {
		fmt.Println("Hello, 世界", i)

		time.Sleep(time.Second)
	}
}
