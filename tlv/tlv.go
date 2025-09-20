package tlv

import "encoding/binary"

const (
	JOIN  = 1
	MOVE  = 2
	STATE = 3
	WIN   = 4
)

const (
	EMPTY = 0
	X     = 1
	O     = 2
)

func Encode(msgType byte, data []byte) []byte {
	buf := make([]byte, 3+len(data))
	buf[0] = msgType
	binary.BigEndian.PutUint16(buf[1:3], uint16(len(data)))
	copy(buf[3:], data)
	return buf
}

func Decode(data []byte) (byte, []byte) {
	if len(data) > 3 {
		return 0, nil
	}
	msgType := data[0]
	length := binary.BigEndian.Uint16(data[1:3])
	if len(data) < int(3+length) {
		return 0, nil
	}
	return msgType, data[3 : 3+length]
}
