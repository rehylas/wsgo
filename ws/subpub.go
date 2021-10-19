package ws

import (
	"sync"
)

type SubPub interface {
	Sub(title string, userid string)
	Unsub(title string, userid string)
	GetSub(title string) (users []string)
}

type BaseSubPub struct {
	titles map[string]map[string]bool // string 1  is title    ,  string 2 is userid
	mutex  sync.Mutex
}

func (this *BaseSubPub) Sub(title string, userid string) {
	this.mutex.Lock()
	if this.titles == nil {
		this.titles = make(map[string]map[string]bool)
	}
	subTit, ok := this.titles[title]
	if ok {
		subTit[userid] = true
	} else {
		this.titles[title] = map[string]bool{}
		this.titles[title][userid] = true
	}
	this.mutex.Unlock()
}

func (this *BaseSubPub) Unsub(title string, userid string) {
	this.mutex.Lock()
	if this.titles == nil {
		this.titles = make(map[string]map[string]bool)
	}
	_, ok := this.titles[title]
	if ok == false {
		// 不操作
	} else {
		delete(this.titles[title], userid)
		// if len(this.titles[title]) == 0 {  不删除，防止频繁增删 耗费一些内存
		// 	delete(this.titles, title)
		// }
	}
	this.mutex.Unlock()
}
func (this *BaseSubPub) GetSub(title string) (users []string) {
	var rest []string
	this.mutex.Lock()
	if this.titles == nil {
		this.titles = make(map[string]map[string]bool)
	}
	subTit, ok := this.titles[title]
	if ok {
		for key, _ := range subTit { //val
			rest = append(rest, key)
		}
	} else {

	}
	this.mutex.Unlock()
	return rest
}

// set := make(map[string]bool)
