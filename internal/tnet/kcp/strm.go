package kcp

import (
	"github.com/golang/snappy"
	"github.com/xtaci/smux"
)

type Strm struct {
	*smux.Stream
	r *snappy.Reader
	w *snappy.Writer
}

func NewStrm(stream *smux.Stream) *Strm {
	return &Strm{
		Stream: stream,
		r:      snappy.NewReader(stream),
		w:      snappy.NewBufferedWriter(stream),
	}
}

func (s *Strm) Read(p []byte) (n int, err error) {
	return s.r.Read(p)
}

func (s *Strm) Write(p []byte) (n int, err error) {
	n, err = s.w.Write(p)
	if err != nil {
		return n, err
	}
	return n, s.w.Flush()
}

func (s *Strm) Close() error {
	// Flush any remaining data in the snappy buffer
	if s.w != nil {
		s.w.Flush()
	}
	// Close the underlying stream
	return s.Stream.Close()
}

func (s *Strm) SID() int {
	return int(s.ID())
}
