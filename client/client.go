package main

import (
	"bufio"
	"log"
	"os"
	"time"
)

// Read user input
func ReadInput(scanner *bufio.Scanner) string {
	scanner.Scan()
	return scanner.Text()
}

// Print usage. QoL feature
func PrintUsage() {
	println()
	println("Usage:")
	println("  bid <amount>   -- Place a bid in the auction")
	println("  result         -- Query the state of the auction")
	println("  usage          -- See this page")
	println("  exit           -- Stop the client")
	println()
}

func main() {
	if len(os.Args) < 2 {
		// Even though we're really only looking for 1 (or more) arguments, we expect >= 2 because
		// the first (0th) argument always directs to the .exe of the .go file that's run
		println("Usage: go run client.go <node_1> [<node_2> ...]")
		return
	}

	// Parse node argument(s)
	nodes := os.Args[1:]
	for _, node := range nodes { //TODO: Debug
		log.Println(node)
	}

	// Print usage
	PrintUsage()

	// Continuously read user input
	scanner := bufio.NewScanner(os.Stdin)
	for {
		message := ReadInput(scanner)
		time.Sleep(100 * time.Millisecond)
		// ^^Fixes an annoying behaviour where the client terminal prints
		// 5x "Did not understand: to see usage run 'usage'" when Ctrl+C is used to stop the client

		switch message {
		case "bid":
			println("BID\n")
		case "result":
			println("RESULT\n")
		case "usage":
			PrintUsage()
		case "exit":
			os.Exit(0)
		default:
			println("Did not understand: to see usage run 'usage'")
		}
	}
}
