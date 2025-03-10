package cnet

import (
	"fmt"
	"net"

	"github.com/Cai-ki/cinx/ciface"
	"github.com/Cai-ki/cinx/cutils"
)

// iServer 接口实现，定义一个Server服务类
type Server struct {
	//服务器的名称
	Name string
	//tcp4 or other
	TcpVersion string
	//服务绑定的 IP 地址
	IP string
	//服务绑定的端口
	Port int
	//当前 server 的消息管理模块，用来绑定 MsgId 和对应的处理方法
	msgHandler ciface.IMsgHandle
	//当前 server 的链接管理器
	ConnMgr ciface.IConnManager

	OnConnStart func(conn ciface.IConnection)
	OnConnStop  func(conn ciface.IConnection)
}

// 开启 server 服务（无阻塞）
func (s *Server) Start() {
	// 输出 server 信息
	fmt.Println("[Cinx] Server Name:", s.Name, "listenner at IP:", s.IP, " Port:", s.Port)
	fmt.Printf("[Cinx] Version: %s, MaxConn: %d,  MaxPacketSize: %d\n",
		cutils.GlobalObject.Version,
		cutils.GlobalObject.MaxConn,
		cutils.GlobalObject.MaxPacketSize)

	// 创建协程不间断处理链接
	go func() {
		//0 开启工作池
		s.msgHandler.StartWorkerPool()

		//1 封装 tcp 地址
		addr, err := net.ResolveTCPAddr(s.TcpVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			fmt.Println("[Cinx] resolve tcp address err: ", err)
			return
		}

		//2 创建监听 socket
		listenner, err := net.ListenTCP(s.TcpVersion, addr)
		if err != nil {
			fmt.Println("[Cinx] listen", s.TcpVersion, "err", err)
			return
		}

		//输出监听成功信息
		fmt.Println("[Cinx] start success, now listenning...")

		// 简单实现一个自增的连接 ID
		var cid uint32 = 0

		//3 持续监听客户端连接
		for {
			//3.1 阻塞等待客户端建立连接请求
			conn, err := listenner.AcceptTCP()
			if err != nil {
				fmt.Println("[Cinx] Accept err ", err)
				continue
			}

			//3.2 判断当前服务器的连接数是否已经超过最大连接数
			if s.ConnMgr.Len() >= cutils.GlobalObject.MaxConn {
				conn.Close()
				continue
			}

			//3.3 初始化连接模块
			dealConn := NewConntion(s, conn, cid, s.msgHandler)
			cid++

			//3.4 启动协程处理当前连接的业务
			go dealConn.Start()
		}
	}()
}

// 停止 server 服务
func (s *Server) Stop() {
	fmt.Println("[Cinx] stop server , name ", s.Name)

	// 通过 ConnManager 清除并停止所有连接
	s.ConnMgr.ClearConn()
}

// 开启 server 服务（阻塞）
func (s *Server) Serve() {
	s.Start()

	//TODO Server.Serve() 是否在启动服务的时候 还要处理其他的事情呢 可以在这里添加

	// 阻塞
	select {}
}

// 为特定消息注册处理函数
func (s *Server) AddRouter(msgId uint32, router ciface.IRouter) {
	s.msgHandler.AddRouter(msgId, router)

	fmt.Println("[Cinx] Add Router success! ")
}

// 得到链接管理
func (s *Server) GetConnMgr() ciface.IConnManager {
	return s.ConnMgr
}

// 设置 server 的连接创建时 hook 函数
func (s *Server) SetOnConnStart(hookFunc func(ciface.IConnection)) {
	s.OnConnStart = hookFunc
}

// 设置 server 的连接断开时 hook 函数
func (s *Server) SetOnConnStop(hookFunc func(ciface.IConnection)) {
	s.OnConnStop = hookFunc
}

// 调用 hook 函数
func (s *Server) CallOnConnStart(conn ciface.IConnection) {
	if s.OnConnStart != nil {
		fmt.Println("[Cinx] CallOnConnStart....")
		s.OnConnStart(conn)
	}
}

// 调用 hook 函数
func (s *Server) CallOnConnStop(conn ciface.IConnection) {
	if s.OnConnStop != nil {
		fmt.Println("[Cinx] CallOnConnStop....")
		s.OnConnStop(conn)
	}
}

// 创建 server 实例
func NewServer() ciface.IServer {
	//初始化全局配置文件
	cutils.GlobalObject.Reload()

	s := &Server{
		Name:       cutils.GlobalObject.Name,
		TcpVersion: "tcp4",
		IP:         cutils.GlobalObject.Host,
		Port:       cutils.GlobalObject.TcpPort,
		msgHandler: NewMsgHandle(),
		ConnMgr:    NewConnManager(),
	}
	return s
}

func (s *Server) GetMsgHandler() ciface.IMsgHandle {
	return s.msgHandler
}
