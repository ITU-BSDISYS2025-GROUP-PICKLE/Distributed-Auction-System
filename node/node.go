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

	// Auction-related fields
	is_auction_live    bool
	highest_bid_amount int32
	highest_bidder_id  int32
	bidders            []int32
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
	// Cheesy way of synchronising auction-start across nodes: wait until the same second
	log.Println("Auction starting within the minute...")
	for time.Now().Second() != 0 {
		time.Sleep(time.Second)
	}

	// Flip boolean to true, log
	n.mu.Lock()
	n.is_auction_live = true
	log.Printf("Auction is live! Accepting bets for the next minute.")
	n.mu.Unlock()

	// Wait: receive bids and queries in this time
	time.Sleep(1 * time.Minute)

	// Flip boolean back to false, log
	n.mu.Lock()
	n.is_auction_live = false
	if n.highest_bidder_id == -1 {
		log.Println("Auction ended. No bids were placed.")
	} else {
		log.Printf("Auction ended. %d placed the highest bid at %d.00 DKK.", n.highest_bidder_id, n.highest_bid_amount)
	}
	n.mu.Unlock()

	// Wait again, then shut down the server
	time.Sleep(1 * time.Minute)
	log.Println("Server shutting down...")
	os.Exit(0)
}

// RPC function
func (n *AuctionNode) Bid(_ context.Context, proposed_bid *pb.ProposedBid) (*pb.Acknowledgement, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// First call to bid registers the bidder (this won't be used for anything but is an assignment requirement)
	should_append := true
	for _, b := range n.bidders {
		if b == proposed_bid.BiddingClientId {
			should_append = false
		}
	}
	if should_append {
		n.bidders = append(n.bidders, proposed_bid.BiddingClientId)
		log.Printf("Client #%d registered", proposed_bid.BiddingClientId)
	}

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
func (n *AuctionNode) Result(_ context.Context, result_request *pb.ResultRequest) (*pb.Outcome, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	announcement := fmt.Sprintf("Client #%d is querying the state of the auction", result_request.RequestingClientId)

	// If the auction isn't alive, return a 'Result'-outcome including the highest bidder and bid
	if !n.is_auction_live {
		log.Printf("%s: Returning the result", announcement)
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
	log.Printf("%s: Returning the highest bid so far", announcement)
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
