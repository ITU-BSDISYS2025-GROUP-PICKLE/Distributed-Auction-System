package main

import (
	"context"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"

	pb "module/proto"
)

type Auction struct {
	pb.UnimplementedAuctionServer

	mu          sync.Mutex
	address     string
	highest_bid *pb.Bid
	is_over     bool
}

func (a *Auction) StartServer() {
	// Create listener
	lis, err := net.Listen("tcp", a.address)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Create and register server
	server := grpc.NewServer()
	pb.RegisterAuctionServer(server, a)

	// Log for transparency
	log.Printf("Auction now listening on %s", a.address)

	// Serve
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func (a *Auction) EvaluateBid(_ context.Context, proposed_bid *pb.Bid) (*pb.Acknowledgement, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if proposed_bid.BidAmount > a.highest_bid.BidAmount {
		a.highest_bid = proposed_bid
		return &pb.Acknowledgement{Type: pb.Acknowledgement_SUCCESS}, nil
	}

	return &pb.Acknowledgement{Type: pb.Acknowledgement_FAIL}, nil
}

func (a *Auction) EvaluateResult(_ context.Context, _ *pb.Empty) (*pb.Outcome, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.is_over {
		return &pb.Outcome{
			OutcomeType: &pb.Outcome_Result{
				Result: &pb.Result{
					Id:         a.highest_bid.ClientId,
					HighestBid: a.highest_bid.BidAmount},
			},
		}, nil
	}

	return &pb.Outcome{
		OutcomeType: &pb.Outcome_HighestBid{
			HighestBid: a.highest_bid.BidAmount,
		},
	}, nil
}

func Main() {
	auction := &Auction{
		address:     "localhost:50051",
		highest_bid: &pb.Bid{BidAmount: 0},
	}

	auction.StartServer()
}
