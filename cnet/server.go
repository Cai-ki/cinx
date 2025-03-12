package cnet

import (
	"fmt"
	"net"

	"github.com/Cai-ki/cinx/ciface"
	"github.com/Cai-ki/cinx/cutils"
)

type Server struct {
	Name      string
	IPVersion string
	IP        string
	Port      int

	msgHandler ciface.IMsgHandle
	ConnMgr    ciface.IConnManager

	onConnStart func(conn ciface.IConn)
	onConnStop  func(conn ciface.IConn)
}

func NewServer() ciface.IServer {

	cutils.GlobalObject.Reload()

	s := &Server{
		Name:        cutils.GlobalObject.Name,
		IPVersion:   "tcp4",
		IP:          cutils.GlobalObject.Host,
		Port:        cutils.GlobalObject.TcpPort,
		msgHandler:  NewMsgHandle(),
		ConnMgr:     NewConnManager(),
		onConnStart: func(conn ciface.IConn) {},
		onConnStop:  func(conn ciface.IConn) {},
	}
	return s
}

func (s *Server) Start() {
	fmt.Println("[Cinx] Server Name:", s.Name, "listenner at IP:", s.IP, " Port:", s.Port)
	fmt.Printf("[Cinx] Version: %s, MaxConn: %d,  MaxPacketSize: %d\n",
		cutils.GlobalObject.Version,
		cutils.GlobalObject.MaxConn,
		cutils.GlobalObject.MaxPacketSize)

	s.msgHandler.StartWorkerPool()

	ready := make(chan struct{})

	go func() {
		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			fmt.Println("[Cinx] resolve tcp address err: ", err)
			return
		}

		listenner, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			fmt.Println("[Cinx] listen", s.IPVersion, "err", err)
			return
		}

		fmt.Println("[Cinx] Listenning...")

		ready <- struct{}{}

		var cid uint32 = 0
		for {
			conn, err := listenner.AcceptTCP()
			if err != nil {
				fmt.Println("[Cinx] Accept err ", err)
				continue
			}

			if s.ConnMgr.Len() >= cutils.GlobalObject.MaxConn {
				conn.Close()
				continue
			}

			dealConn := NewConntion(s, conn, cid, s.msgHandler)
			cid++

			go dealConn.Start()
		}
	}()

	<-ready
}

func (s *Server) Stop() {
	fmt.Println("[Cinx] stop server , name ", s.Name)

	s.ConnMgr.ClearConn()
}

func (s *Server) Serve() {
	s.Start()

	select {}
}

func (s *Server) AddRouter(msgId uint32, router ciface.IRouter) {
	s.msgHandler.AddRouter(msgId, router)
}

func (s *Server) GetConnMgr() ciface.IConnManager {
	return s.ConnMgr
}

func (s *Server) SetOnConnStart(hookFunc func(ciface.IConn)) {
	s.onConnStart = hookFunc
}

func (s *Server) SetOnConnStop(hookFunc func(ciface.IConn)) {
	s.onConnStop = hookFunc
}

func (s *Server) GetOnConnStart() func(ciface.IConn) {
	return s.onConnStart
}
func (s *Server) GetOnConnStop() func(ciface.IConn) {
	return s.onConnStop
}

func (s *Server) GetMsgHandler() ciface.IMsgHandle {
	return s.msgHandler
}
