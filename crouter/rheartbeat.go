package crouter

import (
	"fmt"
	"time"

	"github.com/Cai-ki/cinx/ciface"
	"github.com/Cai-ki/cinx/cnet"
)

const (
	MsgIDHeartbeatRequest  = 1001 // 心跳请求
	MsgIDHeartbeatResponse = 1002 // 心跳响应
)

type HeartbeatPingRouter struct {
	cnet.BaseRouter
}

func (h *HeartbeatPingRouter) Handle(request ciface.IRequest) {
	// 收到心跳请求，直接回复响应
	fmt.Println("[Cinx] Received heartbeat request, sending pong...")
	err := request.GetConnection().SendMsg(MsgIDHeartbeatResponse, []byte("pong"))
	if err != nil {
		fmt.Println("[Cinx] Send heartbeat response error:", err)
	}
}

type HeartbeatPongRouter struct {
	cnet.BaseRouter
}

func (h *HeartbeatPongRouter) Handle(request ciface.IRequest) {
	// 收到心跳请求，直接回复响应
	request.GetConnection().SetProperty("lastActiveTime", time.Now())
	fmt.Println("[Cinx] Received heartbeat response, updating last active time...")
}
