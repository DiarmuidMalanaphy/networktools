package networktools

import (
	"fmt"
	"io"
	"net"
	"time"
)

// Creates a TCP listener that forwards all requests to a given port on the request channel.
// The request channel is a collection of TCPNetworkData onjects defined clearly in the standards file.
// The function will return a TCP listener object that represents the TCP listener routeine. To stop listening on the TCP port use the Stop command.
//
// Example Usage:
//
//	request_channel := make(chan networktools.TCPNetworkData)
//	listener := Create_TCP_listener(8080, request_channel)
//	(code code code)
//	listener.Stop (When you're done)
func Create_TCP_listener(port uint16, request_channel chan<- TCPNetworkData) *TCPListener {
	tcpListener := &TCPListener{
		StopCh: make(chan struct{}),
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
	tcpListener.Listener = listener
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
		case <-tcpListener.StopCh:
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
