package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

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

	// Start auction
	go n.RunAuction()

	// Serve
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

// Start an auction. Runs for one minute, ends and announces the winner, waits for one more minute, then stops the server
func (n *AuctionNode) RunAuction() {
	n.mu.Lock()
	n.is_auction_live = true
	log.Printf("Auction is live! Accepting bets for the next minute.")
	n.mu.Unlock()

	time.Sleep(1 * time.Minute)

	n.mu.Lock()
	n.is_auction_live = false
	if n.highest_bidder_id == -1 {
		log.Println("Auction ended. No one placed any bids.")
	} else {
		log.Printf("Auction ended. %d placed the highest bid at %d.00 DKK.", n.highest_bidder_id, n.highest_bid_amount)
	}
	n.mu.Unlock()

	time.Sleep(1 * time.Minute)
	log.Println("Server shutting down...")
	os.Exit(0)
}

// RPC function
func (n *AuctionNode) TryBid(_ context.Context, proposed_bid *pb.Bid) (*pb.Acknowledgement, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	announcement := fmt.Sprintf("Client #%d bids %d DKK", proposed_bid.BiddingClientId, proposed_bid.BidAmount)

	// If the auction isn't alive, return an 'Exception'-acknowledgement
	if !n.is_auction_live {
		log.Printf("%s: Returning an Exception-ack", announcement)
		return &pb.Acknowledgement{AckType: pb.Acknowledgement_EXCEPTION}, nil
	}

	// If the auction is alive, and the proposed bid is higher than the current max,
	// overwrite and return a 'Success'-acknowledgement
	if proposed_bid.BidAmount > n.highest_bid_amount {
		n.highest_bid_amount = proposed_bid.BidAmount
		n.highest_bidder_id = proposed_bid.BiddingClientId
		log.Printf("%s: Returning a Success-ack", announcement)
		return &pb.Acknowledgement{AckType: pb.Acknowledgement_SUCCESS}, nil
	}

	// If the auction is alive, but the proposed bid is lower than the current max,
	// simply return a 'Fail'-acknowledgement
	log.Printf("%s: Returning a Fail-ack", announcement)
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

	// Parse port-argument
	port := os.Args[1]

	// Create AuctionNode
	n := &AuctionNode{
		addr: "localhost:" + port,

		is_auction_live:    false,
		highest_bid_amount: 0,
		highest_bidder_id:  -1,
	}

	// Start server
	n.StartServer()
}
