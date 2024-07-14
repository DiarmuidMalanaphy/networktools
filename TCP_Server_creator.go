package networktools

import (
	"fmt"
	"io"
	"net"
	"time"
)

type TCPListener struct {
	stopCh   chan struct{}
	listener net.Listener
}

// Method to stop the listener
func (l *TCPListener) Stop() {
	close(l.stopCh)
	if l.listener != nil {
		l.listener.Close()
	}
}

// uses a unique TCP DataType
func Create_TCP_listener(port uint16, request_channel chan<- TCPNetworkData) *TCPListener {
	tcpListener := &TCPListener{
		stopCh: make(chan struct{}),
	}

	go listen_tcp(port, request_channel, tcpListener)

	return tcpListener
}

func listen_tcp(port uint16, request_channel chan<- TCPNetworkData, tcpListener *TCPListener) {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	tcpListener.listener = listener
	defer listener.Close()

	publicIP, err := GetPublicIP()
	if err != nil {
		fmt.Println("Error getting public IP:", err)
	}

	localIP, err := GetLocalIP()
	if err != nil {
		fmt.Println("Error getting local IP address:", err)
	}

	fmt.Printf("TCP server listening on port %d\n", port)
	fmt.Printf("Server Global IP is - %s\n", publicIP)
	fmt.Printf("Server Local IP is - %s\n", localIP)

	for {
		select {
		case <-tcpListener.stopCh:
			return
		default:
			listener.(*net.TCPListener).SetDeadline(time.Now().Add(time.Second))
			conn, err := listener.Accept()
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				fmt.Println("Error accepting connection:", err)
				continue
			}
			go handleTCPConnection(conn, request_channel)
		}
	}
}

func handleTCPConnection(conn net.Conn, request_channel chan<- TCPNetworkData) {
	buffer := make([]byte, 1424) // Arbitrary buffer size

	for {
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		n, err := conn.Read(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			if err == io.EOF {
				fmt.Println("Connection closed by client")
			} else {
				fmt.Println("Error reading from connection:", err)
			}
			conn.Close()
			return
		}

		req, err := DeserialiseRequest(buffer[:n])
		if err != nil {
			fmt.Println("Error deserialising request:", err)
			continue
		}

		request_channel <- TCPNetworkData{
			Request: req,
			Conn:    conn,
		}
	}
}
