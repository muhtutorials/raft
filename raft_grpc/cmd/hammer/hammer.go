// Binary hammer sends requests to your Raft cluster as fast as it can.
// It sends the written out version of the Dutch numbers up to 2000.
// In the end it asks the Raft cluster what the longest three words were.
package main

import (
	"context"
	"fmt"
	grpcRetry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	ntw "moul.io/number-to-words"
	"raft_grpc/proto"
	"sync"
	"time"
)

func main() {
	serviceConfig := `{"healthCheckConfig": {"serviceName": "Example"}, "loadBalancingConfig": [ { "round_robin": {} } ]}`

	retryOpts := []grpcRetry.CallOption{
		grpcRetry.WithBackoff(grpcRetry.BackoffExponential(100 * time.Millisecond)),
		grpcRetry.WithMax(5),
	}

	conn, err := grpc.Dial(
		"multi:///localhost:50051,localhost:50052,localhost:50053",
		grpc.WithDefaultServiceConfig(serviceConfig),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
		grpc.WithUnaryInterceptor(grpcRetry.UnaryClientInterceptor(retryOpts...)),
	)
	if err != nil {
		log.Fatalf("dialing failed: %v", err)
	}
	defer conn.Close()

	client := proto.NewExampleClient(conn)

	ch := generateWords()
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for w := range ch {
				_, err = client.AddWord(context.Background(), &proto.AddWordRequest{Word: w})
				if err != nil {
					log.Fatalf("AddWord RPC failed: %v", err)
				}
			}
		}()
	}
	wg.Wait()

	resp, err := client.GetWords(context.Background(), &proto.GetWordsRequest{})
	if err != nil {
		log.Fatalf("GetWords RPC failed: %v", err)
	}
	fmt.Println(resp)
}

func generateWords() <-chan string {
	ch := make(chan string, 1)
	go func() {
		for i := 0; i < 2000; i++ {
			ch <- ntw.IntegerToNlNl(i)
		}
		close(ch)
	}()
	return ch
}
