package network

import "net"

type UdpSocketClient struct {
	ClientId   int
	RemoteAddr *net.UDPAddr
}

type UdpBufferTransport struct {
	Buff     []byte
	ClientId int
}
