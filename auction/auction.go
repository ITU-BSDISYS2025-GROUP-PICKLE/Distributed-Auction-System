package main

import (
	"fmt"
	"sync"
)

type Auction struct {
	mu sync.Mutex
}

func main() {
	fmt.Println("Auction")
}
