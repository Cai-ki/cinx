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

	IsClosed     atomic.Bool
	IsAborted    atomic.Bool
	IsClosedOnce atomic.Bool

	MsgHandler ciface.IMsgHandle

	ExitChan chan struct{}

	msgChan     chan []byte
	msgBuffChan chan []byte

	// readerClosedChan chan struct{}
	writerClosedChan chan struct{}

	property     map[string]interface{}
	propertyLock sync.RWMutex

	onConnStart func(conn ciface.IConn)
	onConnStop  func(conn ciface.IConn)

	wg *sync.WaitGroup
}

func NewConntion(server ciface.IServer, conn *net.TCPConn, connID uint32, msgHandler ciface.IMsgHandle) (*Connection, error) {
	c := &Connection{
		TcpServer:        server,
		Conn:             conn,
		ConnID:           connID,
		MsgHandler:       msgHandler,
		ExitChan:         make(chan struct{}),
		msgChan:          make(chan []byte),
		msgBuffChan:      make(chan []byte, cutils.GlobalObject.MaxMsgChanLen),
		writerClosedChan: make(chan struct{}),
		property:         make(map[string]interface{}),
		onConnStart:      server.GetOnConnStart(),
		onConnStop:       server.GetOnConnStop(),
		wg:               &sync.WaitGroup{},
	}
	c.IsClosed.Store(false)
	c.IsAborted.Store(false)
	c.IsClosedOnce.Store(false)

	err := server.GetConnMgr().Add(c)
	return c, err
}

func NewClientConn(client ciface.IClient, conn *net.TCPConn) ciface.IConn {
	c := &Connection{
		TcpServer:   NewServer(),
		Conn:        conn,
		ConnID:      0,
		MsgHandler:  client.GetMsgHandler(),
		ExitChan:    make(chan struct{}),
		msgChan:     make(chan []byte),
		msgBuffChan: make(chan []byte, cutils.GlobalObject.MaxMsgChanLen),
		property:    make(map[string]interface{}),
		onConnStart: client.GetOnConnStart(),
		onConnStop:  client.GetOnConnStop(),
		wg:          &sync.WaitGroup{},
	}
	c.IsClosed.Store(false)
	c.IsAborted.Store(false)

	return c
}

func (c *Connection) StartReader() {
	fmt.Println("Reader Goroutine is  running")
	// defer fmt.Println(c.RemoteAddr().String(), " conn reader exit!")
	defer c.Stop()

	for {
		dp := NewDataPack()

		headData := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(c.GetTCPConn(), headData); err != nil {
			c.IsClosed.Store(true)
			if err == io.EOF {
				fmt.Println("Connection closed by peer")
			} else {
				fmt.Println("read msg head error ", err)
				c.IsAborted.Store(true)
			}
			break
		}

		msg, err := dp.Unpack(headData)
		if err != nil {
			c.IsClosed.Store(true)
			c.IsAborted.Store(true)
			fmt.Println("unpack error ", err)
			break
		}

		var data []byte
		if msg.GetDataLen() > 0 {
			data = make([]byte, msg.GetDataLen())
			if _, err := io.ReadFull(c.GetTCPConn(), data); err != nil {
				c.IsClosed.Store(true)
				c.IsAborted.Store(true)
				fmt.Println("read msg data error ", err)
				break
			}
		}
		msg.SetData(data)

		req := Request{
			conn: c,
			msg:  msg,
			wg:   c.wg,
		}

		c.wg.Add(1)
		if cutils.GlobalObject.WorkerPoolSize > 0 {
			c.MsgHandler.SendMsgToTaskQueue(&req)
		} else {
			go c.MsgHandler.DoMsgHandler(&req)
		}
	}
}

func (c *Connection) StartWriter() {

	fmt.Println("[Writer Goroutine is running]")
	// defer fmt.Println(c.RemoteAddr().String(), "[conn Writer exit!]")

	for {
		select {
		case data, ok := <-c.msgChan:
			if ok {
				if _, err := c.Conn.Write(data); err != nil {
					fmt.Println("Send Data error:, ", err, " Conn Writer exit")
					// return
				}
			} else {
				fmt.Println("msgChan is Closed")
				// return
			}

		case data, ok := <-c.msgBuffChan:
			if ok {

				if _, err := c.Conn.Write(data); err != nil {
					fmt.Println("Send Buff Data error:, ", err, " Conn Writer exit")
					// return
				}
			} else {
				fmt.Println("msgBuffChan is Closed")
				// return
			}
		}
	}
}

func (c *Connection) Start() {
	go c.StartWriter()
	go c.StartReader()

	c.onConnStart(c)

	// for {
	// 	select {
	// 	case <-c.ExitBuffChan:

	// 		return
	// 	}
	// }
}

func (c *Connection) Stop() {
	if c.IsClosedOnce.Load() {
		return
	}
	c.IsClosedOnce.Store(true)
	c.IsClosed.Store(true)

	c.wg.Wait()

	c.onConnStop(c)

	c.Conn.Close()

	c.TcpServer.GetConnMgr().Remove(c)

	// c.ExitBuffChan <- struct{}{}
	// close(c.ExitBuffChan)
	// close(c.msgBuffChan)
	// close(c.msgChan)
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

	dp := NewDataPack()
	msg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id = ", msgId)
		return errors.New("Pack error msg ")
	}

	// if c.IsClosed.Load() {
	// 	return errors.New("Connection closed when send msg")
	// }
	c.msgChan <- msg

	return nil
}

func (c *Connection) SendBuffMsg(msgId uint32, data []byte) error {
	dp := NewDataPack()
	msg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id = ", msgId)
		return errors.New("Pack error msg ")
	}

	// if c.IsClosed.Load() {
	// 	return errors.New("Connection closed when send buff msg")
	// }
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
