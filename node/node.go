package main

import (
	"sync"

	pb "module/proto"
)

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
	isAuctionLive        bool
	highest_bid_amount   int32
	highest_bid_clientId int32
}
