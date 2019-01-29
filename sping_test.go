package sping

import (
	"context"
	"fmt"
	"testing"
)

var (
	server *PingServer
	client *PingClient
)

func BenchmarkMutualTLS(b *testing.B) {

	logmsgs = false
	server = NewServer()
	go server.ServeMutualTLS(50051)

	client = NewClient(MutualTLS, "localhost:50051", "tester", 100, 8)
	defer client.Connection.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.Echo(context.Background(), client.Next())
		if err != nil {
			fmt.Printf("failed echo RPC call: %s", err)
			break
		}
	}
}

func BenchmarkServerTLS(b *testing.B) {

	logmsgs = false
	server = NewServer()
	go server.ServeTLS(50052)

	client = NewClient(TLS, "localhost:50052", "tester", 100, 8)
	defer client.Connection.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.Echo(context.Background(), client.Next())
		if err != nil {
			fmt.Printf("failed echo RPC call: %s", err)
			break
		}
	}
}

func BenchmarkInsecure(b *testing.B) {

	logmsgs = false
	server = NewServer()
	go server.ServeInsecure(50053)

	client = NewClient(Insecure, "localhost:50053", "tester", 100, 8)
	defer client.Connection.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := client.Echo(context.Background(), client.Next())
		if err != nil {
			fmt.Printf("failed echo RPC call: %s", err)
			break
		}
	}
}
