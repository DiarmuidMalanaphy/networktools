package testing

import (
	"fmt"
	networktool "github.com/diarmuidmalanaphy/networktools"
	"google.golang.org/protobuf/proto"
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

func (b *basic) ConvertToProto() *BasicProto {
	return &BasicProto{
		Name: b.Name[:],
	}
}
func ConvertFromProto(pb *BasicProto) basic {
	var b basic
	copy(b.Name[:], pb.Name)
	return b
}

func SerializeBasic(b *basic) ([]byte, error) {
	return proto.Marshal(b.ConvertToProto())
}

// DeserializeBasic deserializes bytes to a Basic struct using Protocol Buffers
func DeserializeBasic(data []byte) (basic, error) {
	pb := &BasicProto{}
	err := proto.Unmarshal(data, pb)
	if err != nil {
		return basic{}, err
	}
	return ConvertFromProto(pb), nil
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

			//var basicRequest basic
			//err := networktool.DeserialiseData(&basicRequest, data.Request.Payload)
			//if err != nil {
			//	t.Errorf("Error during deserialization: %s", err)
			//	return
			//}
			deserialized, err := DeserializeBasic(data.Request.Payload)
			if err != nil {
				fmt.Println("Deserialization error:", err)
				return
			}
			expected := "tested"
			result := deserialized.to_string()
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
	test_data_basic, _ := SerializeBasic(&test_data)
	req, _ := networktool.GenerateRequest(test_data_basic, 1)
	ip_address := fmt.Sprintf("127.0.0.1:%d", port)
	networktool.Handle_Single_TCP_Exchange(ip_address, req, 1024)

}

func stringToUsername(s string) [20]byte {
	var username [20]byte
	copy(username[:], s)
	return username
}
