package main

import (
	"context"
	"log"
	"net"
	"os"
	"sync"

	pb "module/proto"

	"google.golang.org/grpc"
)

// AuctionNode constructor
type AuctionNode struct {
	// Server-related fields
	pb.UnimplementedAuctionNodeServer
	mu   sync.Mutex
	addr string

	// Replication-related fields %% THESE SHOULD BE IN THE CLIENT
	// peerAddresses []string // Here it's important that a node is removed from the slice if it crashes
	// numExpectedReplies int32 %% Equal to the length of peer-slice
	// numReceivedReplies int32

	// Auction-related fields
	is_auction_live    bool
	highest_bid_amount int32
	highest_bidder_id  int32
}

// Start an AuctionNode server. Runs indefinitely
func (n *AuctionNode) StartServer() {
	// Create listener
	lis, err := net.Listen("tcp", n.addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Create and register server
	server := grpc.NewServer()
	pb.RegisterAuctionNodeServer(server, n)

	// Log for transparency
	log.Printf("AuctionNode server now listening on %s", n.addr)

	// Serve
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

// RPC function
func (n *AuctionNode) TryBid(_ context.Context, proposed_bid *pb.Bid) (*pb.Acknowledgement, error) {
	// Needs implementation
	return nil, nil
}

// RPC function
func (n *AuctionNode) TryResult(_ context.Context, _ *pb.Empty) (*pb.Outcome, error) {
	// Needs implementation
	return nil, nil
}

func main() {
	if len(os.Args) < 2 {
		// Even though we're really only looking for 1 argument, we expect 2 or more because
		// the first (0th) argument always directs to the .exe of the .go file that's run
		println("Usage: go run node.go <port_number>")
		return
	}

	port := os.Args[1]

	n := &AuctionNode{
		addr: "localhost:" + port,

		is_auction_live:    false,
		highest_bid_amount: 0,
		highest_bidder_id:  -1,
	}

	n.StartServer()
}
