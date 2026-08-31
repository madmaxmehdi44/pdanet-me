package protocol

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	Magic          uint32 = 0x50444F50 // "PDOP"
	Version        byte   = 1
	HeaderSize            = 12
	MaxPayloadSize        = 64 * 1024
)

var ErrInvalidFrame = errors.New("invalid frame")

type FrameType byte

const (
	FrameIP   FrameType = 1
	FramePing FrameType = 2
	FramePong FrameType = 3
)

type Frame struct {
	Type    FrameType
	Payload []byte
}

func Write(w io.Writer, f Frame) error {
	if f.Type == 0 {
		return fmt.Errorf("%w: missing frame type", ErrInvalidFrame)
	}
	if len(f.Payload) > MaxPayloadSize {
		return fmt.Errorf("payload too large: %d", len(f.Payload))
	}

	header := make([]byte, HeaderSize)
	binary.LittleEndian.PutUint32(header[0:4], Magic)
	header[4] = Version
	header[5] = byte(f.Type)
	header[6] = 0 // flags
	header[7] = 0 // reserved
	binary.LittleEndian.PutUint32(header[8:12], uint32(len(f.Payload)))

	if _, err := io.CopyN(writerOnly{w}, bytesReader(header), int64(len(header))); err != nil {
		return err
	}
	if len(f.Payload) == 0 {
		return nil
	}
	_, err := io.CopyN(writerOnly{w}, bytesReader(f.Payload), int64(len(f.Payload)))
	return err
}

type writerOnly struct{ io.Writer }
type bytesReader []byte

func (b bytesReader) Read(p []byte) (int, error) {
	if len(b) == 0 {
		return 0, io.EOF
	}
	n := copy(p, b)
	return n, nil
}

func Read(r *bufio.Reader) (Frame, error) {
	var f Frame
	header := make([]byte, HeaderSize)
	if _, err := io.ReadFull(r, header); err != nil {
		return f, err
	}
	if binary.LittleEndian.Uint32(header[0:4]) != Magic {
		return f, fmt.Errorf("%w: bad magic", ErrInvalidFrame)
	}
	if header[4] != Version {
		return f, fmt.Errorf("%w: unsupported version %d", ErrInvalidFrame, header[4])
	}
	if header[6] != 0 || header[7] != 0 {
		return f, fmt.Errorf("%w: unsupported flags/reserved bits", ErrInvalidFrame)
	}
	length := binary.LittleEndian.Uint32(header[8:12])
	if length > MaxPayloadSize {
		return f, fmt.Errorf("%w: payload length %d", ErrInvalidFrame, length)
	}
	payload := make([]byte, int(length))
	if _, err := io.ReadFull(r, payload); err != nil {
		return f, err
	}
	f.Type = FrameType(header[5])
	f.Payload = payload
	return f, nil
}
