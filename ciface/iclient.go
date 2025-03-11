package ciface

// 客户端接口
type IClient interface {
	Restart()
	Start()
	Stop()
	AddRouter(msgId uint32, router IRouter)
	Conn() IConn
	GetMsgHandler() IMsgHandle

	SetOnConnStart(func(IConn))
	SetOnConnStop(func(IConn))
	GetOnConnStart() func(IConn)
	GetOnConnStop() func(IConn)
}
