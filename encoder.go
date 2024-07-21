package networktools

import (
	"bytes"
	"encoding/binary"
	"fmt"
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
	req := NewRequest(reqType, serialisedData)
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
func __serialiseRequest(req Request) ([]byte, error) {
	buf := new(bytes.Buffer)

	// Write the Type field
	if err := binary.Write(buf, binary.LittleEndian, req.Type); err != nil {
		return nil, err
	}

	// Write the length of the Payload
	payloadLength := int32(len(req.Payload))
	if err := binary.Write(buf, binary.LittleEndian, payloadLength); err != nil {
		return nil, err
	}

	// Write the Payload bytes
	if _, err := buf.Write(req.Payload); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
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
func DeserialiseRequest(data []byte) (Request, error) {
	var req Request
	buf := bytes.NewReader(data)

	// Read the Type field
	if err := binary.Read(buf, binary.LittleEndian, &req.Type); err != nil {
		return Request{}, err
	}

	// Read the length of the Payload
	var payloadLength int32
	if err := binary.Read(buf, binary.LittleEndian, &payloadLength); err != nil {
		return Request{}, err
	}

	// Read the Payload bytes
	req.Payload = make([]byte, payloadLength)
	if _, err := buf.Read(req.Payload); err != nil {
		return Request{}, err
	}

	return req, nil
}
