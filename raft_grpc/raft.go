package main

import (
	"fmt"
	transport "github.com/Jille/raft-grpc-transport"
	"github.com/hashicorp/raft"
	boltDB "github.com/hashicorp/raft-boltdb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"path/filepath"
)

func newRaft(myID, myAddr string, fsm raft.FSM) (*raft.Raft, *transport.Manager, error) {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(myID)

	baseDir := filepath.Join(*raftDir, myID)

	logStore, err := boltDB.NewBoltStore(filepath.Join(baseDir, "logs.dat"))
	if err != nil {
		return nil, nil, fmt.Errorf("boltdb.NewBoltStore(%q): %v", filepath.Join(baseDir, "logs.dat"), err)
	}

	stableStore, err := boltDB.NewBoltStore(filepath.Join(baseDir, "stable.dat"))
	if err != nil {
		return nil, nil, fmt.Errorf("boltDB.NewBoltStore(%q): %v", filepath.Join(baseDir, "stable.dat"), err)
	}

	snapshots, err := raft.NewFileSnapshotStore(baseDir, 3, os.Stderr)
	if err != nil {
		return nil, nil, fmt.Errorf("raft.NewFileSnapshotStore(%q, ...): %v", baseDir, err)
	}

	tr := transport.New(raft.ServerAddress(myAddr), []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())})

	ra, err := raft.NewRaft(config, fsm, logStore, stableStore, snapshots, tr.Transport())
	if err != nil {
		return nil, nil, fmt.Errorf("raft.NewRaft: %v", err)
	}

	if *raftBootstrap {
		cfg := raft.Configuration{
			Servers: []raft.Server{
				{
					Suffrage: raft.Voter,
					ID:       raft.ServerID(myID),
					Address:  raft.ServerAddress(myAddr),
				},
			},
		}
		fut := ra.BootstrapCluster(cfg)
		if err = fut.Error(); err != nil {
			return nil, nil, fmt.Errorf("raft.Raft.BootstrapCluster: %v", err)
		}
	}

	return ra, tr, nil
}
