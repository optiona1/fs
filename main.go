package main

import (
	"log"
	"myfs/p2p"
)

func main() {
	tr := p2p.NewTCPTransport(":3000")

	if err := tr.ListenAddAccept(); err != nil {
		log.Fatal(err)
	}

	select {}

}
