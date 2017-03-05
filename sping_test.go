package sping

import (
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
	client = NewClient("tester", 100, 8)

	go server.ServeMutualTLS(50051)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := client.PingMutualTLS("localhost:50051")
		if err != nil {
			fmt.Println(err)
			break
		}
	}

}

func BenchmarkServerTLS(b *testing.B) {

	logmsgs = false
	server = NewServer()
	client = NewClient("tester", 100, 8)

	go server.ServeTLS(50052)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := client.PingTLS("localhost:50052")
		if err != nil {
			fmt.Println(err)
			break
		}
	}
}

func BenchmarkInsecure(b *testing.B) {

	logmsgs = false
	server = NewServer()
	client = NewClient("tester", 100, 8)

	go server.ServeInsecure(50053)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := client.PingInsecure("localhost:50053")
		if err != nil {
			fmt.Println(err)
			break
		}
	}
}
