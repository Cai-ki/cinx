package cnet

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"

	"github.com/Cai-ki/cinx/ciface"
	"github.com/Cai-ki/cinx/cutils"
)

type Connection struct {
	TcpServer ciface.IServer

	Conn *net.TCPConn

	ConnID uint32

	IsClosed atomic.Bool

	IsAborted atomic.Bool

	MsgHandler ciface.IMsgHandle

	ExitBuffChan chan struct{}

	msgChan     chan []byte
	msgBuffChan chan []byte

	property     map[string]interface{}
	propertyLock sync.RWMutex

	onConnStart func(conn ciface.IConn)
	onConnStop  func(conn ciface.IConn)
}

func NewConntion(server ciface.IServer, conn *net.TCPConn, connID uint32, msgHandler ciface.IMsgHandle) *Connection {
	c := &Connection{
		TcpServer:    server,
		Conn:         conn,
		ConnID:       connID,
		MsgHandler:   msgHandler,
		ExitBuffChan: make(chan struct{}, 1),
		msgChan:      make(chan []byte),
		msgBuffChan:  make(chan []byte, cutils.GlobalObject.MaxMsgChanLen),
		property:     make(map[string]interface{}),
		onConnStart:  server.GetOnConnStart(),
		onConnStop:   server.GetOnConnStop(),
	}
	c.IsClosed.Store(false)
	c.IsAborted.Store(false)

	server.GetConnMgr().Add(c)
	return c
}

func NewClientConn(client ciface.IClient, conn *net.TCPConn) ciface.IConn {
	c := &Connection{
		TcpServer:    NewServer(),
		Conn:         conn,
		ConnID:       0,
		MsgHandler:   client.GetMsgHandler(),
		ExitBuffChan: make(chan struct{}, 1),
		msgChan:      make(chan []byte),
		msgBuffChan:  make(chan []byte, cutils.GlobalObject.MaxMsgChanLen),
		property:     make(map[string]interface{}),
		onConnStart:  client.GetOnConnStart(),
		onConnStop:   client.GetOnConnStop(),
	}
	c.IsClosed.Store(false)
	c.IsAborted.Store(false)

	return c
}

func (c *Connection) StartReader() {
	fmt.Println("Reader Goroutine is  running")
	defer fmt.Println(c.RemoteAddr().String(), " conn reader exit!")
	defer c.Stop()

	for {

		dp := NewDataPack()

		headData := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(c.GetTCPConn(), headData); err != nil {
			fmt.Println("read msg head error ", err)
			c.IsAborted.Store(true)
			break
		}

		msg, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("unpack error ", err)
			c.IsAborted.Store(true)
			break
		}

		var data []byte
		if msg.GetDataLen() > 0 {
			data = make([]byte, msg.GetDataLen())
			if _, err := io.ReadFull(c.GetTCPConn(), data); err != nil {
				fmt.Println("read msg data error ", err)
				c.IsAborted.Store(true)
				break
			}
		}
		msg.SetData(data)

		req := Request{
			conn: c,
			msg:  msg,
		}

		if cutils.GlobalObject.WorkerPoolSize > 0 {
			c.MsgHandler.SendMsgToTaskQueue(&req)
		} else {
			go c.MsgHandler.DoMsgHandler(&req)
		}
	}
}

/*
写消息Goroutine， 用户将数据发送给客户端
*/
func (c *Connection) StartWriter() {

	fmt.Println("[Writer Goroutine is running]")
	defer fmt.Println(c.RemoteAddr().String(), "[conn Writer exit!]")

	for {
		select {
		case data, ok := <-c.msgChan:
			if ok {

				if _, err := c.Conn.Write(data); err != nil {
					fmt.Println("Send Data error:, ", err, " Conn Writer exit")
					return
				}
			} else {
				fmt.Println("msgChan is Closed")
				return
			}

		case data, ok := <-c.msgBuffChan:
			if ok {

				if _, err := c.Conn.Write(data); err != nil {
					fmt.Println("Send Buff Data error:, ", err, " Conn Writer exit")
					return
				}
			} else {
				fmt.Println("msgBuffChan is Closed")
				return
			}
		}
	}
}

func (c *Connection) Start() {
	go c.StartReader()
	go c.StartWriter()

	c.onConnStart(c)

	for {
		select {
		case <-c.ExitBuffChan:

			return
		}
	}
}

func (c *Connection) Stop() {
	if c.IsClosed.Load() {
		return
	}
	c.IsClosed.Store(true)

	c.onConnStop(c)

	c.Conn.Close()

	c.ExitBuffChan <- struct{}{}

	c.TcpServer.GetConnMgr().Remove(c)

	close(c.ExitBuffChan)
	close(c.msgBuffChan)
	close(c.msgChan)
}

func (c *Connection) GetTCPConn() *net.TCPConn {
	return c.Conn
}

func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

func (c *Connection) SendMsg(msgId uint32, data []byte) error {
	IsClosed := c.IsClosed.Load()
	if IsClosed {
		return errors.New("Connection closed when send msg")
	}

	dp := NewDataPack()
	msg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id = ", msgId)
		return errors.New("Pack error msg ")
	}

	c.msgChan <- msg

	return nil
}

func (c *Connection) SendBuffMsg(msgId uint32, data []byte) error {
	IsClosed := c.IsClosed.Load()
	if IsClosed {
		return errors.New("Connection closed when send buff msg")
	}

	dp := NewDataPack()
	msg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id = ", msgId)
		return errors.New("Pack error msg ")
	}

	c.msgBuffChan <- msg

	return nil
}

func (c *Connection) SetProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	c.property[key] = value
}

func (c *Connection) GetProperty(key string) (interface{}, bool) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	if value, ok := c.property[key]; ok {
		return value, true
	} else {
		return nil, false
	}
}

func (c *Connection) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	delete(c.property, key)
}
