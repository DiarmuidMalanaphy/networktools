package networktools

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

func SendUDP(address string, data []byte) error {
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return err
	}

	localAddr, err := net.ResolveUDPAddr("udp", ":8000")
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp", localAddr, udpAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	// send the transmission

	_, err = conn.Write(data)

	if err != nil {
		return err
	}
	return nil
}

func SendInitialTCP(address string, data []byte) (*net.TCPConn, error) {
	// Resolve the TCP address
	tcpAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, err
	}

	// Establish a TCP connection
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}

	// Send the data
	_, err = conn.Write(data)
	if err != nil {
		conn.Close() // Close the connection if there's an error sending data
		return nil, err
	}
	defer conn.Close()
	// Return the connection and nil error
	return conn, nil
}

// Distinction between SendTCPReply and SendInitialTCP is if you've already established a connection.
func SendTCPReply(conn *net.TCPConn, data []byte) error {
	if conn == nil {
		return fmt.Errorf("connection is nil")
	}

	// Send the data
	_, err := conn.Write(data)
	if err != nil {
		return fmt.Errorf("error sending data: %w", err)
	}

	return nil
}

func GetPublicIP() (string, error) {
	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(ip), nil
}

func GetLocalIP() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue // Interface down or loopback
		}

		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// Skip loopback and undefined addresses
			if ip == nil || ip.IsLoopback() || ip.To4() == nil {
				continue
			}

			ipStr := ip.String()

			// Check if the IP address is from a common virtual network range
			if !strings.HasPrefix(ipStr, "192.168.56.") {
				return ipStr, nil
			}
		}
	}

	return "", fmt.Errorf("cannot find local IP address")
}
