package networktools

import (
	"bytes"
	"encoding/binary"
	"fmt"
	pb "github.com/DiarmuidMalanaphy/networktools/standards"
	"google.golang.org/protobuf/proto"
	"io"
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
	serialisedData, err := __serialiseData(data)
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

// Private method defined to convert a struct to a fixed byte amount able to be transferred over a network.
// In practice this method is primarily used on the Request Datatype but it generalises effectively due to the use of interfaces.
func __serialiseData(data interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)

	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Slice {
		// Handle slice serialization
		for i := 0; i < v.Len(); i++ {
			err := binary.Write(buf, binary.LittleEndian, v.Index(i).Interface())
			if err != nil {
				return nil, err
			}
		}
	} else {
		// Handle non-slice serialization
		err := binary.Write(buf, binary.LittleEndian, data)
		if err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

// DeserialiseData converts a section of raw data to a given struct or slice of structs.
// The use of this method requires an instance of the struct.
//
// Example:
//
//	var frame ImageFrame
//	err := deserialiseData(&frame, req.Request.Payload)
//	//The payload is read inplace into the frame variable
//
// In practice this function should be paired with the DeserialiseRequest function.
func DeserialiseData(data_type interface{}, raw_data []byte) error {
	buf := bytes.NewReader(raw_data)
	v := reflect.ValueOf(data_type)

	// Check if dataType is a pointer
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("data_type must be a pointer")
	}

	v = v.Elem()

	if v.Kind() == reflect.Slice {
		// Handle slice deserialization
		sliceElementType := v.Type().Elem()

		for {
			elemPtr := reflect.New(sliceElementType)
			err := binary.Read(buf, binary.LittleEndian, elemPtr.Interface())
			if err == io.EOF {
				break // End of data
			}
			if err != nil {
				return err
			}
			v.Set(reflect.Append(v, elemPtr.Elem()))
		}
	} else {
		// Handle non-slice deserialization
		err := binary.Read(buf, binary.LittleEndian, data_type)
		if err != nil {
			return err
		}
	}

	return nil
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
