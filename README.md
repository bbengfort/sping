# Secure Ping

This is a simple example of setting up secured communication with gRPC.

For more, see the write up at &ldquo;[Secure gRPC with TLS/SSL](http://bbengfort.github.io/snippets/2017/03/03/secure-grpc.html)&rdquo;.

**NOTE**: the certificates in this repository are for example only. They were generated on the command-line and only work for the localhost. If you actually wanted to set up this ping across a network, you'd have to generate your own certificates. Here are two good resources for this:

- [Golang TLS examples](https://gist.github.com/denji/12b3a568f092ab951456)
- [certstrap](https://github.com/square/certstrap)

## Demo Quick Start

First clone the repository (using `go get` would work, but the demo only works from the repository working directory) and install the various dependencies:

    $ git clone https://github.com/bbengfort/sping.git
    $ cd sping
    $ go get ./...

Now that you're in the working directory, you'll have to create two processes (the server and the client). I recommend opening up a second terminal tab. In the first tab:

    $ go run cmd/sping serve

This will run the server on the localhost with the default port and options. If you'd like to see all the options for either of the commands just use `--help`. Then in the second terminal:

    $ go run cmd/sping echo localhost

You should see the result in both the server and client windows. You have just conducted a secure Echo RPC request between client and server using [mutual authentication](https://en.wikipedia.org/wiki/Mutual_authentication) with [TLS](https://en.wikipedia.org/wiki/Transport_Layer_Security).

The client will automatically shutdown after 8 messages. Shut the server down with an `INTERRUPT` (CTRL+C).
