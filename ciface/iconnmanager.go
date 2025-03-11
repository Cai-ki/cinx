package ciface

/*
	连接管理抽象层
*/
type IConnManager interface {
	Add(conn IConn)                   //添加链接
	Remove(conn IConn)                //删除连接
	Get(connID uint32) (IConn, error) //利用ConnID获取链接
	Len() int                         //获取当前连接
	ClearConn()                       //删除并停止所有链接
}
