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
func GenerateRequest(data []byte, reqType uint8) ([]byte, error) {

	// Create a Request with the serialized data as the payload
	req := Request_Type{
		Type:    reqType,
		Payload: data,
	}
	serialisedRequest, err := __serialiseRequest(req)
	if err != nil {
		return nil, err
	}

	// Return the serialized request
	return serialisedRequest, nil
}

func DeserialiseData(msg proto.Message, raw_data []byte) error {
	return proto.Unmarshal(raw_data, msg)
}

// This function should not be accessed, but it handles the conversion of a request object to bytes.
func __serialiseRequest(req Request_Type) ([]byte, error) {
	// Convert custom Request_Type to protobuf Request
	protoReq := NewRequest(req)

	// Marshal the protobuf Request to a byte slice
	data, err := proto.Marshal(protoReq)
	if err != nil {
		return nil, err
	}

	return data, nil
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
		Type:    uint8(request.Type), // Note: Converting uint32 to uint8
		Payload: request.Payload,
	}, nil
}

func NewRequest(req Request_Type) *pb.Request {
	return &pb.Request{
		Type:    uint32(req.Type), // Note: Changed to uint32 to match proto3 syntax
		Payload: req.Payload,
	}
}
