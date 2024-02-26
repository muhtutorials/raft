package main

import (
	"context"
	"github.com/hashicorp/raft"
	"raft_grpc/proto"
	"time"
)

type rpcInterface struct {
	wordTracker *wordTracker
	raft        *raft.Raft
	proto.UnimplementedExampleServer
}

func (r rpcInterface) AddWord(ctx context.Context, req *proto.AddWordRequest) (*proto.AddWordResponse, error) {
	f := r.raft.Apply([]byte(req.GetWord()), time.Second)

	if err := f.Error(); err != nil {
		return nil, err
	}

	return &proto.AddWordResponse{
		CommitIndex: f.Index(),
	}, nil
}

func (r rpcInterface) GetWords(ctx context.Context, req *proto.GetWordsRequest) (*proto.GetWordsResponse, error) {
	r.wordTracker.mu.RLock()
	defer r.wordTracker.mu.RUnlock()

	return &proto.GetWordsResponse{
		BestWords:   cloneWords(r.wordTracker.words),
		ReadAtIndex: r.raft.AppliedIndex(),
	}, nil
}
