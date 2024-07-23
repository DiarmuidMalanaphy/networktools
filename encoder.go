package networktools

import (
	"fmt"
	pb "github.com/DiarmuidMalanaphy/networktools/standards"
	"google.golang.org/protobuf/proto"
	"reflect"
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
func GenerateRequest(data interface{}, reqType uint8) ([]byte, error) {
	// First, serialize the data
	fmt.Println("THIS SHOULD 100% be printed")
	serialisedData, err := __serialiseData(data)
	fmt.Println("HereC")
	if err != nil {
		return nil, err
	}

	// Create a Request with the serialized data as the payload
	req := Request_Type{
		Type:    reqType,
		Payload: serialisedData,
	}
	serialisedRequest, err := __serialiseRequest(req)
	if err != nil {
		return nil, err
	}

	// Return the serialized request
	return serialisedRequest, nil
}

func __serialiseData(data_to_be_serialised interface{}) ([]byte, error) {
	// Get the reflect.Type of the input
	t := reflect.TypeOf(data_to_be_serialised)

	// Check if the type is a pointer, and if so, get the element type
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Check if the type has a ToProto method
	_, exists := t.MethodByName("ToProto")
	if !exists {
		return nil, fmt.Errorf("input does not implement ToProto() method")
	}

	// Call the ToProto method
	v := reflect.ValueOf(data_to_be_serialised)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return nil, fmt.Errorf("nil pointer")
	}

	protoMsg := v.MethodByName("ToProto").Call(nil)[0].Interface()

	// Convert to proto.Message
	msg, ok := protoMsg.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("ToProto() does not return a proto.Message")
	}

	// Marshal the protobuf message to a byte slice
	return proto.Marshal(msg)
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
