package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"kv_store/service"
	"kv_store/store"
	"log"
	"net/http"
	"os"
	"os/signal"
)

const (
	defaultHTTPAddr = "localhost:11000"
	defaultRaftAddr = "localhost:12000"
)

var (
	inmem    bool
	httpAddr string
	raftAddr string
	joinAddr string
	nodeID   string
)

func init() {
	flag.BoolVar(&inmem, "inmem", false, "Use in-memory storage for Raft")
	flag.StringVar(&httpAddr, "haddr", defaultHTTPAddr, "Set HTTP bind address")
	flag.StringVar(&raftAddr, "raddr", defaultRaftAddr, "Set Raft bind address")
	flag.StringVar(&joinAddr, "join", "", "Set address address, if any")
	flag.StringVar(&nodeID, "id", "", "Node ID. If not set, same as Raft address")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <raft-data-path> \n", os.Args[0])
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()
	if flag.NFlag() == 0 {
		fmt.Fprintf(os.Stderr, "No Raft storage directory specified\n")
		os.Exit(1)
	}

	if nodeID == "" {
		nodeID = raftAddr
	}

	// ensure Raft storage exists
	raftDir := flag.Arg(0)

	if raftDir == "" {
		log.Fatalln("No Raft storage directory specified")
	}

	if err := os.MkdirAll(raftDir, 0700); err != nil {
		log.Fatalf("failed to create path for Raft storage: %s", err.Error())
	}

	s := store.New(inmem)
	s.RaftAddr = raftAddr
	s.RaftDir = raftDir
	if err := s.Open(joinAddr == "", nodeID); err != nil {
		log.Fatalf("failed to open store: %s", err.Error())
	}

	h := service.New(httpAddr, s)
	if err := h.Start(); err != nil {
		log.Fatalf("failed to HTTP service: %s", err.Error())
	}

	// If join was specified make the join request.
	// join address is leader's raft address
	if joinAddr != "" {
		if err := join(joinAddr, raftAddr, nodeID); err != nil {
			log.Fatalf("failed to join node at %s: %s", joinAddr, err.Error())
		}
	}

	// We're up and running!
	log.Printf("Raft started successfully, listening on http://%s", httpAddr)

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	<-terminate
	log.Println("Raft exiting")
}

func join(joinAddr, raftAddr, nodeID string) error {
	b, err := json.Marshal(map[string]string{"addr": raftAddr, "id": nodeID})
	if err != nil {
		return err
	}

	resp, err := http.Post(
		fmt.Sprintf("http://%s/join", joinAddr),
		"application/json",
		bytes.NewReader(b),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
