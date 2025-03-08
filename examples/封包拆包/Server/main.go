package main

import (
	"fmt"

	"github.com/Cai-ki/cinx/ciface"
	"github.com/Cai-ki/cinx/cnet"
)

// ping test 自定义路由
type PingRouter struct {
	cnet.BaseRouter
}

// Test Handle
func (this *PingRouter) Handle(request ciface.IRequest) {
	fmt.Println("Call PingRouter Handle")
	//先读取客户端的数据，再回写ping...ping...ping
	fmt.Println("recv from client : msgId=", request.GetMsgID(), ", data=", string(request.GetData()))

	//回写数据
	err := request.GetConnection().SendMsg(1, []byte("ping...ping...ping"))
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	//创建一个server句柄
	s := cnet.NewServer()

	//配置路由
	s.AddRouter(&PingRouter{})

	//开启服务
	s.Serve()
}
