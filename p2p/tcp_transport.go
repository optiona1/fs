package p2p

import (
	"fmt"
	"net"
	"sync"
)

// TCPPeer represents the remote node over a TCP established connection.
type TCPPeer struct {
	// TCP 连接
	// conn is the underlying connection of the peer
	conn net.Conn
	// 数据流动方向，出站还是入站
	// if we dial and retrieve a conn => outbound == true
	// if we accept and retrieve a conn => outbound == false
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

// TCPTransport 管理 TCP 连接的传输层
type TCPTransport struct {
	// 监听的地址
	listenAddress string
	// 用于接受新的连接
	listener net.Listener

	sharkHands HandSharkFunc

	decoder Decoder

	// 读写锁，用于接受新的连接
	mu sync.RWMutex
	// 一个存储已连接 p2p 节点的映射，key 是节点地址，值是 Peer 接口
	peers map[net.Addr]Peer
}

// NewTCPTransport创建并返回一个新的 TCPTransport 实例，并设置监听地址。
func NewTCPTransport(listenAddr string) *TCPTransport {
	return &TCPTransport{
		sharkHands:    NOPHandsharkFunc,
		listenAddress: listenAddr,
	}
}

func (t *TCPTransport) ListenAddAccept() error {
	var err error

	// 在指定地址上监听新的 TCP 连接
	t.listener, err = net.Listen("tcp", t.listenAddress)
	if err != nil {
		return err
	}

	// 监听成功，启用一个新的 goroutine 来处理接受连接的循环
	go t.startAcceptLoop()

	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	// 不断接受新的连接。
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
		}

		// 每当接受到一个新连接时，启动一个新的 goroutine 来处理连接。
		go t.handleConn(conn)
	}

}

type Temp struct{}

// 处理新连接。
func (t *TCPTransport) handleConn(conn net.Conn) {
	peer := NewTCPPeer(conn, true)

	if err := t.sharkHands(peer); err != nil {
		fmt.Println("")
	}

	msg := &Temp{}
	for {
		if err := t.decoder.Decode(conn, msg); err != nil {
			fmt.Printf("TCP error: %s", err)
			continue
		}
	}
	fmt.Printf("new incomming connection %+v\n", peer)
}
