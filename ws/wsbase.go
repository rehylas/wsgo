package ws

import (
	"errors"
	"sync"

	"github.com/gorilla/websocket"
)

const INTERVAL_HEARTBEAT = 10

type inBizer interface {
	OnDisconnect(conn *ServConnection)
	OnReceivedMsg(conn *ServConnection, data []byte) (isOk bool)
	// OnLingin(conn *Connection, userid string, data interface{})
	// OnLingout(conn *Connection, userid string, data interface{})
	// OnReq(conn *Connection, userid string, data interface{})
	// OnSub(conn *Connection, userid string, data interface{})
	// OnUnsub(conn *Connection, userid string, data interface{})
}

type ServConnection struct {
	wsConnect *websocket.Conn
	inChan    chan []byte
	outChan   chan []byte
	closeChan chan byte
	mutex     sync.Mutex // 对closeChan关闭上锁
	isClosed  bool       // 防止closeChan被关闭多次
	bizer     inBizer    //处理业务
}

func (this *ServConnection) InitConnection(wsConn *websocket.Conn, bizer inBizer) (connect *ServConnection, err error) {

	conn := &ServConnection{}
	conn.wsConnect = wsConn
	conn.inChan = make(chan []byte, 1000)
	conn.outChan = make(chan []byte, 1000)
	conn.closeChan = make(chan byte, 1)
	conn.bizer = bizer

	// 启动读协程
	go conn.readLoop()
	// 启动写协程
	go conn.writeLoop()
	return conn, nil
}

func (conn *ServConnection) ReadMessage() (data []byte, err error) {

	select {
	case data = <-conn.inChan:
	case <-conn.closeChan:
		err = errors.New("connection is closeed")
	}
	return
}

func (conn *ServConnection) WriteMessage(data []byte) (err error) {

	select {
	case conn.outChan <- data:
	case <-conn.closeChan:
		err = errors.New("connection is closeed")
	}
	return
}

func (conn *ServConnection) Close() {
	// 线程安全，可多次调用
	conn.wsConnect.Close()
	// 利用标记，让closeChan只关闭一次
	conn.mutex.Lock()
	if !conn.isClosed {
		close(conn.closeChan)
		conn.isClosed = true
	}
	conn.mutex.Unlock()
}

// 内部实现
func (conn *ServConnection) readLoop() {
	var (
		data []byte
		err  error
	)
	for {
		if _, data, err = conn.wsConnect.ReadMessage(); err != nil {
			goto ERR
		}
		//阻塞在这里，等待inChan有空闲位置
		select {
		case conn.inChan <- data:
		case <-conn.closeChan: // closeChan 感知 conn断开
			goto ERR
		}

	}

ERR:

	conn.Close()
	conn.bizer.OnDisconnect(conn) //注意位置
}

func (conn *ServConnection) writeLoop() {
	var (
		data []byte
		err  error
	)

	for {
		select {
		case data = <-conn.outChan:
		case <-conn.closeChan:
			goto ERR
		}
		if err = conn.wsConnect.WriteMessage(websocket.TextMessage, data); err != nil {
			goto ERR
		}
	}

ERR:

	conn.Close()
	conn.bizer.OnDisconnect(conn)

}
