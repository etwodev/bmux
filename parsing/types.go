package parsing

type PacketEnvelope struct {
	HeadLen uint8
	BodyLen uint16
	RawHead []byte
	RawBody []byte
}
