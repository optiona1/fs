package main

import (
	"fmt"
	"log"
	"myfs/p2p"
)

func OnPeer(peer p2p.Peer) error {
	fmt.Println("Doing some logic with the peer outside of TCPTransport")
	return nil
}

func main() {
	tcpOpts := p2p.TCPTransportOpts{
		ListenAddr:    ":3000",
		HandshakeFunc: p2p.NOPHandsharkFunc,
		Decoder:       p2p.DefaultDecoder{},
		OnPeer:        OnPeer,
	}

	tr := p2p.NewTCPTransport(tcpOpts)

	go func() {
		for {
			msg := <-tr.Consume()
			fmt.Printf("%+v\n", msg)
		}
	}()

	if err := tr.ListenAddAccept(); err != nil {
		log.Fatal(err)
	}

	select {}

}
