package p2p

// Peer is an interface that represents the remote nodes.
type Peer interface {
	Close() error
}

// Transport is anything handles the communication
// between the nodes in the network. This can be of the
// form (TCP, UDP, websockets, ...)
type Transport interface {
	ListenAddAccept() error
	Consume() <-chan RPC
}
