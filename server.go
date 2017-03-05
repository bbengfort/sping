package sping

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"golang.org/x/net/context"

	pb "github.com/bbengfort/sping/echo"
)

// Server Certificates
const (
	ServerCert = "cert/server.crt"
	ServerKey  = "cert/server.key"
	ServerName = "localhost"
	ExampleCA  = "cert/sping_example.crt"
)

// PingServer responds to Ping requests and tracks the number of messages
// sent per sender (responding with the correct sequence).
type PingServer struct {
	sync.Mutex
	sequence map[string]int64 // mapping of named hosts to pings received
	srv      *grpc.Server     // handle to the grpc server
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
	Output("received ping %d/%d from %s\n", ping.Sseq, rseq, ping.Sender)
	return pong, nil
}

// Serve ping requests from gRPC messages
func (s *PingServer) Serve(port uint) error {
	// Initialize server variables
	s.sequence = make(map[string]int64)
	addr := fmt.Sprintf(":%d", port)

	// Load the certificates from disk
	certificate, err := tls.LoadX509KeyPair(ServerCert, ServerKey)
	if err != nil {
		return fmt.Errorf("could not load server key pair: %s", err)
	}

	// Create a certificate pool from the certificate authority
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(ExampleCA)
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
	creds := credentials.NewTLS(&tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    certPool,
	})

	// Create a new GRPC server with the credentials
	s.srv = grpc.NewServer(grpc.Creds(creds))
	pb.RegisterSecurePingServer(s.srv, s)

	if err := s.srv.Serve(lis); err != nil {
		return fmt.Errorf("grpc serve error: %s", err)
	}

	return nil
}

// Shutdown the grpc server instance
func (s *PingServer) Shutdown() {
	s.srv.GracefulStop()
}

// ServeMutualTLS is an alias for Serve. It is mostly here for benchmarking.
func (s *PingServer) ServeMutualTLS(port uint) error {
	return s.Serve(port)
}

// ServeTLS is a helper method for server-side encryption that does not expect
// client authentication or credentials. It is mostly here for benchmarking.
func (s *PingServer) ServeTLS(port uint) error {
	// Initialize server variables
	s.sequence = make(map[string]int64)
	addr := fmt.Sprintf(":%d", port)

	// Create the TLS credentials
	creds, err := credentials.NewServerTLSFromFile(ServerCert, ServerKey)
	if err != nil {
		return fmt.Errorf("could not load TLS keys: %s", err)
	}

	// Create the gRPC server with the gredentials
	s.srv = grpc.NewServer(grpc.Creds(creds))

	// Register the handler object
	pb.RegisterSecurePingServer(s.srv, s)

	// Create the channel to listen on
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("could not listen on %s: %s", addr, err)
	}

	// Serve and Listen
	if err := s.srv.Serve(lis); err != nil {
		return fmt.Errorf("grpc serve error: %s", err)
	}

	return nil
}

// ServeInsecure is a helper method for no server-side encryption.
// It is mostly here for benchmarking.
func (s *PingServer) ServeInsecure(port uint) error {
	// Initialize server variables
	s.sequence = make(map[string]int64)
	addr := fmt.Sprintf(":%d", port)

	// Create the gRPC server
	s.srv = grpc.NewServer()

	// Register the handler object
	pb.RegisterSecurePingServer(s.srv, s)

	// Create the channel to listen on
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("could not listen on %s: %s", addr, err)
	}

	// Serve and Listen
	if err := s.srv.Serve(lis); err != nil {
		return fmt.Errorf("grpc serve error: %s", err)
	}

	return nil
}
