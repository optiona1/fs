package p2p

type HandSharkFunc func(Peer) error

func NOPHandsharkFunc(Peer) error { return nil }
