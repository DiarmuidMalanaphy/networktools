package networktools

import (
	"fmt"
	"net"
	"time"
)

type UDPListener struct {
	StopCh chan struct{}
}

// Method to stop the listener
func (l *UDPListener) Stop() {
	close(l.StopCh)
}

// Creates a UDP listener that forwards all requests to a given port to the request channel.
// The request channel is a collection of UDPNetworkData objects defined clearly in the standards file.
// The function will return a UDP listener object that represents the UDP listener routine.
// To stop listening on the UDP port use the Stop command.
//
// Example usage:
//
//	requestChannel := make(chan networktools.UDPNetworkData)
//	listener := Create_UDP_listener(8080, requestChannel)
//	(code code code)
//	listener.Stop (when you're done with the listener)
func Create_UDP_Listener(port uint16) (chan UDPNetworkData, *UDPListener) {
	request_channel := make(chan UDPNetworkData)
	listener := &UDPListener{
		StopCh: make(chan struct{}),
	}

	go listen(port, request_channel, listener.StopCh)

	return request_channel, listener
}

func listen(port uint16, request_channel chan UDPNetworkData, stopCh chan struct{}) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Println("Error resolving address:", err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	defer conn.Close()

	buffer := make([]byte, 1424)

	publicIP, err := GetPublicIP()
	if err != nil {
		fmt.Println("Error getting public IP:", err)
	}

	localIP, err := GetLocalIP()
	if err != nil {
		fmt.Println("Error getting local IP address:", err)
	}

	fmt.Printf("UDP server listening on port %d\n", port)
	fmt.Printf("Server Global IP is - %s\n", publicIP)
	fmt.Printf("Server Local IP is - %s\n", localIP)

	for {
		select {
		case <-stopCh:
			return
		default:
			conn.SetReadDeadline(time.Now().Add(time.Second))
			n, remoteAddr, err := conn.ReadFromUDP(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				fmt.Println("Error reading from UDP:", err)
				continue
			}

			req, err := DeserialiseRequest(buffer[:n])
			if err != nil {
				fmt.Println("Error deserialising request:", err)
				continue
			}

			request_channel <- UDPNetworkData{Request: req, Addr: remoteAddr}
		}
	}
}
