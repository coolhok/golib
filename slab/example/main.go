package main

import (
	"fmt"
	"github.com/cjysmat/golib/slab"
)

func main() {
	arena := slab.NewArena(32, 1<<9, 2., nil)
	fmt.Println("1")
	buf, _ := arena.Alloc(130)
	fmt.Println("2")
	arena.Alloc(131)
	arena.AddRef(buf)
	fmt.Printf("got the buf: len=%d, cap=%d raw=%#v\n", len(buf), cap(buf), buf)
}
