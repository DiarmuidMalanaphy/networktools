package networktools

import (
	pb "github.com/DiarmuidMalanaphy/networktools/standards"
	"google.golang.org/protobuf/proto"
)

// GenerateRequest an object or slice of objects, with their request type and serialises them into a byte format that is able to be transmitted over a network.
//
// Example:
//
//	//Excerpt from previous project.
//	var ic ImportedCamera
//	_ := deserialiseData(&ic, req.Request.Payload)
//	newCamera := (Logic to generate camera object)
//	outgoingReq, err := generateRequest(newCamera, RequestSuccessful)
func GenerateRequest(data proto.Message, reqType uint8) ([]byte, error) {
	var serializedRequest []byte
	var err error
	if data != nil {
		// Serialize the proto.Message
		serializedData, err := proto.Marshal(data)
		if err != nil {
			return nil, err
		}
		req := &pb.Request{
			Type:        uint32(reqType),
			PayloadSize: uint64(len(serializedData)),
			Payload:     serializedData,
		}
		serializedRequest, err = proto.Marshal(req)
		if err != nil {
			return nil, err
		}
	} else {
		serializedRequest, err = NewNullRequest(uint32(reqType))
		if err != nil {
			return nil, err
		}

	}

	return serializedRequest, nil
}

func DeserialiseData(msg proto.Message, raw_data []byte) error {
	return proto.Unmarshal(raw_data, msg)
}

// DeserialiseRequest handles the deserialisation of raw data read from a socket into the request standard.
// You will have to pair this with the DeserialiseData function as the meaning of each request type is left to the programmer.
//
// Example:
//
//	n, remoteAddr, err := conn.ReadFromUDP(buffer)
//	// the buffer could be any sort of raw data
//	req, err := deserialiseRequest(buffer[:n])
//	var c Camera
//	err := deserialiseData(&c, req.Request.Payload)
//	cameraMap.removeCamera(c)

func DeserialiseRequest(data []byte) (Request_Type, error) {
	request := &pb.Request{}
	if err := proto.Unmarshal(data, request); err != nil {
		return Request_Type{}, err
	}

	// Convert the protobuf Request to your custom Request_Type
	return Request_Type{
		Type:          uint8(request.Type), // Note: Converting uint32 to uint8
		PayloadLength: request.PayloadSize,
		Payload:       request.Payload,
	}, nil
}

func NewNullRequest(requestType uint32) ([]byte, error) {
	req := &pb.Request{
		Type: requestType,
	}
	return proto.Marshal(req)
}
