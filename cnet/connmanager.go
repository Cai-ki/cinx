package cnet

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/Cai-ki/cinx/ciface"
)

type ConnManager struct {
	conns      sync.Map
	connsCount atomic.Int32
	isClosed   atomic.Bool
}

func NewConnManager() *ConnManager {
	return &ConnManager{}
}

func (connMgr *ConnManager) Add(conn ciface.IConn) error {
	if connMgr.isClosed.Load() {
		return errors.New("ConnManager closed")
	}

	connMgr.conns.Store(conn.GetConnID(), conn)
	connMgr.connsCount.Add(1)

	fmt.Println("connection add to ConnManager successfully: conn num = ", connMgr.Len())

	return nil
}

func (connMgr *ConnManager) Remove(conn ciface.IConn) {
	connMgr.conns.Delete(conn.GetConnID())
	connMgr.connsCount.Add(-1)

	fmt.Println("connection Remove ConnID=", conn.GetConnID(), " successfully: conn num = ", connMgr.Len())
}

func (connMgr *ConnManager) Get(connID uint32) (ciface.IConn, error) {
	if conn, ok := connMgr.conns.Load(connID); ok {
		return conn.(ciface.IConn), nil
	} else {
		return nil, errors.New("connection not found")
	}
}

func (connMgr *ConnManager) Len() int {
	return int(connMgr.connsCount.Load())
}

func (connMgr *ConnManager) ClearConns() {
	connMgr.isClosed.Store(true)

	connMgr.conns.Range(func(key, value any) bool {
		value.(ciface.IConn).Stop()
		return true
	})

	fmt.Println("Clear All Conns successfully: conn num = ", connMgr.Len())
}
