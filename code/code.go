package code

import (
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
)

type ErrorCode int

const (
	UnknownCode ErrorCode = -1 // 位置错误
)
const (
	ReadTimesBreak ErrorCode = iota + 1
	ReadTimeout
)

func (s ErrorCode) ToInt() int {
	return int(s)
}

// ToString 字符转换
func (s ErrorCode) ToString() string {
	switch s {
	case ReadTimesBreak:
		return "读取次数上限"
	case ReadTimeout:
		return "读取超时"
	}
	return ""
}

// NewCodeErrorf 使用自定义翻译覆盖
func NewCodeErrorf(code ErrorCode, format string, args ...interface{}) error {
	return gerror.NewCodeSkipf(gcode.New(code.ToInt(), code.ToString(), ""), 1, format, args...)
}

// NewCodeError 默认使用 ErrorCode 的错误码翻译
func NewCodeError(code ErrorCode) error {
	return gerror.NewCodeSkipf(gcode.New(code.ToInt(), code.ToString(), ""), 1, code.ToString())
}

// GetCodeInError Code 转换
func GetCodeInError(err error) ErrorCode {
	return ErrorCode(gerror.Code(err).Code())
}
