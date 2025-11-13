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
	highest_bid pb.Bid
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

func (a *Auction) EvaluateBid(ctx context.Context, proposed_bid *pb.Bid) (*pb.Acknowledgement, error) {
	// needs implementation
	return nil, nil
}

func (a *Auction) EvaluateResult(context.Context, *pb.Empty) (*pb.Outcome, error) {
	// needs implementation
	return nil, nil
}

func Main() {
	auction := &Auction{
		address:     "localhost:50051",
		highest_bid: pb.Bid{BidAmount: 0},
	}

	auction.StartServer()
}
