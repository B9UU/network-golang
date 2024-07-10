package tftp

import (
	"bytes"
	"encoding/binary"
)

const (
	DatagramSize = 516              // manimum supported datagram size
	BlockSize    = DatagramSize - 4 // header size
)

type OpCode uint16

const (
	OpRRQ OpCode = iota + 1
	_
	OpData
	OpAck
	OpErr
)

type ErrCode uint16

const (
	ErrUnknown ErrCode = iota
	ErrNotFound
	ErrAccessViolation
	ErrDiskFull
	ErrIllegalOp
	ErrUnknownId
	ErrFileExists
	ErrNoUser
)

type ReadReq struct {
	Filename string
	Mode     string
}

// Reads the request packet structure
// 2 bytes - opCode | n bytes - filename | 1 byte - 0 | n byte - mode | 1 byte - 0
func (q ReadReq) MarshalBinary() ([]byte, error) {
	mode := "octet"
	if q.Mode != "" {
		mode = q.Mode
	}
	cap := 2 + 2 + len(q.Filename) + 1 + len(q.Mode) + 1
	b := new(bytes.Buffer)
	b.Grow(cap)
	err := binary.Write(b, binary.BigEndian, OpRRQ)
	if err != nil {
		return nil, err
	}
	_, err = b.WriteString(q.Filename)
	if err != nil {
		return nil, err
	}
	err = b.WriteByte(0)
	if err != nil {
		return nil, err
	}
	_, err = b.WriteString(mode)
	if err != nil {
		return nil, err
	}
	err = b.WriteByte(0)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
