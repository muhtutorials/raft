package main

import (
	"fmt"
	"github.com/hashicorp/raft"
	"io"
	"strings"
	"sync"
)

type wordTracker struct {
	mu    sync.RWMutex
	words [3]string
}

var _ raft.FSM = &wordTracker{}

func (wt *wordTracker) Apply(l *raft.Log) any {
	wt.mu.Lock()
	defer wt.mu.Unlock()

	words := string(l.Data)
	for i := 0; i < len(wt.words); i++ {
		if compareWords(words, wt.words[i]) {
			// moves the words to the right in the slice
			// in case of "banana" ["orange", "apple", "kiwi"]
			// result is ["banana", "orange", "apple"]
			copy(wt.words[i+1:], wt.words[i:])
			wt.words[i] = words
			break
		}
	}
	return nil
}

func (wt *wordTracker) Snapshot() (raft.FSMSnapshot, error) {
	return &snapshot{words: cloneWords(wt.words)}, nil
}

func (wt *wordTracker) Restore(rc io.ReadCloser) error {
	b, err := io.ReadAll(rc)
	if err != nil {
		return err
	}
	words := strings.Split(string(b), "\n")
	copy(wt.words[:], words)
	return nil
}

type snapshot struct {
	words []string
}

func (s *snapshot) Persist(sink raft.SnapshotSink) error {
	_, err := sink.Write([]byte(strings.Join(s.words, "\n")))
	if err != nil {
		sink.Cancel()
		return fmt.Errorf("sink.Write(): %v", err)
	}
	return sink.Close()
}

func (s *snapshot) Release() {}
