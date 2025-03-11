package ciface

import "net"

type IConn interface {
	Start()
	Stop()

	GetTCPConn() *net.TCPConn
	GetConnID() uint32

	RemoteAddr() net.Addr

	SendMsg(msgId uint32, data []byte) error
	SendBuffMsg(msgId uint32, data []byte) error

	SetProperty(key string, value interface{})
	GetProperty(key string) (interface{}, bool)
	RemoveProperty(key string)
}
