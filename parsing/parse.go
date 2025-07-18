package parsing

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"reflect"
	"strings"
	"time"

	"github.com/etwodev/bmux/config"
	"google.golang.org/protobuf/proto"
)

// ParseEnvelope extracts the raw header and body using the first three bytes.
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

// ParseHeader parses raw header bytes into a user-defined protobuf struct.
// It unmarshals the header and extracts the Msgid field (case-insensitive, supports underscores and dashes).
func ParseHeader(rawHead []byte, headerPtr any) (int32, error) {
	v := reflect.ValueOf(headerPtr)
	if v.Kind() != reflect.Pointer || v.Elem().Kind() != reflect.Struct {
		return 0, errors.New("headerPtr must be a pointer to a struct")
	}

	pm, ok := headerPtr.(proto.Message)
	if !ok {
		return 0, errors.New("headerPtr must implement proto.Message")
	}

	if err := proto.Unmarshal(rawHead, pm); err != nil {
		return 0, fmt.Errorf("failed to unmarshal protobuf header: %w", err)
	}

	structVal := v.Elem()
	structType := structVal.Type()

	isMsgIDField := func(name string) bool {
		normalized := strings.ToLower(name)
		normalized = strings.ReplaceAll(normalized, "_", "")
		normalized = strings.ReplaceAll(normalized, "-", "")
		return normalized == "msgid"
	}

	for i := 0; i < structVal.NumField(); i++ {
		fieldType := structType.Field(i)
		if isMsgIDField(fieldType.Name) {
			field := structVal.Field(i)
			if !field.IsValid() {
				return 0, errors.New("field Msgid is invalid")
			}
			switch field.Kind() {
			case reflect.Int32, reflect.Int:
				return int32(field.Int()), nil
			case reflect.Uint32, reflect.Uint, reflect.Uint64:
				return int32(field.Uint()), nil
			default:
				return 0, fmt.Errorf("unsupported Msgid field kind %s", field.Kind())
			}
		}
	}

	return 0, errors.New("no recognizable 'Msgid' field found in protobuf header")
}

// WritePacket marshals and writes a packet with the given header and body to the provided connection.
// The packet layout is: [headLen:1][bodyLen:2][headBytes][bodyBytes].
// It injects the given msgID into the header's Msgid field if available.
// It enforces max header (255 bytes) and body (65535 bytes) size constraints.
func WritePacket(ctx context.Context, conn net.Conn, header proto.Message, body proto.Message) error {
	if timeout := config.WriteTimeout(); timeout > 0 {
		_ = conn.SetWriteDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
	}

	var (
		headBytes []byte
		bodyBytes []byte
		err       error
	)

	if header != nil {
		headBytes, err = proto.Marshal(header)
		if err != nil {
			return fmt.Errorf("marshal header: %w", err)
		}
		if len(headBytes) > 255 {
			return fmt.Errorf("header too large to encode (max 255 bytes), got %d", len(headBytes))
		}
	}

	if body != nil {
		bodyBytes, err = proto.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal body: %w", err)
		}
		if len(bodyBytes) > 65535 {
			return fmt.Errorf("body too large to encode (max 65535 bytes), got %d", len(bodyBytes))
		}
	}

	packet := make([]byte, 3+len(headBytes)+len(bodyBytes))
	packet[0] = byte(len(headBytes))
	binary.LittleEndian.PutUint16(packet[1:3], uint16(len(bodyBytes)))
	copy(packet[3:], headBytes)
	copy(packet[3+len(headBytes):], bodyBytes)

	_, err = conn.Write(packet)
	if err != nil {
		return fmt.Errorf("write packet: %w", err)
	}
	return nil
}
