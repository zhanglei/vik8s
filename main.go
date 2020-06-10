package main

import (
	"github.com/ihaiker/vik8s/cmd"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	cmd.Execute()
}
