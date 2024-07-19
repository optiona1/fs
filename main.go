package main

import (
	"log"
	"myfs/p2p"
)

func main() {
	tcpOpts := p2p.TCPTransportOpts{
		ListenAddr:    ":3000",
		HandshakeFunc: p2p.NOPHandsharkFunc,
		Decoder:       p2p.DefaultDecoder{},
	}

	tr := p2p.NewTCPTransport(tcpOpts)

	if err := tr.ListenAddAccept(); err != nil {
		log.Fatal(err)
	}

	select {}

}
