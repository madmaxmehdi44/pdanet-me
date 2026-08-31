package protocol

import (
    "bufio"
    "encoding/binary"
    "errors"
    "fmt"
    "io"
)

const (
    Magic   uint32 = 0x50444F50 // PDOP
    Version byte   = 1
    HeaderSize       = 12
    MaxPayload       = 64 * 1024

    FrameIP   byte = 1
    FramePing byte = 2
    FramePong byte = 3
)

type Frame struct {
    Type    byte
    Flags   byte
    Payload []byte
}

func Write(w io.Writer, f Frame) error {
    if len(f.Payload) > MaxPayload {
        return fmt.Errorf("payload too large: %d", len(f.Payload))
    }
    h := make([]byte, HeaderSize)
    binary.LittleEndian.PutUint32(h[0:4], Magic)
    h[4] = Version
    h[5] = f.Type
    h[6] = f.Flags
    h[7] = 0
    binary.LittleEndian.PutUint32(h[8:12], uint32(len(f.Payload)))
    if _, err := w.Write(h); err != nil { return err }
    if len(f.Payload) > 0 {
        _, err := w.Write(f.Payload)
        return err
    }
    return nil
}

func Read(r *bufio.Reader) (Frame, error) {
    var f Frame
    h := make([]byte, HeaderSize)
    if _, err := io.ReadFull(r, h); err != nil { return f, err }
    if binary.LittleEndian.Uint32(h[0:4]) != Magic { return f, errors.New("invalid magic") }
    if h[4] != Version { return f, fmt.Errorf("unsupported version: %d", h[4]) }
    n := binary.LittleEndian.Uint32(h[8:12])
    if n > MaxPayload { return f, fmt.Errorf("payload too large: %d", n) }
    f.Type, f.Flags = h[5], h[6]
    if n == 0 { return f, nil }
    f.Payload = make([]byte, n)
    if _, err := io.ReadFull(r, f.Payload); err != nil { return Frame{}, err }
    return f, nil
}
