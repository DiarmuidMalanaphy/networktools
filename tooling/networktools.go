package networktools

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// SendUDP takes an address and data and uses UDP to send transmit the data.
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
// Be aware you will have to close the connection yourself. It was chosen to defer this to the programmer so more complex exchanges could be handled.
//
// Example:
//
//	conn, err := SendInitialTCP(target_addr, data)
//	if err != nil {
//		return nil, fmt.Errorf("error in SendInitialTCP: %w", err)
//	}
//	defer conn.Close() // Ensure the connection is closed when we're done
func SendInitialTCP(target_address string, data []byte) (net.Conn, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", target_address)
	if err != nil {
		return nil, fmt.Errorf("error resolving address: %w", err)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, fmt.Errorf("error dialing TCP: %w", err)
	}

	_, err = conn.Write(data)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("error sending initial data: %w", err)
	}

	return conn, nil
}

// Get_TCP_Reply is used get the reply on a TCP connection. The function pairs well with SendInitialTCP, which is why the function Handle_Single_TCP_Exchange is provided.
// Something to note is that the buffer size will have to be defined based on how large you're expecting a given reply to be.
//
// Example:
//
//	buff, err := Get_TCP_Reply(conn, buff_size)
//	if err != nil {
//		return nil, fmt.Errorf("error in Get_TCP_Reply: %w", err)
//	}
func Get_TCP_Reply(conn net.Conn, buff_size uint16) ([]byte, error) {
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	buffer := make([]byte, buff_size)
	n, err := conn.Read(buffer)
	fmt.Printf("Read %d bytes from connection\n", n)

	if n == 0 {
		if err == io.EOF {
			return nil, fmt.Errorf("connection closed by remote")
		} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil, fmt.Errorf("read timeout: no data received within deadline")
		} else if err != nil {
			return nil, fmt.Errorf("error reading from connection: %v", err)
		} else {
			return nil, fmt.Errorf("no data read, but no error reported")
		}
	}

	if err != nil {
		return nil, fmt.Errorf("partial read with error: %v", err)
	}

	return buffer[:n], nil

}

// SendTCPReply is a function to reply to a given TCP connection.
// The function takes a given connection and data to send and returns an error value, with nil implying there has been no error.
//
// Example:
//
//	err := SendTCPReply(previous_conn, data)
//	if err != nil {
//		return nil, err
//	}
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

// Handle_Single_TCP_Exchange handles a single exchange of TCP and then closes the connection.
// You will have to implement more complex exchanges yourself using functions within this package.
//
// Example:
//
//	(Purposefully excluded error handling)
//	garb := NewGarb(8)
//	req, _ := networktools.GenerateRequest(garb, 14)
//	data, _ := networktools.Handle_Single_TCP_Exchange("192.168.1.76:5057", req, 1024)
func Handle_Single_TCP_Exchange(target_addr string, data []byte, buff_size uint16) ([]byte, error) {
	conn, err := SendInitialTCP(target_addr, data)
	if err != nil {
		return nil, fmt.Errorf("error in SendInitialTCP: %w", err)
	}
	defer conn.Close() // Ensure the connection is closed when we're done

	buff, err := Get_TCP_Reply(conn, buff_size)
	if err != nil {
		return nil, fmt.Errorf("error in Get_TCP_Reply: %w", err)
	}

	return buff, nil
}

// GetPublicIP is a function to get the public IP address of the machine.
// The function takes no input and returns a string identifying the IP address.
// It will be noted it is currently relying on a public API so in future there may be bugs / it may not work and will have to be updated.
//
// Example:
//
//	publicIP, err := GetPublicIP()
//	if err != nil {
//		fmt.Println("Error getting public IP:", err)
//	}
//	fmt.Printf("Server Global IP is - %s\n", publicIP)
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
//
// Example:
//
//	localIP, err := GetLocalIP()
//	if err != nil {
//	fmt.Println("Error getting Local IP:", err)
//		}
//	fmt.Printf("Server Local IP is - %s\n", localIP)
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
