package parsing

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"reflect"
)

func ParseEnvelope(conn net.Conn) (*PacketEnvelope, error) {
	header := make([]byte, 3)
	if _, err := io.ReadFull(conn, header); err != nil {
		return nil, fmt.Errorf("failed to read header fields: %w", err)
	}

	headLen := header[0]
	bodyLen := binary.BigEndian.Uint16(header[1:3])
	fmt.Printf("Parsed envelope header: headLen=%d, bodyLen=%d\n", headLen, bodyLen)

	rawHead := make([]byte, headLen)
	if _, err := io.ReadFull(conn, rawHead); err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	rawBody := make([]byte, bodyLen)
	if _, err := io.ReadFull(conn, rawBody); err != nil {
		fmt.Printf("Timed out or failed reading body, expected: %d bytes\n", bodyLen)
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	return &PacketEnvelope{
		HeadLen: headLen,
		BodyLen: bodyLen,
		RawHead: rawHead,
		RawBody: rawBody,
	}, nil
}

// ParseHeader parses raw header bytes into a user-defined struct with bmux tags,
// and extracts the field tagged with `bmux:"msg_id"` for routing.
func ParseHeader(rawHead []byte, headerPtr any) (msgID uint16, err error) {
	v := reflect.ValueOf(headerPtr)
	if v.Kind() != reflect.Pointer || v.Elem().Kind() != reflect.Struct {
		return 0, errors.New("headerPtr must be a pointer to a struct")
	}

	r := bytes.NewReader(rawHead)
	structVal := v.Elem()
	structType := structVal.Type()

	for i := 0; i < structVal.NumField(); i++ {
		field := structVal.Field(i)
		fieldType := structType.Field(i)

		if !field.CanSet() {
			continue
		}

		switch field.Kind() {
		case reflect.Uint8:
			var tmp uint8
			if err := binary.Read(r, binary.BigEndian, &tmp); err != nil {
				return 0, fmt.Errorf("failed to decode uint8 field '%s': %w", fieldType.Name, err)
			}
			field.SetUint(uint64(tmp))

		case reflect.Uint16:
			var tmp uint16
			if err := binary.Read(r, binary.BigEndian, &tmp); err != nil {
				return 0, fmt.Errorf("failed to decode uint16 field '%s': %w", fieldType.Name, err)
			}
			field.SetUint(uint64(tmp))

		case reflect.Uint32:
			var tmp uint32
			if err := binary.Read(r, binary.BigEndian, &tmp); err != nil {
				return 0, fmt.Errorf("failed to decode uint32 field '%s': %w", fieldType.Name, err)
			}
			field.SetUint(uint64(tmp))

		default:
			return 0, fmt.Errorf("unsupported field kind '%s' in struct '%s'", field.Kind(), fieldType.Name)
		}
	}

	// Extract the msg_id field
	for i := 0; i < structVal.NumField(); i++ {
		fieldType := structType.Field(i)
		if tag := fieldType.Tag.Get("bmux"); tag == "msg_id" {
			return uint16(structVal.Field(i).Uint()), nil
		}
	}

	return 0, errors.New("no field tagged with `bmux:\"msg_id\"` found")
}
