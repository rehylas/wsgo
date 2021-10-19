package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
)

const OP_LOGIN = "login"
const OP_LOGOUT = "logout"
const OP_REQ = "req"
const OP_SUB = "sub"
const OP_UNSUB = "unsub"
const OP_HB = "heartbeat"
const OP_ECHO = "echo"
const OP_P2P = "p2p"

const (
	RC_NOT_LOGINED = 1001
)

// const MSG_CH_MULTICAST = "multicast"
// const MSG_CH_SYSTEM = "system"
// const MSG_CH_P2P = "p2p"
// const MSG_CH_CHAT = "chat"
// const MSG_CH_MARKET = "market"

type WS_MSG struct {
	MsgType  string      `json:"msgtype"`
	SubTitle string      `json:"subtitle"`
	FromUser string      `json:"fromuser"`
	ToUser   string      `json:"touser"`
	RetCode  string      `json:"retcode"`
	Data     interface{} `json:"data"`
}

type AppClient struct {
	Userid   string
	Username string
	conn     *ServConnection
}
type InBaseBizer struct {
	name        string
	Clients     map[string]AppClient
	Mutex       sync.Mutex
	packetcount int64
	recdataSize int64
}

type BaseBizer struct {
	InBaseBizer
	Api    *WsApi
	SubPub SubPub
}

///////////////////////////////////////////////////////////////////////////
// 实现接口

//发现conn 断开
func (bizer *InBaseBizer) OnDisconnect(conn *ServConnection) {
	log.Println("OnDisconnect  ", conn)
	bizer.Mutex.Lock()
	for k, v := range bizer.Clients {
		if v.conn == conn {
			delete(bizer.Clients, k)
			bizer.Mutex.Unlock()
			return
		}

	}
	bizer.Mutex.Unlock()
	log.Println("clients :", len(bizer.Clients))

	//delete(bizer.Clients, 1)
}

//收到conn 数据
func (bizer *InBaseBizer) OnReceivedMsg(conn *ServConnection, data []byte) (isOk bool) {
	// beego.Informational("clinet receive msg :", conn, len(data))

	bizer.packetcount++
	bizer.recdataSize += int64(len(data))

	// bizer.Mutex.Lock() // 把锁加到 login 和 logout
	// bRet := bizer.checkData(conn, data)
	// bizer.Mutex.Unlock()
	return true
}

///////////////////////////////////////////////////////////////////////////
// 实现接口
type P2PMsg struct {
	ToUser string `json:"touser"`
}

type LoginData struct {
	Userid   string `json:"userid"`
	UserName string `json:"username"`
	UserPwd  string `json:"userpwd"`
}

type MS_MSG_LOGIN struct {
	MsgType string    `json:"msgtype"`
	Data    LoginData `json:"data"`
}

func (this *BaseBizer) SetWsApi(api *WsApi) {
	this.Api = api
	if this.SubPub == nil {
		this.SubPub = &BaseSubPub{}
	}
}

func (this *BaseBizer) SetSubPub(subpub SubPub) {
	this.SubPub = subpub

}

func (this *BaseBizer) OnReq(conn *ServConnection, data []byte) {
	fmt.Println("do req ")

	msg := WS_MSG{}
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return
	}

	var respMsg WS_MSG
	respMsg.FromUser = "system"
	respMsg.ToUser = msg.FromUser
	respMsg.MsgType = "resp"
	respMsg.Data = respMsg
	respData, err := json.Marshal(respMsg)
	this.Notice2P(respMsg.ToUser, respData)

}

func (this *BaseBizer) OnSub(conn *ServConnection, data []byte) {
	fmt.Println("do OnSub ")
	msg := WS_MSG{}
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return
	}
	log.Println("sub:", msg.SubTitle, msg.FromUser)
	this.SubPub.Sub(msg.SubTitle, msg.FromUser)
}

func (this *BaseBizer) OnUnsub(conn *ServConnection, data []byte) {
	fmt.Println("do OnUnsub ")
	msg := WS_MSG{}
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return
	}

	this.SubPub.Unsub(msg.SubTitle, msg.FromUser)
}

func (this *BaseBizer) OnLogin(conn *ServConnection, data []byte) {
	/*
		流程， 判断用户是否已存在， 如果已存在，修改conn
		如果不存在， 创建新的appClient
		当前版本，不校验密码和session ，
	*/
	fmt.Println("do login ")

	msg := MS_MSG_LOGIN{}
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return
	}

	// log.Println("msg:", msg)

	//
	this.Mutex.Lock()
	appclient, ok := this.Clients[msg.Data.Userid]
	if ok == true {
		log.Println("old conn:", appclient.conn)
		log.Println("new conn:", conn)
		if appclient.conn != conn {
			log.Println("old conn != new conn : conn.Close")
			appclient.conn.Close()
		}
		appclient.conn = conn
		this.Clients[msg.Data.Userid] = appclient

	} else {
		this.Clients[msg.Data.Userid] = AppClient{conn: conn}
		log.Println("new conn:", conn)
	}
	this.Mutex.Unlock()
	log.Println("Userid:", msg.Data)
	log.Println("Userid:", msg.Data.Userid, len(this.Clients), this.packetcount, this.recdataSize)
	noticeInfo := fmt.Sprintf(`{"msgtype":"login","retcode":0,"data":{"userid":"%s","info":"login successful "}}`, msg.Data.Userid)

	this.Notice2P(msg.Data.Userid, []byte(noticeInfo))
	//bizer.Broadcast([]byte(noticeInfo))
	return
}

func (this *BaseBizer) OnLogout(conn *ServConnection, data []byte) {
	fmt.Println("do logout ")
	msg := MS_MSG_LOGIN{}
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return
	}

	this.Mutex.Lock()
	_, ok := this.Clients[msg.Data.Userid]
	if ok {
		delete(this.Clients, msg.Data.Userid)
	}
	conn.Close()
	this.Mutex.Unlock()

	return
}

func (this *BaseBizer) OnP2P(conn *ServConnection, data []byte) {
	msg := P2PMsg{}
	err := json.Unmarshal([]byte(data), &msg)
	if err != nil {
		return
	}

	if msg.ToUser == "" {
		return
	}
	fmt.Println("Userid:", msg.ToUser)
	this.Mutex.Lock()
	appclient, ok := this.Clients[msg.ToUser]
	if ok {
		appclient.conn.WriteMessage([]byte(data))
	}
	this.Mutex.Unlock()

	return

}

///////////////////////////////////////////////////////////////////////////
// 内部操作

//收到conn 数据
func (this *BaseBizer) OnReceivedMsg(conn *ServConnection, data []byte) (isOk bool) {
	// beego.Informational("clinet receive msg :", conn, len(data))

	this.packetcount++
	this.recdataSize += int64(len(data))

	// bizer.Mutex.Lock() // 把锁加到 login 和 logout
	bRet := this.ProData(conn, data)
	// bizer.Mutex.Unlock()
	return bRet
}

func (this *BaseBizer) ProData(conn *ServConnection, data []byte) (isOk bool) {
	msg := WS_MSG{}
	err := json.Unmarshal(data, &msg)
	if err != nil {
		log.Println("receive err data:", string(data))
		return false
	}
	if msg.MsgType == OP_HB {
		// beego.Informational("heartbeat ")
		return true
	}
	if msg.MsgType == OP_LOGIN {
		this.OnLogin(conn, data)
		return true
	}

	// check logined
	logined := true
	if logined == false {
		msg.RetCode = string(RC_NOT_LOGINED)
		repData, _ := json.Marshal(msg)
		conn.WriteMessage(repData)
		return false
	}

	if msg.MsgType == OP_LOGOUT {
		this.OnLogout(conn, data)
		return true
	}

	if msg.MsgType == OP_REQ {
		this.OnReq(conn, data)
		return true
	}

	if msg.MsgType == OP_SUB {
		this.OnSub(conn, data)
		return true
	}

	if msg.MsgType == OP_UNSUB {
		this.OnUnsub(conn, data)
		return true
	}

	if msg.MsgType == OP_P2P {
		this.OnP2P(conn, data)
		return true
	}

	if msg.MsgType == OP_ECHO {
		conn.WriteMessage(data)
		return true
	}

	log.Println("msgtype error :", msg.MsgType)
	return false
}

//发送数据给特定的客户端
func (this *BaseBizer) Notice2P(userid string, data []byte) {

	appclient, ok := this.Clients[userid]
	if ok == true {
		err := appclient.conn.WriteMessage(data)
		if err != nil {
			log.Println("appclient.conn.WriteMessage(data) error :", userid, err)
		}

	} else {
		//warn ....
		log.Println("Notice2P cant find appclient , userid:", userid)
	}

}

//广播数据给所有客户端
func (bizer *BaseBizer) Broadcast(data []byte) {

	bizer.Mutex.Lock()
	for _, v := range bizer.Clients {
		v.conn.WriteMessage(data)
	}
	bizer.Mutex.Unlock()

	//delete(bizer.Clients, 1)
}

///////////////////////////////////////////////////////////////////////////
// 业务操作操作
// func (this *WsApi) Resp(toUser string, data interface{}) {

// }

func (this *BaseBizer) Pub(title string, data interface{}) {
	var respMsg WS_MSG
	respMsg.FromUser = "system"
	respMsg.ToUser = ""
	respMsg.MsgType = "pub"
	respMsg.SubTitle = title
	respMsg.Data = data
	respData, err := json.Marshal(respMsg)
	if err != nil {
		return
	}

	users := this.SubPub.GetSub(title)
	for _, user := range users {
		this.Notice2P(user, respData)
	}
}

// func (this *WsApi) P2P() {

// }

func (this *BaseBizer) Broad(data interface{}) {
	var respMsg WS_MSG
	respMsg.FromUser = "system"
	respMsg.ToUser = ""
	respMsg.MsgType = "broad"
	respMsg.Data = data
	respData, err := json.Marshal(respMsg)
	if err != nil {
		return
	}
	log.Println("broadcast")
	this.Broadcast(respData)
}
