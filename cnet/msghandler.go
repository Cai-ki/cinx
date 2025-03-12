package cnet

import (
	"fmt"
	"strconv"

	"github.com/Cai-ki/cinx/ciface"
	"github.com/Cai-ki/cinx/cutils"
)

type MsgHandle struct {
	Apis           map[uint32]ciface.IRouter // 映射 MsgID 对应的处理方法
	WorkerPoolSize uint32                    // worker pool 大小
	TaskQueues     []chan ciface.IRequest    // worker 获取任务的消息队列
}

func NewMsgHandle() *MsgHandle {
	return &MsgHandle{
		Apis:           make(map[uint32]ciface.IRouter),
		WorkerPoolSize: cutils.GlobalObject.WorkerPoolSize,
		//一个 worker 对应一个 task 队列
		TaskQueues: make([]chan ciface.IRequest, cutils.GlobalObject.WorkerPoolSize),
	}
}

// 非阻塞方式处理消息
func (mh *MsgHandle) DoMsgHandler(request ciface.IRequest) {
	handler, ok := mh.Apis[request.GetMsgID()]
	if !ok {
		fmt.Println("api msgId = ", request.GetMsgID(), " is not FOUND!")
		return
	}

	//执行对应处理方法
	handler.PreHandle(request)
	handler.Handle(request)
	handler.PostHandle(request)

	request.Done()
}

// 为消息添加具体的处理逻辑
func (mh *MsgHandle) AddRouter(msgId uint32, router ciface.IRouter) {
	// 判断当前 msg 绑定的 API 处理方法是否已经存在
	if _, ok := mh.Apis[msgId]; ok {
		panic("[Cinx] Repeated api , msgId = " + strconv.Itoa(int(msgId)))
	}
	// 添加 msg 与 API 的绑定关系
	mh.Apis[msgId] = router
	fmt.Println("[Cinx] Add api msgId = ", msgId)
}

// worker 工作流程
func (mh *MsgHandle) StartOneWorker(workerID int, taskQueue chan ciface.IRequest) {
	// 监听队列中的消息
	for request := range taskQueue {
		mh.DoMsgHandler(request)
	}
	// for {
	// 	select {
	// 	case request := <-taskQueue:
	// 		mh.DoMsgHandler(request)
	// 	}
	// }
}

// 启动 worker 工作池
func (mh *MsgHandle) StartWorkerPool() {
	for i := 0; i < int(mh.WorkerPoolSize); i++ {
		// 初始化当前 worker 消息队列管道
		mh.TaskQueues[i] = make(chan ciface.IRequest, cutils.GlobalObject.MaxWorkerTaskLen)

		fmt.Println("[Cinx] Worker ID = ", i, " is started.")
		// 创建一个 worker 协程
		go mh.StartOneWorker(i, mh.TaskQueues[i])
	}
}

// 分发消息给消息队列处理
func (mh *MsgHandle) SendMsgToTaskQueue(request ciface.IRequest) {
	// 朴素的任务分配策略
	workerID := request.GetConn().GetConnID() % mh.WorkerPoolSize
	fmt.Println("[Cinx] Add ConnID=", request.GetConn().GetConnID(), " request msgID=", request.GetMsgID(), "to workerID=", workerID)

	// 将消息发送给对应的 worker 的消息队列
	mh.TaskQueues[workerID] <- request
}
