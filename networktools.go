package networktools

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// SendUDP takes an address and data and uses UDP to send transmit the data
// It is advised to use the Request format provided in standards.go and serialise it using GenerateRequest. These are included within the package to make your life easier.
//
// Example:
//
//	outgoingReq, err := generateRequest(newCamera, RequestSuccessful)
//
//	if err != nil {
//				fmt.Println(err)
//			}
//
//	err = sendUDP(req.Addr.String(), outgoingReq)
//
//	if err != nil {
//		   fmt.Println(err)
//		}
//
//	fmt.Printf("Successfully transmitted")
func SendUDP(target_address string, data []byte) error {
	udpAddr, err := net.ResolveUDPAddr("udp", target_address)
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

// SendInitialTCP is used to start a connection between two machines using TCP.
// It works similarly to SendUDP with the distinction being this function returns a connection.
// The function takes an address and some initial data to send and returns the established TCP connection.
func SendInitialTCP(target_address string, data []byte) (net.Conn, error) {
	// Resolve the TCP address
	tcpAddr, err := net.ResolveTCPAddr("tcp", target_address)
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

func Get_TCP_Reply(conn net.Conn, buff_size uint16) ([]byte, error) {
	conn.SetReadDeadline(time.Now().Add(time.Second))
	buffer := make([]byte, buff_size)
	_, err := conn.Read(buffer)
	if err != nil {
		if err != nil {
			if err == io.EOF {
				return buffer, fmt.Errorf("Connection closed by server")
			} else {
				return buffer, fmt.Errorf("Error reading from connection:", err)
			}
		}
	}
	return buffer, nil

}

// SendTCPReply is a function to reply to a given TCP connection.
// The function takes a given connection and data to send and returns an error value, with nil implying there has been no error.
func SendTCPReply(conn net.Conn, data []byte) error {
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

func Handle_Single_TCP_Exchange(target_addr string, data []byte, buff_size uint16) ([]byte, error) {
	conn, err := SendInitialTCP(target_addr, data)
	if err != nil {
		return nil, err
	}
	buff, err := Get_TCP_Reply(conn, buff_size)
	if err != nil {
		return nil, err
	}
	return buff, nil
}

// GetPublicIP is a function to get the public IP address of the machine.
// The function takes no input and returns a string identifying the IP address.
// It will be noted it is currently relying on a public API so in future there may be bugs / it may not work and will have to be updated.
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

// GetLocalIP is a function to get the Local IP address on the machine's WiFi network.
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
