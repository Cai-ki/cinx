package main

import (
	"fmt"
	"time"

	"github.com/Cai-ki/cinx/ciface"
	"github.com/Cai-ki/cinx/cnet"
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
	s := cnet.NewServer()

	s.AddRouter(0, &helloRouter{})

	s.Start()

	time.Sleep(5 * time.Second)
	s.Stop()
}
