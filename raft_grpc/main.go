package main

import (
	"flag"
	"fmt"
	"github.com/Jille/raft-grpc-leader-rpc/leaderhealth"
	"github.com/Jille/raftadmin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"raft_grpc/proto"
)

var (
	myAddr        = flag.String("my_addr", "localhost:5000", "TCP address for this node")
	raftID        = flag.String("raft_id", "", "Node ID used by raft")
	raftDir       = flag.String("raft_dir", "/data", "Raft data directory")
	raftBootstrap = flag.Bool("raft_bootstrap", false, "Whether to bootstrap the raft cluster")
)

func main() {
	flag.Parse()

	if *raftID == "" {
		log.Fatal("flag -raft_id is required")
	}

	_, port, err := net.SplitHostPort(*myAddr)
	if err != nil {
		log.Fatalf("failed to parse local address (%q): %v", *myAddr, err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	wt := &wordTracker{}

	ra, tr, err := newRaft(*raftID, *myAddr, wt)
	if err != nil {
		log.Fatalf("failed to start raft: %v", err)
	}

	server := grpc.NewServer()
	proto.RegisterExampleServer(server, &rpcInterface{
		wordTracker: wt,
		raft:        ra,
	})

	tr.Register(server)
	leaderhealth.Setup(ra, server, []string{"Example"})
	raftadmin.Register(server, ra)
	reflection.Register(server)

	if err = server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
