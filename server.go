// Package sping implements secure ping
package sping

import (
	"fmt"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"golang.org/x/net/context"

	pb "github.com/bbengfort/pground/sping/echo"
)

// PingServer responds to Ping requests and tracks the number of messages
// sent per sender (responding with the correct sequence).
type PingServer struct {
	sync.Mutex
	sequence map[string]int64 // mapping of named hosts to pings received
}

// Echo implements echo.SecurePing
func (s *PingServer) Echo(ctx context.Context, ping *pb.Ping) (*pb.Pong, error) {

	// Lock the server to ensure safety of sequence state
	s.Lock()
	defer s.Unlock()

	// If the sender is not in the sequence, assign it
	if _, ok := s.sequence[ping.Sender]; !ok {
		s.sequence[ping.Sender] = 0
	}

	// If the ping sseq is one, reset the sequence counter
	// Otherwise increment the sequence count accordingly.
	if ping.Sseq == 1 {
		s.sequence[ping.Sender] = 1
	} else {
		s.sequence[ping.Sender]++
	}

	// Success is true if the sequence is not out of order
	success := ping.Sseq == s.sequence[ping.Sender]
	rseq := s.sequence[ping.Sender]

	// Create the reply message
	pong := &pb.Pong{
		Success: success,
		Sseq:    ping.Sseq,
		Rseq:    rseq,
		Sent:    ping.Sent,
	}

	// Log the ping and return
	log.Printf("received ping %d/%d from %s\n", ping.Sseq, rseq, ping.Sender)
	return pong, nil
}

// Serve ping requests from gRPC messages
func (s *PingServer) Serve(port uint) error {
	s.sequence = make(map[string]int64)
	addr := fmt.Sprintf(":%d", port)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("could not listen on %s: %s", addr, err)
	}

	// Create TLS credentials
	creds, err := credentials.NewServerTLSFromFile("certs/example.crt", "certs/example.key")
	if err != nil {
		return fmt.Errorf(("could not create TLS credentials: %s"), err)
	}

	// Create a new GRPC server with the credentials
	srv := grpc.NewServer(grpc.Creds(creds))
	pb.RegisterSecurePingServer(srv, s)

	if err := srv.Serve(lis); err != nil {
		return fmt.Errorf("grpc serve error: %s", err)
	}

	return nil
}
