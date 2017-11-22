package main

import (
	"flag"
	"github.com/aniketawati/sukdf"
	"fmt"
)

func main() {
	v := flag.Bool("v",false,"verbose")
	flag.Parse()
	pass := flag.Arg(0)
	s := sukdf.New(pass)
	s.Verbose = *v
	x, _ := s.Compute()
	fmt.Printf("hash: %x\n",x)
}
