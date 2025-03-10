package main

import (
	"fmt"
	"io"
	"net"
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
	err := request.GetConnection().SendMsg(0, []byte("received"))
	if err != nil {
		fmt.Println("[Cinx] error:", err)
	}
}

func main() {

	fmt.Println("Client Test ... start")
	//3秒之后发起测试请求，给服务端开启服务的机会
	time.Sleep(1 * time.Second)

	conn, err := net.Dial("tcp4", "127.0.0.1:7777")
	if err != nil {
		fmt.Println("client start err, exit!")
		return
	}

	// client := cnet.NewServer()

	// client.AddRouter(crouter.MsgIDHeartbeatRequest, &crouter.HeartbeatPingRouter{})
	// client.AddRouter(crouter.MsgIDHeartbeatResponse, &crouter.HeartbeatPongRouter{})
	// client.AddRouter(0, &helloRouter{})

	// C := cnet.NewConntion(client, conn.(*net.TCPConn), 1, client.GetMsgHandler())

	// client.GetConnMgr().Add(C)

	// C.SendBuffMsg(0, []byte("hello"))

	// go C.StartReader()
	// go C.StartWriter()

	// for {
	// 	C.SendMsg(0, []byte("nihao"))
	// 	time.Sleep(1 * time.Second)
	// }
	//发封包message消息
	dp := cnet.NewDataPack()
	call := func(id uint32, data string) {
		fmt.Println("==> Call Msg: ID=", id, ", data=", data)

		msg, _ := dp.Pack(cnet.NewMsgPackage(id, []byte(data)))
		_, err := conn.Write(msg)
		if err != nil {
			fmt.Println("write error err ", err)
			return
		}

	}

	get := func() (msg *cnet.Message) {
		msg = &cnet.Message{}
		msg.Id = 404
		//先读出流中的head部分
		headData := make([]byte, dp.GetHeadLen())
		_, err = io.ReadFull(conn, headData) //ReadFull 会把msg填充满为止
		if err != nil {
			fmt.Println("read head error")
			return
		}
		//将headData字节流 拆包到msg中
		msgHead, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("server unpack err:", err)
			return
		}

		if msgHead.GetDataLen() > 0 {
			//msg 是有data数据的，需要再次读取data数据
			msg = msgHead.(*cnet.Message)
			msg.Data = make([]byte, msg.GetDataLen())

			//根据dataLen从io中读取字节流
			_, err := io.ReadFull(conn, msg.Data)
			if err != nil {
				fmt.Println("server unpack data err:", err)
				return
			}

			fmt.Println("==> Recv Msg: ID=", msg.Id, ", len=", msg.DataLen, ", data=", string(msg.Data))
		}

		return
	}

	call(0, "hello")

	for {
		msg := get()
		if msg.Id == 0 {
			call(0, "get: "+string(msg.Data))
		} else if msg.Id == 1001 {
			// call(1002, "get: "+string(msg.Data))
		} else if msg.Id == 1002 {
			call(1001, "get: "+string(msg.Data))
		} else {
			fmt.Println("error")
			break
		}
		time.Sleep(1 * time.Second)
	}
}
