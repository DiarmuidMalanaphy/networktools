package testing

import (
	networktool "github.com/diarmuidmalanaphy/networktools"
	"testing"
	"fmt"
)

type basic struct {
	name [20]byte
}

func (b *basic) to_string() string {
    trimmedName := strings.TrimRight(string(b.name[:]), "\x00")
    return trimmedName
}

func test_transmission(t *testing.T) {
	port := 5050
	request_channel, listener := networktool.Create_TCP_Listener(port)
	go transmit(port)
	for {
		select {
			case data := <- request_channel:
				var basic_request basic
				err := networktool.DeserialiseData(&basic_request, data.Request.Payload)
				if err != nil {
					t.Errorf("Error during deserialisation, %s", err)
					
					}
				expected := "tested"
				if basic_request.name.to_string() != expected {
					t.Errorf("expected %s, got %s", expected, result)
				}
				
}

func transmit(port int) {
	test_data := basic {
	name : stringToUsername(test)
	}
	req, _ := networktool.GenerateRequest(test_data,1)
	ip_address := fmt.Sprintf("127:0:0:1:%d", port)
	data, _ := networktool.Handle_Single_TCP_Exchange(ip_address, req, 1024)
	
}

func stringToUsername(s string) Username {
	var username [20]byte
	copy(username[:], s)
	return username
}
