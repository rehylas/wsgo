package ws

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		// 允许跨域
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

type Bizer interface {
	OnDisconnect(conn *ServConnection)
	OnReceivedMsg(conn *ServConnection, data []byte) (isOk bool)

	OnReq(conn *ServConnection, data []byte)
	OnP2P(conn *ServConnection, data []byte)
	OnSub(conn *ServConnection, data []byte)
	OnUnsub(conn *ServConnection, data []byte)
	OnLogin(conn *ServConnection, data []byte)
	OnLogout(conn *ServConnection, data []byte)
	SetWsApi(api *WsApi)
}

type WsApi struct {
	CustBizer Bizer
	Serv      ServConnection // only for create conn
}

func (this *WsApi) Start(bizer Bizer, path string, port int) { // ws ="/ws" port = 8080  //clients: make(map[string]AppClient)
	bizer.SetWsApi(this)
	this.CustBizer = bizer
	http.HandleFunc(path, this.wsHandler)
	listenStr := fmt.Sprintf("0.0.0.0:%d", port)
	fmt.Println(listenStr)
	err := http.ListenAndServe(listenStr, nil)
	if err != nil {
		log.Println("ListenAndServe:", err)
		return

	}
}

// func (this *WsApi) Resp(toUser string, data interface{}) {

// }

// func (this *WsApi) Pub() {

// }

// func (this *WsApi) P2P() {

// }

// func (this *WsApi) Broad() {

// }

//////////////////////////////////////////////////////////////////////////////////
//  内部

func (this *WsApi) wsHandler(w http.ResponseWriter, r *http.Request) {

	var wsConn *websocket.Conn
	var connClient *ServConnection
	var err error
	var data []byte

	// 完成ws协议的握手操作
	// Upgrade:websocket
	if wsConn, err = upgrader.Upgrade(w, r, nil); err != nil {
		return
	}

	if connClient, err = this.Serv.InitConnection(wsConn, this.CustBizer); err != nil { //create new connect for client
		goto ERR
	}

	// 心跳线程，不断发消息
	go func() {
		var (
			err error
		)
		for {
			if err = connClient.WriteMessage([]byte(`{"msgtype":"heartbeat"}`)); err != nil {
				return
			}
			time.Sleep(INTERVAL_HEARTBEAT * time.Second)
		}
	}()

	for {
		// beego.Debug("ReadMessage .....")
		if data, err = connClient.ReadMessage(); err != nil {
			goto ERR
		}

		// log.Println("receive msg:", string(data))
		if connClient.bizer.OnReceivedMsg(connClient, data) == false {
			//conn.WriteMessage([]byte(`{"msgtype":"error","data":"error data format"}`))
			goto ERR
		}

		// if err = conn.WriteMessage(data); err != nil {
		// 	goto ERR
		// }
	}

ERR:
	connClient.bizer.OnDisconnect(connClient)
	connClient.Close()
}
