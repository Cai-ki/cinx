package ciface

type IMsgHandle interface {
	StartWorkerPool()
	AddRouter(msgId uint32, router IRouter)
	DoMsgHandler(request IRequest)
	SendMsgToTaskQueue(request IRequest)
	Stop()
}
