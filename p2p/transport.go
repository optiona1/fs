package p2p

import "net"

// Peer is an interface that represents the remote nodes.
type Peer interface {
	net.Conn
	Send([]byte) error
}

// Transport is anything handles the communication
// between the nodes in the network. This can be of the
// form (TCP, UDP, websockets, ...)
type Transport interface {
	Dial(string) error
	ListenAddAccept() error
	Consume() <-chan RPC
	Close() error
}
