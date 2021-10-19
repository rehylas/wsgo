package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"golang.org/x/net/websocket"
)

const TIME_WRITE_INTERVAL = 30

//var addr = flag.String("addr", "192.168.2.110:9090", "http service address")
// var addr = flag.String("addr", "192.168.2.144:9090", "http service address")
var addr = flag.String("addr", "192.168.0.57:8080", "http service address")
var servip = "192.168.2.144:9090"
var __count = 0

func main() {

	fmt.Println("example: client.exe 10 127.0.0.1:8080")
	rand.Seed(time.Now().UnixNano())

	ncount := 1000
	if len(os.Args) >= 2 {
		ncount, _ = strconv.Atoi(os.Args[1])
	}

	if len(os.Args) >= 3 {
		servip = os.Args[2]
	}

	MakeMultClient(ncount)
	//MakeWsClient()

}

func MakeMultClient(count int) {
	for i := 0; i < count; i++ {
		__count++
		fmt.Println("client count:", __count)
		go MakeWsClient()
		time.Sleep(time.Millisecond * 100)
	}
	MakeWsClient()
}

func MakeWsClient() {

	defer func() { // 必须要先声明defer，否则不能捕获到panic异常
		fmt.Println("d----------")
		if err := recover(); err != nil {
			fmt.Println(err) // 这里的err其实就是panic传入的内容
		}
		fmt.Println("e-----------")
	}()

	flag.Parse()
	url := "ws://" + servip + "/market"
	origin := "test://1111111/"
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		fmt.Println(err)
	}

	go timeWriter(ws)

	//login
	time.Sleep(time.Millisecond * 100)

	logdata := fmt.Sprintf(`{"msgtype":"login","data":{"userid":"%08d"}}`, rand.Intn(int(time.Now().Unix()%99999999)))
	websocket.Message.Send(ws, logdata)

	for {
		var msg [512]byte
		_, err := ws.Read(msg[:]) //此处阻塞，等待有数据可读
		if err != nil {
			fmt.Println("read:", err)
			return
		}

		fmt.Printf("received: %s\n", msg)
	}

}

func timeWriter(conn *websocket.Conn) {
	for {
		time.Sleep(time.Second * TIME_WRITE_INTERVAL)
		websocket.Message.Send(conn, `{"msgtype":"heartbeat"}`)
	}
}
