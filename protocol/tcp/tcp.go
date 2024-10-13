package tcp

import (
	"github.com/gogf/gf/v2/encoding/gbinary"
	"io"
	"math"
	"net"
	"sort"
	"time"
)

const (
	FixedLengthDataFragment = 4 + 4 + 1 + 2 + 4
)

const (
	MaxPayload = 4096
)
const (
	ReaderUnlimited = -1 // 无数次读取
	ReaderUnTimeout = -1 // 无读取时间限制
)

// DataFragment 全局协议
type DataFragment struct {
	GlobalSeq     uint32 // 全局序列号
	subSeq        uint32 // 子序列号
	isEnd         bool   // 是否结尾
	Control       uint16 // 控制报文
	PayloadLength uint32
	Payload       []byte
}

// Encode 协议序列化
func (s *DataFragment) Encode() []byte {
	return gbinary.Encode(s.GlobalSeq, s.subSeq, s.isEnd, s.Control, s.PayloadLength, s.Payload)
}

// Reader 协议读取器。当 callback 返回 false 则退出读取
func Reader(conn net.Conn, timeout time.Duration, callback func(fragment *DataFragment) bool) error {
	connReaderIns := newConnReader(conn, timeout)
	for {
		header := make([]byte, FixedLengthDataFragment)
		// 读满整个头部
		_, err := io.ReadFull(connReaderIns, header)
		if err != nil {
			return err
		}
		ins := &DataFragment{
			GlobalSeq:     0,
			subSeq:        0,
			isEnd:         false,
			Control:       0,
			PayloadLength: 0,
			Payload:       nil,
		}
		if err := gbinary.Decode(header, &ins.GlobalSeq, &ins.subSeq, &ins.isEnd,
			&ins.Control, &ins.PayloadLength); err != nil {
			return err
		}
		payloadBody := make([]byte, ins.PayloadLength)
		if _, err = io.ReadFull(connReaderIns, payloadBody); err != nil {
			return err
		}
		ins.Payload = payloadBody
		if !callback(ins) {
			break
		}
	}
	return nil
}

// SendMsg 发送数据
func SendMsg(conn net.Conn, globalSeq uint32, controlFlag uint16, payload []byte) error {
	var subSeq uint32 = 1
	cutCount := int(math.Ceil(float64(len(payload)) / float64(MaxPayload)))
	maxPayload := MaxPayload
	if maxPayload > len(payload) {
		maxPayload = len(payload)
	}
	var startOffset = 0
	for i := 0; i < cutCount; i++ {
		var isEnd = true
		payloadFrag := payload[startOffset:]
		if i != cutCount-1 {
			payloadFrag = payload[startOffset : startOffset+maxPayload]
			isEnd = false
		}
		data := &DataFragment{
			GlobalSeq:     globalSeq,
			subSeq:        subSeq,
			isEnd:         isEnd,
			Control:       controlFlag,
			PayloadLength: uint32(len(payloadFrag)),
			Payload:       payloadFrag,
		}
		subSeq++
		startOffset += maxPayload
		if _, err := conn.Write(data.Encode()); err != nil {
			return err
		}
	}
	return nil
}

// SendMsgWithRelay 发送消息并且返回一条完整的协议包数据,要求来回响应是要相同的 globalSeq
// 用于登录、权限等单次认证。一条流要保证数据要线性
func SendMsgWithRelay(conn net.Conn, globalSeq uint32,
	controlFlag uint16, payload []byte, timeout time.Duration) (*DataFragment, error) {
	if err := SendMsg(conn, globalSeq, controlFlag, payload); err != nil {
		return nil, err
	}
	var fragmentArray = make([]*DataFragment, 0)
	var endSubSeq bool
	if err := Reader(conn, timeout, func(fragment *DataFragment) bool {
		if fragment.GlobalSeq != globalSeq {
			return true
		}
		fragmentArray = append(fragmentArray, fragment)
		if fragment.isEnd {
			endSubSeq = true
		}
		if !endSubSeq {
			return true
		}
		sort.SliceStable(fragmentArray, func(i, j int) bool {
			return fragmentArray[i].subSeq < fragmentArray[j].subSeq
		})
		// 判断数据包是否顺序
		var prev uint32 = 0
		for _, v := range fragmentArray {
			if prev == 0 {
				prev = v.subSeq
				continue
			}
			// 非连续,继续读。还有数据
			if v.subSeq != prev+1 {
				return true
			}
			prev = v.subSeq
		}
		return false
	}); err != nil {
		return nil, err
	}
	lastFragment := fragmentArray[len(fragmentArray)-1]
	// 组成一个 Fragment 返回
	dataFragment := &DataFragment{
		GlobalSeq:     globalSeq,
		subSeq:        lastFragment.subSeq,
		isEnd:         true,
		Control:       lastFragment.Control,
		PayloadLength: 0,
		Payload:       make([]byte, 0),
	}
	for _, v := range fragmentArray {
		dataFragment.Payload = append(dataFragment.Payload, v.Payload...)
	}
	dataFragment.PayloadLength = uint32(len(dataFragment.Payload))
	return dataFragment, nil
}
