package ciface

type IServer interface {
	Start()
	Stop()
	Serve()

	AddRouter(msgId uint32, router IRouter)

	GetConnMgr() IConnManager
	GetMsgHandler() IMsgHandle

	SetOnConnStart(func(IConn))
	SetOnConnStop(func(IConn))
	GetOnConnStart() func(IConn)
	GetOnConnStop() func(IConn)
}
