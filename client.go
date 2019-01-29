package sping

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"golang.org/x/net/context"

	pb "github.com/bbengfort/sping/echo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Client Certificates
const (
	ClientCert = "cert/client.crt"
	ClientKey  = "cert/client.key"
)

type Dailer func(target string) (*grpc.ClientConn, error)

// PingClient sends echo requests to the ping server on demand.
type PingClient struct {
	Name       string
	Delay      time.Duration
	Limit      uint
	sequence   int64
	Connection *grpc.ClientConn
	pb.SecurePingClient
}

// Next returns the next ping request in the sequence
func (c *PingClient) Next() *pb.Ping {
	c.sequence++

	return &pb.Ping{
		Sender: c.Name,
		Sseq:   c.sequence,
		Sent:   pb.Now(),
		Ttl:    50,
	}
}

// Run the ping client against the server.
func (c *PingClient) Run() error {

	var idx uint
	ticker := time.NewTicker(c.Delay)

	for range ticker.C {
		idx++
		if idx > c.Limit {
			return nil
		}

		pong, err := c.Echo(context.Background(), c.Next())
		if err != nil {
			return fmt.Errorf("failed echo RPC call: %s", err)
		}

		delta := time.Since(pong.Sent.Parse())
		Output("ping %d/%d took %s", pong.Sseq, pong.Rseq, delta)
	}

	return errors.New("run finished unexpectedly")
}

// Ping sends an Ping request to the server and awaits a response.
// Right now we create a new connection for every single ping.
func (c *PingClient) Ping(addr string) (*pb.Pong, error) {

	// Load the certificates from disk
	certificate, err := tls.LoadX509KeyPair(ClientCert, ClientKey)
	if err != nil {
		return nil, fmt.Errorf("could not load client key pair: %s", err)
	}

	// Create a certificate pool from the certificate authority
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(ExampleCA)
	if err != nil {
		return nil, fmt.Errorf("could not read ca certificate: %s", err)
	}

	// Append the client certificates from the CA
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		return nil, errors.New("failed to append ca certs")
	}

	// Create the TLS credentials for transport
	creds := credentials.NewTLS(&tls.Config{
		ServerName:   ServerName,
		Certificates: []tls.Certificate{certificate},
		RootCAs:      certPool,
	})

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("could not connect to %s: %s", addr, err)
	}

	client := pb.NewSecurePingClient(conn)
	pong, err := client.Echo(context.Background(), c.Next())
	if err != nil {
		return nil, fmt.Errorf("failed echo RPC call: %s", err)
	}

	return pong, nil
}

// PingMutualTLS is an alias for Ping. It is mostly here for benchmarking.
func (c *PingClient) PingMutualTLS(addr string) (*pb.Pong, error) {
	return c.Ping(addr)
}

func MutualTLS(addr string) (*grpc.ClientConn, error) {

	// Load the certificates from disk
	certificate, err := tls.LoadX509KeyPair(ClientCert, ClientKey)
	if err != nil {
		return nil, fmt.Errorf("could not load client key pair: %s", err)
	}

	// Create a certificate pool from the certificate authority
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(ExampleCA)
	if err != nil {
		return nil, fmt.Errorf("could not read ca certificate: %s", err)
	}

	// Append the client certificates from the CA
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		return nil, errors.New("failed to append ca certs")
	}

	// Create the TLS credentials for transport
	creds := credentials.NewTLS(&tls.Config{
		ServerName:   ServerName,
		Certificates: []tls.Certificate{certificate},
		RootCAs:      certPool,
	})

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("could not connect to %s: %s", addr, err)
	}
	return conn, nil
}

// PingTLS is a helper method for server-side encryption that does not expect
// client authentication or credentials. It is mostly here for benchmarking.
func TLS(addr string) (*grpc.ClientConn, error) {

	// Create the client TLS credentials
	creds, err := credentials.NewClientTLSFromFile(ServerCert, "")
	if err != nil {
		return nil, fmt.Errorf("could not load tls cert: %s", err)
	}

	// Create a connection with the TLS credentials
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("could not dial %s: %s", addr, err)
	}

	return conn, nil
}

// PingInsecure is a helper method for no server-side encryption.
// It is mostly here for benchmarking.
func Insecure(addr string) (*grpc.ClientConn, error) {
	// Create an insecure connection
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("could not dial %s: %s", addr, err)
	}
	return conn, nil
}
