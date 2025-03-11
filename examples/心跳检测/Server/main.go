package main

import (
	"fmt"

	"github.com/Cai-ki/cinx/chook"
	"github.com/Cai-ki/cinx/ciface"
	"github.com/Cai-ki/cinx/cnet"
	"github.com/Cai-ki/cinx/crouter"
)

type helloRouter struct {
	cnet.BaseRouter
}

func (h *helloRouter) Handle(request ciface.IRequest) {
	// 请求，直接回复响应
	fmt.Println("[Cinx] Received:", string(request.GetData()))
	err := request.GetConn().SendMsg(0, []byte("received"))
	if err != nil {
		fmt.Println("[Cinx] error:", err)
	}
}

func main() {
	//创建一个server句柄
	s := cnet.NewServer()
	s.AddRouter(crouter.MsgIDHeartbeatRequest, &crouter.HeartbeatPingRouter{})
	s.AddRouter(crouter.MsgIDHeartbeatResponse, &crouter.HeartbeatPongRouter{})
	s.SetOnConnStart(func(conn ciface.IConn) {
		go chook.StartHeartbeat(conn)
		go chook.StartHeartbeatChecker(conn)
	})
	s.AddRouter(0, &helloRouter{})
	//开启服务
	s.Serve()
}
