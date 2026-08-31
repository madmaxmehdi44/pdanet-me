package transport

import "io"

type PacketTransport interface {
	Reader() io.Reader
	Writer() io.Writer
}

type stream struct {
	r io.Reader
	w io.Writer
}

func NewStream(rw io.ReadWriter) PacketTransport {
	return stream{r: rw, w: rw}
}

func (s stream) Reader() io.Reader { return s.r }
func (s stream) Writer() io.Writer { return s.w }
