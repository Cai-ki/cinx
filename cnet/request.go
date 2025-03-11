package cnet

import "github.com/Cai-ki/cinx/ciface"

type Request struct {
	conn ciface.IConn    // 请求归属的连接
	msg  ciface.IMessage // 解析好的消息
}

//获取请求连接信息
func (r *Request) GetConn() ciface.IConn {
	return r.conn
}

//获取请求消息的数据
func (r *Request) GetData() []byte {
	return r.msg.GetData()
}

//获取请求的消息的ID
func (r *Request) GetMsgID() uint32 {
	return r.msg.GetMsgId()
}
