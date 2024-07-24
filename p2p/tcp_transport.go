package p2p

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

// TCPPeer represents the remote node over a TCP established connection.
type TCPPeer struct {
	// TCP 连接
	// conn is the underlying connection of the peer
	// conn net.Conn

	// The underlying connection of the peer. which in this case
	// is a TCP connection.
	net.Conn
	// 数据流动方向，出站还是入站
	// if we dial and retrieve a conn => outbound == true
	// if we accept and retrieve a conn => outbound == false
	outbound bool

	Wg *sync.WaitGroup
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
		Wg:       &sync.WaitGroup{},
	}
}

func (p *TCPPeer) Send(b []byte) error {
	_, err := p.Conn.Write(b)
	if err != nil {
		return err
	}
	return nil
}

type TCPTransportOpts struct {
	ListenAddr    string
	HandshakeFunc HandSharkFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
}

// TCPTransport 管理 TCP 连接的传输层
type TCPTransport struct {
	TCPTransportOpts
	// 用于接受新的连接
	listener net.Listener
	rpcch    chan RPC
}

// NewTCPTransport创建并返回一个新的 TCPTransport 实例，并设置监听地址。
func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcch:            make(chan RPC),
	}
}

// Consume implements the Transport interface, which whill return read-only channel
// for reading the incoming messages received from another peer in the network.
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}

// Close implements the Transport interface.
func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

// Dial implements the Transport interface.
func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	go t.handleConn(conn, true)

	return nil
}

func (t *TCPTransport) ListenAddAccept() error {
	var err error

	// 在指定地址上监听新的 TCP 连接
	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}

	// 监听成功，启用一个新的 goroutine 来处理接受连接的循环
	go t.startAcceptLoop()

	log.Printf("TCP transport listening on port: %s\n", t.ListenAddr)
	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	// 不断接受新的连接。
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			fmt.Printf("TCP accept error: %s\n", err)
		}

		// 每当接受到一个新连接时，启动一个新的 goroutine 来处理连接。
		go t.handleConn(conn, false)
	}

}

// 处理新连接。
func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	var err error

	defer func() {
		fmt.Printf("Dropping peer connection: %s", err)
		conn.Close()
	}()

	peer := NewTCPPeer(conn, outbound)

	if err = t.HandshakeFunc(peer); err != nil {
		return
	}

	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}

	rpc := RPC{}
	for {
		if err := t.Decoder.Decode(conn, &rpc); err != nil {
			continue
		}
		rpc.From = conn.RemoteAddr().String()
		peer.Wg.Add(1)
		t.rpcch <- rpc
		peer.Wg.Wait()
	}
}
