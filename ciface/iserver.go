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
	CallOnConnStart(conn IConn)
	CallOnConnStop(conn IConn)
}
