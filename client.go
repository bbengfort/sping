package sping

import (
	"errors"
	"fmt"
	"log"
	"time"

	"golang.org/x/net/context"

	pb "github.com/bbengfort/pground/sping/echo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// New returns a ping client with the specified options.
func New(name string, delay int64, limit uint) *PingClient {
	return &PingClient{
		Name:     name,
		Delay:    time.Duration(delay) * time.Millisecond,
		Limit:    limit,
		sequence: 0,
	}
}

// PingClient sends echo requests to the ping server on demand.
type PingClient struct {
	Name     string
	Delay    time.Duration
	Limit    uint
	sequence int64
}

// Run the ping client against the server.
func (c *PingClient) Run(addr string) error {

	var idx uint
	ticker := time.NewTicker(c.Delay)
	for range ticker.C {
		idx++
		if idx > c.Limit {
			return nil
		}

		if err := c.Ping(addr); err != nil {
			return err
		}
	}

	return errors.New("run finished unexpectedly")
}

// Ping sends an Ping request to the server and awaits a response.
// Right now we create a new connection for every single ping.
func (c *PingClient) Ping(addr string) error {

	creds, err := credentials.NewClientTLSFromFile("certs/example.crt", "")
	if err != nil {
		return fmt.Errorf("could not load client tls cert: %s", err)
	}

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return err
	}

	client := pb.NewSecurePingClient(conn)
	pong, err := client.Echo(context.Background(), c.Next())
	if err != nil {
		return err
	}

	delta := time.Since(pong.Sent.Parse())
	log.Printf("ping %d/%d to %s took %s", pong.Sseq, pong.Rseq, addr, delta)
	return nil
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
