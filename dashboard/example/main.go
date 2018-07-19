package main

import (
	"github.com/cjysmat/golib/dashboard"
)

type mydata1 struct {
	i int
}

func (this *mydata1) Data() int {
	this.i++
	return this.i
}

type mydata2 struct {
	i int
}

func (this *mydata2) Data() int {
	this.i += 2
	return this.i
}

func main() {
	d := dashboard.New("test of dashboard", 5)
	g := d.AddGraph("graph1")
	g.AddLine("data1", &mydata1{i: 1})
	g.AddLine("data2", &mydata2{i: 0})
	if err := d.Validate(); err != nil {
		panic(err)
	}
	if err := d.Launch(":8000"); err != nil {
		panic(err)
	}
}
