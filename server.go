// Package sping implements secure ping
package sping

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"golang.org/x/net/context"

	pb "github.com/bbengfort/sping/echo"
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
	// Initialize server variables
	s.sequence = make(map[string]int64)
	addr := fmt.Sprintf(":%d", port)

	// Load the certificates from disk
	certificate, err := tls.LoadX509KeyPair("cert/server.crt", "cert/server.key")
	if err != nil {
		return fmt.Errorf("could not load server key pair: %s", err)
	}

	// Create a certificate pool from the certificate authority
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile("cert/sping_example.crt")
	if err != nil {
		return fmt.Errorf("could not read ca certificate: %s", err)
	}

	// Append the client certificates from the CA
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		return errors.New("failed to append client certs")
	}

	// Open a channel on the address for listening
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("could not listen on %s: %s", addr, err)
	}

	// Create the TLS configuration to pass to the GRPC server
	tlsConfig := &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    certPool,
	}

	// Create a new GRPC server with the credentials
	opt := grpc.Creds(credentials.NewTLS(tlsConfig))
	srv := grpc.NewServer(opt)
	pb.RegisterSecurePingServer(srv, s)

	if err := srv.Serve(lis); err != nil {
		return fmt.Errorf("grpc serve error: %s", err)
	}

	return nil
}
