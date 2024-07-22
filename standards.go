package networktools

import "net"

type Request_Type struct {
	Type    uint8
	Payload []byte // Raw data, can be interpreted based on the request type
}

// The key distinction between the network data types is the fact that UDP is connectionless
type UDPNetworkData struct {
	Request Request_Type
	Addr    net.Addr
}

type TCPNetworkData struct {
	Request Request_Type
	Conn    net.Conn
}

func (d *TCPNetworkData) Get_Addr() net.Addr {
	return d.Conn.RemoteAddr()

}

type TCPListener struct {
	StopCh   chan struct{}
	Listener net.Listener
}

// Method to stop the listener
func (l *TCPListener) Stop() {
	close(l.StopCh)
	if l.Listener != nil {
		l.Listener.Close()

	}
}
