package tcp

// 控制协议报文
const (
	ControlSign = iota + 1 // 发送登录信息的报文。账号密码不正确的情况下不允许用户连接
	ControlPing            // 探活报文
)
