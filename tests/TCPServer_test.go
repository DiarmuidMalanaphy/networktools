package testing

import (
	"fmt"
	networktool "github.com/diarmuidmalanaphy/networktools"
	"strings"
	"testing"
	"time"
)

type basic struct {
	Name [20]byte
}

func (b *basic) to_string() string {
	trimmedName := strings.TrimRight(string(b.Name[:]), "\x00")
	return trimmedName
}

func TestTransmission(t *testing.T) {
	port := uint16(5050)
	requestChannel, _ := networktool.Create_TCP_Listener(port)

	// Start data transmission in a goroutine
	go transmit(port)

	// Set a timeout for the test
	timeout := time.After(2 * time.Second)
	success := false

	for {
		select {
		case data := <-requestChannel:

			var basicRequest basic
			err := networktool.DeserialiseData(&basicRequest, data.Request.Payload)
			if err != nil {
				t.Errorf("Error during deserialization: %s", err)
				return
			}

			expected := "tested"
			result := basicRequest.to_string()
			if result != expected {
				t.Errorf("Expected %s, got %s", expected, result)
			} else {
				t.Logf("\nSuccess: received expected data '%s' - TCP server initialisiation, serialisation and deserialisation seems to work", result)
				success = true
				return
			}

		case <-timeout:
			if !success {
				t.Error("Test timed out waiting for data")
			}
			return
		}
	}
}

func transmit(port uint16) {
	time.Sleep(20 * time.Millisecond)

	test := "tested"
	test_data := basic{

		Name: stringToUsername(test),
	}
	req, _ := networktool.GenerateRequest(test_data, 1)
	ip_address := fmt.Sprintf("127.0.0.1:%d", port)
	networktool.Handle_Single_TCP_Exchange(ip_address, req, 1024)

}

func stringToUsername(s string) [20]byte {
	var username [20]byte
	copy(username[:], s)
	return username
}
