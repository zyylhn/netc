package netc

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

const l1 = "2006-01-02 15:04:05"

// PushEvent 推送事件接口
type PushEvent interface {
	// Push 根据索引将请求响应事件内容推送到其他地方
	Push(index interface{}, event ConnectEvent)
}

// ConnectInfo 连接信息
type ConnectInfo struct {
	//准备建立连接的时间
	GetConnectTime time.Time `json:"getConnectTime"`

	//成功建立连接的时间
	GotConnectTime time.Time `json:"gotConnectTime"`

	//请求目标的地址
	RemoteAddr string `json:"remoteAddr"`

	//请求目标的源地址
	LocalAddr net.Addr `json:"localAddr"`
}

// ConnectEvent 连接信息
type ConnectEvent struct {
	ConnectInfo ConnectInfo `json:"connectInfo"`
	Error       string      `json:"error"`
}

// EventWithIndex 请求事件信息
type EventWithIndex struct {
	ConnectEvent
	Index interface{} `json:"index"`
}

type PushEventToRemoteAddr struct {
	log  log.Logger
	lock sync.RWMutex
	conn net.Conn
}

func NewPushEventToRemoteAddr(addr string) (*PushEventToRemoteAddr, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	push := new(PushEventToRemoteAddr)
	push.conn = conn
	push.lock = sync.RWMutex{}
	return push, nil
}

func (p *PushEventToRemoteAddr) Push(index interface{}, event ConnectEvent) {
	p.lock.Lock()
	p.send(EventWithIndex{Index: index, ConnectEvent: event})
	p.lock.Unlock()
}

func (p *PushEventToRemoteAddr) send(data EventWithIndex) {
	d, err := json.Marshal(&data)
	if err != nil {
		panic(fmt.Sprintf("push request info marshal error:%v,index:%v,data:%v", err, data.Index, data.ConnectInfo))
	}
	_, err = p.conn.Write(append(d, []byte("\n")...))
	if err != nil {
		panic(fmt.Sprintf("push request info write to connect error:%v", err))
	}
}

func (p *PushEventToRemoteAddr) Close() {
	_ = p.conn.Close()
}
