package parsing

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"reflect"
	"time"

	"github.com/etwodev/bmux/config"
	"google.golang.org/protobuf/proto"
)

func ParseEnvelope(conn net.Conn) (*PacketEnvelope, error) {
	if timeout := config.ReadTimeout(); timeout > 0 {
		_ = conn.SetReadDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
	}

	header := make([]byte, 3)
	if _, err := io.ReadFull(conn, header); err != nil {
		return nil, fmt.Errorf("failed to read header fields: %w", err)
	}

	headLen := header[0]
	bodyLen := binary.LittleEndian.Uint16(header[1:3])

	rawHead := make([]byte, headLen)
	if timeout := config.ReadTimeout(); timeout > 0 {
		_ = conn.SetReadDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
	}
	if _, err := io.ReadFull(conn, rawHead); err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	rawBody := make([]byte, bodyLen)
	if timeout := config.ReadTimeout(); timeout > 0 {
		_ = conn.SetReadDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
	}
	if _, err := io.ReadFull(conn, rawBody); err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	return &PacketEnvelope{
		HeadLen: headLen,
		BodyLen: bodyLen,
		RawHead: rawHead,
		RawBody: rawBody,
	}, nil
}

// ParseHeader parses raw header bytes into a user-defined struct,
// using protobuf unmarshalling if applicable.
// For protobuf messages, it extracts the msgID from the field named "Msgid".
// For manual structs, it uses bmux tags, including bmux:"msg_id".
func ParseHeader(rawHead []byte, headerPtr any) (msgID int32, err error) {
	v := reflect.ValueOf(headerPtr)
	if v.Kind() != reflect.Pointer || v.Elem().Kind() != reflect.Struct {
		return 0, errors.New("headerPtr must be a pointer to a struct")
	}

	// Check if protobuf message
	if pm, ok := headerPtr.(proto.Message); ok {
		// Unmarshal protobuf
		if err := proto.Unmarshal(rawHead, pm); err != nil {
			return 0, fmt.Errorf("failed to unmarshal protobuf header: %w", err)
		}

		// Extract msgID by field name "Msgid"
		structVal := v.Elem()
		structType := structVal.Type()

		for i := 0; i < structVal.NumField(); i++ {
			fieldType := structType.Field(i)
			if fieldType.Name == "Msgid" {
				field := structVal.Field(i)
				if !field.IsValid() {
					return 0, errors.New("msg_id field is invalid")
				}
				switch field.Kind() {
				case reflect.Int32, reflect.Int:
					return int32(field.Int()), nil
				case reflect.Uint32, reflect.Uint, reflect.Uint64:
					return int32(field.Uint()), nil
				default:
					return 0, fmt.Errorf("unsupported msg_id field kind %s", field.Kind())
				}
			}
		}
		return 0, errors.New("no field named 'Msgid' found in protobuf header")
	}

	// Manual binary decoding fallback
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

		case reflect.Int8:
			var tmp int8
			if err := binary.Read(r, binary.BigEndian, &tmp); err != nil {
				return 0, fmt.Errorf("failed to decode int8 field '%s': %w", fieldType.Name, err)
			}
			field.SetInt(int64(tmp))

		case reflect.Int16:
			var tmp int16
			if err := binary.Read(r, binary.BigEndian, &tmp); err != nil {
				return 0, fmt.Errorf("failed to decode int16 field '%s': %w", fieldType.Name, err)
			}
			field.SetInt(int64(tmp))

		case reflect.Int32:
			var tmp int32
			if err := binary.Read(r, binary.BigEndian, &tmp); err != nil {
				return 0, fmt.Errorf("failed to decode int32 field '%s': %w", fieldType.Name, err)
			}
			field.SetInt(int64(tmp))

		default:
			return 0, fmt.Errorf("unsupported field kind '%s' in struct '%s'", field.Kind(), fieldType.Name)
		}
	}

	// Extract msg_id by bmux tag "msg_id"
	for i := 0; i < structVal.NumField(); i++ {
		fieldType := structType.Field(i)
		if tag := fieldType.Tag.Get("bmux"); tag == "msg_id" {
			field := structVal.Field(i)
			if !field.IsValid() {
				return 0, errors.New("msg_id field is invalid")
			}
			switch field.Kind() {
			case reflect.Int32, reflect.Int:
				return int32(field.Int()), nil
			case reflect.Uint32, reflect.Uint, reflect.Uint64:
				return int32(field.Uint()), nil
			default:
				return 0, fmt.Errorf("unsupported msg_id field kind %s", field.Kind())
			}
		}
	}

	return 0, errors.New("no field tagged with `bmux:\"msg_id\"` found")
}
