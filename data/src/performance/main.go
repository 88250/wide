package main

import (
	"flag"
	"fmt"
	"time"
)

type X struct {
	s []*int
}

func (x X) F1(n int) int {
	if n <= 2 {
		return n
	}
	return x.F1(n-1) + x.F1(n-2)
}

func (x *X) F2(n int) int {
	if n <= 2 {
		return n
	}
	return x.F2(n-1) + x.F2(n-2)
}

func F3(n int) int {
	if n <= 2 {
		return n
	}
	return F3(n-1) + F3(n-2)
}

var n = flag.Int("n", 40, "")

func main() {
	flag.Parse()
	N := *n

	x := &X{}
	start := time.Now()
	x.F1(N)
	end := time.Now()
	fmt.Printf("F1: %v\n", end.Sub(start))

	start = time.Now()
	x.F2(N)
	end = time.Now()
	fmt.Printf("F2: %v\n", end.Sub(start))

	start = time.Now()
	F3(N)
	end = time.Now()
	fmt.Printf("F3: %v\n", end.Sub(start))
}
