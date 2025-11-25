package main

import (
	"bufio"
	"context"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "module/proto"
)

// Universal/shared variables
var (
	node_ports = []string{}
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

// Reads and parses user input to the terminal, given a scanner. Includes 'bid', 'result', 'usage' and 'exit' commands
func ReadAndParseInput(scanner *bufio.Scanner) {
	// Read user input
	scanner.Scan()
	message := scanner.Text()

	time.Sleep(100 * time.Millisecond) // Fixes a visual bug when Ctrl+C'ing to stop the client

	// Parse input
	if message == "bid" {
		println("Usage bid amount")
	} else if strings.HasPrefix(message, "bid ") {
		PlaceBid(message)
	} else if message == "result" {
		QueryResult()
	} else if message == "usage" {
		PrintUsage()
	} else if message == "exit" {
		os.Exit(0)
	} else {
		println("Did not understand to see usage run 'usage'")
	}
}

// Dial a node at a given address. Returns an intermediate client which may interact with the server
func DialNode(address string) (pb.AuctionNodeClient, *grpc.ClientConn) {
	// Create connection
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("connect to server failed: %v", err)
	}

	// Create and return client
	return pb.NewAuctionNodeClient(conn), conn
}

// Remove a port from the node_port slice. Returns the new slice if successful, otherwise the unupdated slice
func RemovePortFromSlice(port string) []string {
	for index, p := range node_ports {
		if p == port {
			return append(node_ports[:index], node_ports[index+1:]...)
		}
	}
	return node_ports
}

// Attempt placing a bid on all active nodes
// TODO: Finish description
func PlaceBid(message string) {
	// Split the bid message
	bidSplit := strings.Split(message, " ")
	if len(bidSplit) == 1 {
		println("Usage: bid <amount>")
		return
	}

	// Extract the bid amount from split-bid slice
	amount, err := strconv.Atoi(bidSplit[1])
	if err != nil {
		println("Usage: bid <amount>")
		return
	}

	// Log for transparency
	log.Printf("Placing bid of %d DKK...", amount)

	// Create bid object
	bid := &pb.Bid{
		BiddingClientId: 8008135, // TODO: Implement ClientId registration
		BidAmount:       int32(amount),
	}

	// Send the bid to all known Nodes
	for _, node_port := range node_ports {

		// Dial the Node
		medium, conn := DialNode("localhost:" + node_port)
		defer conn.Close()

		// Attempt to place the bid
		ack, err := medium.TryBid(context.Background(), bid)
		if err != nil {
			// If an error is encountered, remove the node-port from the node-ports slice so it's never called again
			node_ports = RemovePortFromSlice(node_port)
			if len(node_ports) == 0 {
				log.Fatalln("place bid failed: all nodes are down")
			}

			continue
		}

		switch ack.GetAckType() {
		case pb.Acknowledgement_SUCCESS:
			log.Printf("localhost:%v --> Success", node_port)
		case pb.Acknowledgement_FAIL:
			log.Printf("localhost:%v --> Fail", node_port)
		case pb.Acknowledgement_EXCEPTION:
			log.Printf("localhost:%v --> Exception", node_port)
		default:
			log.Fatalf("This should never happen")
		}
	}
}

// Attempt querying the state of the auction from one of the active nodes
func QueryResult() {
	// TODO: Implement
	log.Println("GETTING RESULT...")
}

func main() {
	if len(os.Args) < 2 {
		// Even though we're really only looking for 1 (or more) arguments, we expect at least 2
		// because the first (0th) argument always directs to the .exe of the .go file that's run
		println("Usage: go run client.go <node_port_1> [<node_port_2> ...]")
		return
	}

	// Parse node-port argument(s)
	node_ports = os.Args[1:]

	// Print usage
	PrintUsage()

	// Continuously read user input
	scanner := bufio.NewScanner(os.Stdin)
	for {
		ReadAndParseInput(scanner)
	}
}
