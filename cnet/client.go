package cnet

import (
	"fmt"
	"net"

	"github.com/Cai-ki/cinx/ciface"
)

type Client struct {
	Name      string
	IPVersion string
	IP        string
	Port      int
	conn      ciface.IConn
	// connMux     sync.Mutex
	onConnStart func(conn ciface.IConn)
	onConnStop  func(conn ciface.IConn)
	msgHandler  ciface.IMsgHandle

	exitChan  chan struct{}
	readyChan chan struct{}
}

var _ ciface.IClient = (*Client)(nil)

func NewClient(name, ipVersion, ip string, port int) ciface.IClient {
	c := &Client{
		Name:      name,
		IPVersion: ipVersion,
		IP:        ip,
		Port:      port,

		msgHandler:  NewMsgHandle(),
		onConnStart: func(conn ciface.IConn) {},
		onConnStop:  func(conn ciface.IConn) {},
		readyChan:   make(chan struct{}),
	}
	return c
}

func (c *Client) Restart() {
	c.Stop()
	c.Start()
}

func (c *Client) Start() {
	// c.exitChan = make(chan struct{})
	c.msgHandler.StartWorkerPool()

	addr, err := net.ResolveTCPAddr(c.IPVersion, fmt.Sprintf("%s:%d", c.IP, c.Port))
	if err != nil {
		fmt.Println("[Cinx] resolve tcp address err: ", err)
		return
	}

	conn, err := net.DialTCP(c.IPVersion, nil, addr)
	if err != nil {
		fmt.Println("[Cinx] dial tcp err: ", err)
		// c.Stop()
		return
	}

	c.conn = NewClientConn(c, conn)

	go c.conn.Start()
}
func (c *Client) Stop() {
	con := c.Conn()
	con.Stop()
	c.msgHandler.Stop()
}
func (c *Client) Conn() ciface.IConn {
	return c.conn
}
func (c *Client) AddRouter(msgId uint32, router ciface.IRouter) {
	c.msgHandler.AddRouter(msgId, router)
}
func (c *Client) GetMsgHandler() ciface.IMsgHandle {
	return c.msgHandler
}
func (c *Client) SetOnConnStart(f func(ciface.IConn)) {
	c.onConnStart = f
}
func (c *Client) SetOnConnStop(f func(ciface.IConn)) {
	c.onConnStop = f
}
func (c *Client) GetOnConnStart() func(ciface.IConn) {
	return c.onConnStart
}
func (c *Client) GetOnConnStop() func(ciface.IConn) {
	return c.onConnStop
}
