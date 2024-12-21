package tcp

// 控制协议报文
const (
	ControlSign             = iota + 1 // 发送登录信息的报文。账号密码不正确的情况下不允许用户连接
	ControlSignSuccess                 // 登录成功
	ControlSignError                   // 登录失败
	ControlPing                        // 探活报文
	ControlListenerConflict            // 当前 listener 下已经有相同的用户
)
