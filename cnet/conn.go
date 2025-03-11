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
	//当前Conn属于哪个Server
	TcpServer ciface.IServer //当前conn属于哪个server，在conn初始化的时候添加即可
	//当前连接的socket TCP套接字
	Conn *net.TCPConn
	//当前连接的ID 也可以称作为SessionID，ID全局唯一
	ConnID uint32

	// 正常退出
	IsClosed atomic.Bool

	// 异常退出
	IsAborted atomic.Bool

	//消息管理MsgId和对应处理方法的消息管理模块
	MsgHandler ciface.IMsgHandle

	ExitBuffChan chan struct{}

	msgChan     chan []byte
	msgBuffChan chan []byte

	property     map[string]interface{}
	propertyLock sync.RWMutex
}

// 创建连接的方法
func NewConntion(server ciface.IServer, conn *net.TCPConn, connID uint32, msgHandler ciface.IMsgHandle) *Connection {
	c := &Connection{
		TcpServer:    server, //将隶属的server传递进来
		Conn:         conn,
		ConnID:       connID,
		MsgHandler:   msgHandler,
		ExitBuffChan: make(chan struct{}, 1),
		msgChan:      make(chan []byte), //msgChan初始化
		msgBuffChan:  make(chan []byte, cutils.GlobalObject.MaxMsgChanLen),
		property:     make(map[string]interface{}), //对链接属性map初始化
	}
	c.IsClosed.Store(false)
	c.IsAborted.Store(false)

	c.TcpServer.GetConnMgr().Add(c)
	return c
}

func NewClientConn(client ciface.IClient, conn *net.TCPConn) ciface.IConn {
	c := &Connection{
		TcpServer:    NewServer(), // TODO: 临时创建一个server，后续需要修改
		Conn:         conn,
		ConnID:       0, // ignore
		MsgHandler:   client.GetMsgHandler(),
		ExitBuffChan: make(chan struct{}, 1),
		msgChan:      make(chan []byte), //msgChan初始化
		msgBuffChan:  make(chan []byte, cutils.GlobalObject.MaxMsgChanLen),
		property:     make(map[string]interface{}), //对链接属性map初始化
	}
	c.IsClosed.Store(false)
	c.IsAborted.Store(false)

	c.TcpServer.SetOnConnStart(client.GetOnConnStart())
	c.TcpServer.SetOnConnStop(client.GetOnConnStop())
	return c
}

func (c *Connection) StartReader() {
	fmt.Println("Reader Goroutine is  running")
	defer fmt.Println(c.RemoteAddr().String(), " conn reader exit!")
	defer c.Stop()

	for {
		// 创建拆包解包的对象
		dp := NewDataPack()

		//读取客户端的Msg head
		headData := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(c.GetTCPConn(), headData); err != nil {
			fmt.Println("read msg head error ", err)
			c.IsAborted.Store(true)
			break
		}

		//拆包，得到msgid 和 datalen 放在msg中
		msg, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("unpack error ", err)
			c.IsAborted.Store(true)
			break
		}

		//根据 dataLen 读取 data，放在msg.Data中
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
				//有数据要写给客户端
				if _, err := c.Conn.Write(data); err != nil {
					fmt.Println("Send Data error:, ", err, " Conn Writer exit")
					return
				}
			} else {
				fmt.Println("msgChan is Closed")
				return
			}
			//针对有缓冲channel需要些的数据处理
		case data, ok := <-c.msgBuffChan:
			if ok {
				//有数据要写给客户端
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

// 启动连接，让当前连接开始工作
func (c *Connection) Start() {

	//1 开启用户从客户端读取数据流程的Goroutine
	go c.StartReader()
	//2 开启用于写回客户端数据流程的Goroutine
	go c.StartWriter()

	//按照用户传递进来的创建连接时需要处理的业务，执行钩子方法
	c.TcpServer.CallOnConnStart(c)

	for {
		select {
		case <-c.ExitBuffChan:
			// 得到退出消息，不再阻塞
			return
		}
	}
}

// 停止连接，结束当前连接状态
func (c *Connection) Stop() {
	if c.IsClosed.Load() {
		return
	}
	c.IsClosed.Store(true)

	c.TcpServer.CallOnConnStop(c)

	// 关闭socket链接
	c.Conn.Close()

	c.ExitBuffChan <- struct{}{}

	// 将链接从连接管理器中删除
	c.TcpServer.GetConnMgr().Remove(c)

	close(c.ExitBuffChan)
	close(c.msgBuffChan)
	close(c.msgChan)
}

// 从当前连接获取原始的socket TCPConn
func (c *Connection) GetTCPConn() *net.TCPConn {
	return c.Conn
}

// 获取当前连接ID
func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}

// 获取远程客户端地址信息
func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

// 直接将Message数据发送数据给远程的TCP客户端
func (c *Connection) SendMsg(msgId uint32, data []byte) error {
	IsClosed := c.IsClosed.Load()
	if IsClosed {
		return errors.New("Connection closed when send msg")
	}
	//将data封包，并且发送
	dp := NewDataPack()
	msg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id = ", msgId)
		return errors.New("Pack error msg ")
	}

	//写回客户端
	c.msgChan <- msg //将之前直接回写给conn.Write的方法 改为 发送给Channel 供Writer读取

	return nil
}

func (c *Connection) SendBuffMsg(msgId uint32, data []byte) error {
	IsClosed := c.IsClosed.Load()
	if IsClosed {
		return errors.New("Connection closed when send buff msg")
	}
	//将data封包，并且发送
	dp := NewDataPack()
	msg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id = ", msgId)
		return errors.New("Pack error msg ")
	}

	//写回客户端
	c.msgBuffChan <- msg

	return nil
}

// 设置链接属性
func (c *Connection) SetProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	c.property[key] = value
}

// 获取链接属性
func (c *Connection) GetProperty(key string) (interface{}, bool) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	if value, ok := c.property[key]; ok {
		return value, true
	} else {
		return nil, false
	}
}

// 移除链接属性
func (c *Connection) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	delete(c.property, key)
}
