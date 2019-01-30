// Package sping is a simple example of a client and server that conduct
// communication using gRPC in a secure fashion with SSL/TLS.
package sping

import (
	"log"
	"time"

	pb "github.com/bbengfort/sping/echo"
)

var logmsgs bool

func init() {
	logmsgs = true
}

// NewClient returns a ping client with the specified options.
func NewClient(dailer Dailer, address, name string, delay int64, limit uint) *PingClient {
	conn, err := dailer(address)
	if err != nil {
		log.Fatalln(err)
	}
	return &PingClient{
		Name:             name,
		Delay:            time.Duration(delay) * time.Millisecond,
		Limit:            limit,
		sequence:         0,
		Connection:       conn,
		SecurePingClient: pb.NewSecurePingClient(conn),
	}
}

// NewServer returns a ping server with specified options.
func NewServer() *PingServer {
	return new(PingServer)
}

// Output a message if logmsgs is true.
func Output(format string, args ...interface{}) {
	if logmsgs {
		log.Printf(format, args...)
	}
}
