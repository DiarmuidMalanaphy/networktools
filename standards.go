package networktools

import "net"

type Request struct {
	Type    uint8
	Payload []byte // Raw data, can be interpreted based on the request type
}

func NewRequest(Type uint8, Payload []byte) Request {
	req := Request{
		Type,
		Payload,
	}
	return req
}

// The key distinction between the network data types is the fact that UDP is connectionless
type UDPNetworkData struct {
	Request Request
	Addr    net.Addr
}

type TCPNetworkData struct {
	Request Request
	Conn    net.Conn
}
