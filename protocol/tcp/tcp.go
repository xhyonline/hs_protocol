package tcp

import (
	"github.com/gogf/gf/v2/encoding/gbinary"
	"io"
	"net"
)

const (
	FixedLengthDataFragment = 4 + 4 + 1 + 2 + 4
)

const (
	BufferSize = 4096
)

// DataFragment 全局协议
type DataFragment struct {
	GlobalSeq     uint32 // 全局序列号
	SubSeq        uint32 // 子序列号
	IsEnd         bool   // 是否结尾
	Control       uint16 // 控制报文
	PayloadLength uint32
	Payload       []byte
}

// Encode 协议序列化
func (s *DataFragment) Encode() []byte {
	return gbinary.Encode(s.GlobalSeq, s.SubSeq, s.IsEnd, s.Control, s.PayloadLength, s.Payload)
}

// 协议读取器
func Reader(conn net.Conn, callback func(fragment *DataFragment)) error {
	for {
		header := make([]byte, FixedLengthDataFragment)
		// 读满整个头部
		_, err := io.ReadFull(conn, header)
		if err != nil {
			return err
		}
		ins := &DataFragment{
			GlobalSeq:     0,
			SubSeq:        0,
			IsEnd:         false,
			Control:       0,
			PayloadLength: 0,
			Payload:       nil,
		}
		if err := gbinary.Decode(header, &ins.GlobalSeq, &ins.SubSeq, &ins.IsEnd,
			&ins.Control, &ins.PayloadLength); err != nil {
			return err
		}
		payloadBody := make([]byte, ins.PayloadLength)
		if _, err = io.ReadFull(conn, payloadBody); err != nil {
			return err
		}
		ins.Payload = payloadBody
		callback(ins)
	}
}
