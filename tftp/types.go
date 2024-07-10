package tftp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"strings"
)

const (
	DatagramSize = 516              // manimum supported datagram size
	BlockSize    = DatagramSize - 4 // header size
)

type OpCode uint16

const (
	OpRRQ OpCode = iota + 1 // Read Request Code
	_
	OpData // Data code
	OpAck  // Acknolegment code
	OpErr  // Err code
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

type Data struct {
	Block   uint16    // Block number for serialization by the client
	Payload io.Reader // Data payload to be serialized
}

// Creates the request packet structure
// 2 bytes - opCode | n bytes - filename | 1 byte - 0 | n byte - mode | 1 byte - 0
//
//	WARNING:   I don't we shoud add 4 bytes for OpCode
func (q ReadReq) MarshalBinary() ([]byte, error) {
	mode := "octet"
	if q.Mode != "" {
		mode = q.Mode
	}
	// Here ----------
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

// Reads the request packet structure
// 2 bytes - opCode | n bytes - filename | 1 byte - 0 | n byte - mode | 1 byte - 0
func (q *ReadReq) UnmarshalBinary(p []byte) error {
	r := bytes.NewBuffer(p)
	var code OpCode
	err := binary.Read(r, binary.BigEndian, &code)
	if err != nil {
		return err
	}
	if code != OpRRQ {
		return errors.New("invalid RRQ")
	}
	q.Filename, err = r.ReadString(0)
	if err != nil {
		return errors.New("invalid RRQ")
	}
	q.Filename = strings.TrimRight(q.Filename, "\x00")
	if len(q.Filename) == 0 {
		return errors.New("invalid RRQ")
	}
	q.Mode, err = r.ReadString(0)
	if err != nil {
		return errors.New("invalid RRQ")
	}
	q.Mode = strings.TrimRight(q.Mode, "\x00")
	if len(q.Filename) == 0 {
		return errors.New("invalid RRQ")
	}
	actual := strings.ToLower(q.Mode)
	if actual != "octet" {
		return errors.New("only binary transfers supported")
	}
	return nil
}

// creates the data packet structur
// 2 bytes - opCode | 2 bytes - block number | n byte - payload
func (d *Data) MarshalBinary() ([]byte, error) {
	b := new(bytes.Buffer)
	b.Grow(DatagramSize)
	d.Block++

	err := binary.Write(b, binary.BigEndian, OpData)
	if err != nil {
		return nil, err
	}
	err = binary.Write(b, binary.BigEndian, d.Block)
	if err != nil {
		return nil, err
	}
	_, err = io.CopyN(b, d.Payload, BlockSize)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// reads the data packet structur
// 2 bytes - opCode | 2 bytes - block number | n byte - payload
func (d *Data) UnmarshalBinary(p []byte) error {
	if l := len(p); l < 4 || l > DatagramSize {
		return errors.New("Invalid OpData")
	}

	var code OpCode
	err := binary.Read(bytes.NewReader(p[:2]), binary.BigEndian, &code)
	if err != nil {
		return errors.New("Invalid OpData")
	}
	err = binary.Read(bytes.NewReader(p[2:4]), binary.BigEndian, &d.Block)
	if err != nil {
		return errors.New("Invalid OpData")
	}
	d.Payload = bytes.NewBuffer(p[4:])

	return nil

}

//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
//
