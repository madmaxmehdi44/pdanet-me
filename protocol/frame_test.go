package protocol

import (
    "bufio"
    "bytes"
    "testing"
)

func TestFrameRoundTrip(t *testing.T) {
    in := Frame{Type: FrameIP, Flags: 7, Payload: []byte{1, 2, 3, 4, 5}}
    var buf bytes.Buffer
    if err := Write(&buf, in); err != nil { t.Fatal(err) }
    out, err := Read(bufio.NewReader(&buf))
    if err != nil { t.Fatal(err) }
    if out.Type != in.Type || out.Flags != in.Flags || !bytes.Equal(out.Payload, in.Payload) {
        t.Fatalf("round-trip mismatch: %#v != %#v", out, in)
    }
}
