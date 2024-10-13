package tcp

import (
	"github.com/xhyonline/hs_protocol/code"
	"net"
	"time"
)

type connReadRes struct {
	n   int
	err error
}

type connReader struct {
	conn        net.Conn
	onceTimeout time.Duration
	readSig     chan *connReadRes
}

func (s *connReader) Read(in []byte) (n int, err error) {
	go func(in []byte) {
		n, err := s.conn.Read(in)
		ins := &connReadRes{
			n:   n,
			err: err,
		}
		s.readSig <- ins
	}(in)
	if s.onceTimeout != ReaderUnTimeout {
		select {
		case <-time.After(s.onceTimeout):
			_ = s.conn.Close()
			return 0, code.NewCodeError(code.ReadTimeout)
		case sig := <-s.readSig:
			return sig.n, sig.err
		}
	}
	sig := <-s.readSig
	return sig.n, sig.err
}

// newConnReader 实例化读取器
func newConnReader(conn net.Conn, timeout time.Duration) *connReader {
	return &connReader{
		conn:        conn,
		onceTimeout: timeout,
		readSig:     make(chan *connReadRes),
	}
}
