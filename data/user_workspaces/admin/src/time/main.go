package main

import (
	"fmt"
	"time"
	"time/pkg"
)

func main() {
	for i := 0; i < 50; i++ {
		fmt.Println("Hello, 世界", pkg.Now())

		time.Sleep(time.Second)
	}

}
