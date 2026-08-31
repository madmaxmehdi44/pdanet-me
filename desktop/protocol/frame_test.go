package protocol

import (
	"bufio"
	"bytes"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	want := Frame{Type: FrameIP, Payload: []byte{1, 2, 3, 4}}
	if err := Write(&buf, want); err != nil {
		t.Fatal(err)
	}
	got, err := Read(bufio.NewReader(&buf))
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != want.Type || !bytes.Equal(got.Payload, want.Payload) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestRejectsOversize(t *testing.T) {
	var buf bytes.Buffer
	payload := make([]byte, MaxPayloadSize+1)
	if err := Write(&buf, Frame{Type: FrameIP, Payload: payload}); err == nil {
		t.Fatal("expected error")
	}
}
