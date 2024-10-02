package netc

import (
	"context"
	"fmt"
	"github.com/zyylhn/getlocaladdr"
	"golang.org/x/net/proxy"
	"log"
	"net"
	"sync"
	"time"
)

type Dialer struct {
	PushEvent []PushEvent //推送请求响应事件接口

	proxy     string
	proxyDial proxy.Dialer
	log       log.Logger
	localIP   string
	usingPort map[int]struct{} //为了方便删除
	lock      sync.RWMutex
}

func NewDialer() *Dialer {
	re := new(Dialer)
	re.usingPort = make(map[int]struct{})
	return re
}

func (d *Dialer) SetLocalIP(ip string) error {
	Ip := net.ParseIP(ip)
	if Ip == nil {
		return fmt.Errorf("ip address %v parsing failure", ip)
	}
	d.localIP = ip
	d.usingPort = make(map[int]struct{})
	return nil
}

func (d *Dialer) RemoveLocalIp() {
	d.localIP = ""
	d.usingPort = make(map[int]struct{})
}

func (d *Dialer) dialCtl(ctx context.Context, network string, addr string, timeout time.Duration, index interface{}, localIP string) (net.Conn, error) {
	var err error
	var conn net.Conn
	var event ConnectEvent
	connectInfo := ConnectInfo{RemoteAddr: addr} //这里不以真实连接的地址为准，以指定的目标为准。这样可以避免代理的干扰
	defer func() {
		for _, p := range d.PushEvent {
			p.Push(index, event)
		}
	}()
	var localIp string
	var localPort int
	defer func() {
		if localIp != "" {
			connectInfo.LocalAddr = &net.TCPAddr{
				IP:   net.ParseIP(localIp),
				Port: localPort,
				Zone: "",
			}
		}
		if conn != nil {
			connectInfo.LocalAddr = conn.LocalAddr()
			connectInfo.GotConnectTime = time.Now()
		}
		event.ConnectInfo = connectInfo
		if err != nil {
			event.Error = err.Error()
		}
	}()
	if d.localIP != "" {
		localIp = d.localIP
	}
	if localIP != "" {
		localIp = localIP
	}
	if d.proxyDial != nil {
		connectInfo.GetConnectTime = time.Now()
		//todo 代理功能不能指定超时和ctx容易发生异常
		conn, err = d.proxyDial.Dial(network, addr)
	} else {
		dialer := net.Dialer{}
		if localIp != "" {
			localPort = d.getFreePort()
			dialer.LocalAddr = &net.TCPAddr{IP: net.ParseIP(localIp), Port: localPort}
			defer func() {
				//todo 使用完从白名单中剔除端口白名单，这个白名单只是保证不要让两个线程同时获取到了一个端口号。当端口出现被占用的状态时也就不会获取到
				d.lock.Lock()
				delete(d.usingPort, localPort)
				d.lock.Unlock()
			}()
		}
		if timeout != 0 {
			dialer.Timeout = timeout
		}
		connectInfo.GetConnectTime = time.Now()
		if ctx != nil {
			conn, err = dialer.DialContext(ctx, network, addr)
		} else {
			conn, err = dialer.Dial(network, addr)
		}
	}
	return conn, err
}

func (d *Dialer) getFreePort() int {
	//todo 获取完成端口添加到白名单中。暂时这块全加上锁，如果对速度影响较大的话在考虑复制提升速度
	d.lock.Lock()
	defer func() {
		d.lock.Unlock()
	}()
	port := getLocalAddr.GetFreePortMap(d.usingPort)
	d.usingPort[port] = struct{}{}
	return port
}

func (d *Dialer) Dial(network string, addr string) (net.Conn, error) {
	return d.dialCtl(context.Background(), network, addr, 0, nil, "")
}

func (d *Dialer) DialWithLocalAddr(network string, connectAddr string, localIP string) (net.Conn, error) {
	return d.dialCtl(context.Background(), network, connectAddr, 0, nil, localIP)
}

func (d *Dialer) DialWithTimeout(network string, addr string, timeout time.Duration) (net.Conn, error) {
	return d.dialCtl(context.Background(), network, addr, timeout, nil, "")
}

func (d *Dialer) DialWithContext(ctx context.Context, network string, addr string) (net.Conn, error) {
	return d.dialCtl(ctx, network, addr, 0, nil, "")
}

func (d *Dialer) DialWithIndex(network string, addr string, index interface{}) (net.Conn, error) {
	return d.dialCtl(context.Background(), network, addr, 0, index, "")
}

func (d *Dialer) DialTcpWithTimeoutIndex(addr string, timeout time.Duration, index interface{}) (net.Conn, error) {
	return d.dialCtl(context.Background(), "tcp", addr, timeout, index, "")
}

// DialTcpWithTimeoutIndexLocalIp 自行获取连接目标地址去连接目标
func (d *Dialer) DialTcpWithTimeoutIndexLocalIp(addr string, timeout time.Duration, index interface{}) (net.Conn, error) {
	var localIP string
	if d.localIP == "" {
		ip, _, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}
		localIP = getLocalAddr.GetLocalIPWithTargetIP(ip)
		if localIP == "" {
			return nil, fmt.Errorf("get route to target error")
		}
	}

	return d.dialCtl(context.Background(), "tcp", addr, timeout, index, localIP)
}

func (d *Dialer) DialCtl(ctx context.Context, network string, addr string, timeout time.Duration, index interface{}) (net.Conn, error) {
	return d.dialCtl(ctx, network, addr, timeout, index, "")
}

func (d *Dialer) AppendEventPush(push PushEvent) {
	d.PushEvent = append(d.PushEvent, push)
}
