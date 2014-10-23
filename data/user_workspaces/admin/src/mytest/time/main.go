package main

import (
	"fmt"
	"mytest/time/pkg"
	"time"
)

func main() {
	for i := 0; i < 50; i++ {
		fmt.Println("Hello, 世界", pkg.Now())

		time.Sleep(time.Second)
	}

}
