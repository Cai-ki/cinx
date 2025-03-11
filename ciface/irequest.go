package ciface

type IRequest interface {
	GetConn() IConn
	GetData() []byte
	GetMsgID() uint32
}
