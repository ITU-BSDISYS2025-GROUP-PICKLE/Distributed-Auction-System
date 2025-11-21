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

	// TODO: Auction timer logic might be started here

	// Serve
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

// RPC function
func (n *AuctionNode) TryBid(_ context.Context, proposed_bid *pb.Bid) (*pb.Acknowledgement, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// If the auction isn't alive, return an 'Exception'-acknowledgement
	if !n.is_auction_live {
		return &pb.Acknowledgement{AckType: pb.Acknowledgement_EXCEPTION}, nil
	}

	// If the auction is alive, and the proposed bid is higher than the current max,
	// overwrite and return a 'Success'-acknowledgement
	if proposed_bid.BidAmount > n.highest_bid_amount {
		n.highest_bid_amount = proposed_bid.BidAmount
		n.highest_bidder_id = proposed_bid.BiddingClientId
		return &pb.Acknowledgement{AckType: pb.Acknowledgement_SUCCESS}, nil
	}

	// If the auction is alive, but the proposed bid is lower than the current max,
	// simply return a 'Fail'-acknowledgement
	return &pb.Acknowledgement{AckType: pb.Acknowledgement_FAIL}, nil
}

// RPC function
func (n *AuctionNode) TryResult(_ context.Context, _ *pb.Empty) (*pb.Outcome, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// If the auction isn't alive, return a 'Result'-outcome including the highest bidder and bid
	if !n.is_auction_live {
		return &pb.Outcome{
			OutcomeType: &pb.Outcome_Result{
				Result: &pb.Result{
					HighestBidderId: n.highest_bidder_id,
					HighestBidFinal: n.highest_bid_amount,
				},
			},
		}, nil
	}

	// If the auction is alive, return a 'Highest-bid-so-far'-outcome
	return &pb.Outcome{
		OutcomeType: &pb.Outcome_HighestBidSoFar{
			HighestBidSoFar: n.highest_bid_amount,
		},
	}, nil
}

func main() {
	if len(os.Args) < 2 {
		// Even though we're really only looking for 1 argument, we expect 2 (or more) because
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
