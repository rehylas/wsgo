package main

import (
	"fmt"
	"time"
	"wsgo/ws"
)

func main() {

	bizer := &ws.BaseBizer{SubPub: &ws.BaseSubPub{}}
	bizer.Clients = make(map[string]ws.AppClient)
	go testPub(bizer)
	go testBroad(bizer)
	ws := &ws.WsApi{}
	ws.Start(bizer, "/market", 8080)
}

func testSubPub() {
	sp := &ws.BaseSubPub{}
	sp.Sub("ru2201", "1001")
	sp.Sub("ru2201", "1002")
	sp.Sub("ru2201", "1003")
	sp.Unsub("ru2201", "1002")
	sp.Sub("ru2202", "1002")

	fmt.Println(sp.GetSub("ru2201"))
	fmt.Println(sp.GetSub("ru2202"))
	fmt.Println(sp.GetSub("ru2203"))
}

func testPub(biz *ws.BaseBizer) {
	for {
		time.Sleep(time.Second * 10)
		biz.Pub("ru2201", "ru2201 data")
	}
}

func testBroad(biz *ws.BaseBizer) {
	for {
		time.Sleep(time.Second * 10)
		biz.Broad("system broad msg ")
	}
}
