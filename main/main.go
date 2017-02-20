package main

import (
	"fmt"

	"github.com/shogg/cantstop"
)

func main() {
	fmt.Print(cantstop.NewSim(1000000).Run())
}
