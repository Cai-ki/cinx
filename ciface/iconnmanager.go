package ciface

type IConnManager interface {
	Add(conn IConn) error
	Remove(conn IConn)
	Get(connID uint32) (IConn, error)
	Len() int
	ClearConns()
}
