package main

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// Universal/shared variables
var (
	nodes = []string{}
)

// Print client usage instructions to the terminal
func PrintUsage() {
	println()
	println("Usage:")
	println("  bid <amount>   -- Place a bid in the auction")
	println("  result         -- Query the state of the auction")
	println("  usage          -- See this page")
	println("  exit           -- Stop the client")
	println()
}

func Bid(message string) {
	bidSplit := strings.Split(message, " ")
	if len(bidSplit) == 1 {
		println("Usage: bid <amount>")
		return
	}

	amountInt, err := strconv.Atoi(bidSplit[1])
	if err != nil {
		println("Usage: bid <amount>")
		return
	}

	log.Printf("Placing bid of %d DKK...", amountInt)
}

func Result() {
	// TODO: Implement
	log.Println("GETTING RESULT...")
}

func main() {
	if len(os.Args) < 2 {
		// Even though we're really only looking for 1 (or more) arguments, we expect >= 2 because
		// the first (0th) argument always directs to the .exe of the .go file that's run
		println("Usage: go run client.go <node_1> [<node_2> ...]")
		return
	}

	// Parse node argument(s)
	nodes = os.Args[1:]

	// Print usage
	PrintUsage()

	// Continuously read user input
	scanner := bufio.NewScanner(os.Stdin)
	for {
		// Read user input
		scanner.Scan()
		message := scanner.Text()

		time.Sleep(100 * time.Millisecond)
		// ^^Fixes a visual bug when Ctrl+C'ing to stop the client

		if strings.HasPrefix(message, "bid") {
			Bid(message)
		} else if message == "result" {
			Result()
		} else if message == "usage" {
			PrintUsage()
		} else if message == "exit" {
			os.Exit(0)
		} else {
			println("Did not understand: to see usage run 'usage'")
		}
	}
}
