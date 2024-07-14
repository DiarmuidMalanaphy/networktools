package networktools

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
)

type Request struct {
	Type    uint8
	Payload []byte
}

func NewRequest(Type uint8, Payload []byte) Request {
	req := Request{
		Type,
		Payload,
	}
	return req
}

func GenerateRequest(data interface{}, reqType uint8) ([]byte, error) {
	// First, serialize the data
	serialisedData, err := __serialiseData(data)
	if err != nil {
		return nil, err
	}

	// Create a Request with the serialized data as the payload
	req := newRequest(reqType, serialisedData)
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

//This method converts a section of raw data to a given struct or slice of structs.
//The use of this method requires an instance of the struct.
// Example usage on a singular object ->
// var c Camera
// err := deserialiseData(req.Request.Payload, &c)
// where the payload is read into the c variable.

func DeserialiseData(raw_data []byte, data_type interface{}) error {
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

// This function handles the conversion of raw data into a request type.
// You will have to define a standard of what each request type means and what object you would like to associate with each type.
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
