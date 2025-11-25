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
	client_id  = -1
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
		println("Did not understand: to see usage run 'usage'")
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

// Attempt placing a bid on all active nodes. If a node is down, its port is removed from the node-port slice and consequently won't be dialed again
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
	bid := &pb.ProposedBid{
		BiddingClientId: int32(client_id),
		BidAmount:       int32(amount),
	}

	// Send the bid to all known Nodes
	for _, node_port := range node_ports {

		// Dial the Node
		medium, conn := DialNode("localhost:" + node_port)
		defer conn.Close()

		// Attempt to place the bid
		ack, err := medium.Bid(context.Background(), bid)
		if err != nil {
			// If an error is encountered, the node-port is removed from the node-ports slice so it's never called again
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
			log.Fatalln("This should never happen")
		}
	}
}

// Attempt querying the state of the auction from one of the active nodes
func QueryResult() {
	// Log for transparency
	log.Println("Querying the state of the auction...")

	// Query the state of the auction from all nodes
	for _, node_port := range node_ports {

		// Dial the Node
		medium, conn := DialNode("localhost:" + node_port)
		defer conn.Close()

		// Create a ResultRequest
		result_request := &pb.ResultRequest{
			RequestingClientId: int32(client_id),
		}

		outcome, err := medium.Result(context.Background(), result_request)
		if err != nil {
			// If an error is encountered, the node-port is removed from the node-ports slice so it's never called again
			node_ports = RemovePortFromSlice(node_port)
			if len(node_ports) == 0 {
				log.Fatalln("query result failed: all nodes are down")
			}

			continue
		}

		if outcome.GetResult() == nil {
			log.Printf("localhost:%s --> Highest current bid: %d DKK", node_port, outcome.GetHighestBidSoFar())
		} else {
			result := outcome.GetResult()

			if result.HighestBidderId == -1 {
				log.Printf("localhost:%s --> Auction has concluded: no bets were placed", node_port)
			} else {
				log.Printf("localhost:%s --> Client #%d won the auction with a bid of %d DKK", node_port, result.HighestBidderId, result.HighestBidFinal)
			}
		}
	}
}

func main() {
	if len(os.Args) < 3 {
		// Even though we're really only looking for 2 (or more) arguments, we expect at least 3
		// because the first (0th) argument always directs to the .exe of the .go file that's run
		println("Usage: go run client.go <client_id> <node_port_1> [<node_port_2> ...]")
		return
	}

	// Parse client ID
	tmp, err := strconv.Atoi(os.Args[1])
	if err != nil {
		println("Usage: go run client.go <client_id> <node_port_1> [<node_port_2> ...]")
	}
	client_id = tmp

	// Parse node-port argument(s)
	node_ports = os.Args[2:]

	// Print usage
	PrintUsage()

	// Continuously read user input
	scanner := bufio.NewScanner(os.Stdin)
	for {
		ReadAndParseInput(scanner)
	}
}
