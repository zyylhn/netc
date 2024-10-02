package netc

import (
	"context"
	"net"
	"time"
)

var Default *Dialer

func Dial(network string, addr string) (net.Conn, error) {
	return Default.Dial(network, addr)
}

func DialWithTimeout(network string, addr string, timeout time.Duration) (net.Conn, error) {
	return Default.DialWithTimeout(network, addr, timeout)
}

func DialWithContext(ctx context.Context, network string, addr string) (net.Conn, error) {
	return Default.DialWithContext(ctx, network, addr)
}

func DialWithIndex(network string, addr string, index interface{}) (net.Conn, error) {
	return Default.DialWithIndex(network, addr, index)
}

func DialTcpWithTimeoutIndex(addr string, timeout time.Duration, index interface{}) (net.Conn, error) {
	return Default.DialTcpWithTimeoutIndex(addr, timeout, index)
}

// DialTcpWithTimeoutIndexLocalIp 自行获取连接目标的地址
func DialTcpWithTimeoutIndexLocalIp(addr string, timeout time.Duration, index interface{}) (net.Conn, error) {
	return Default.DialTcpWithTimeoutIndexLocalIp(addr, timeout, index)
}

func DialCtl(ctx context.Context, network string, addr string, timeout time.Duration, index interface{}) (net.Conn, error) {
	return Default.DialCtl(ctx, network, addr, timeout, index)
}

func SetLocalIP(ip string) error {
	return Default.SetLocalIP(ip)
}

func RemoveLocalIp() {
	Default.RemoveLocalIp()
}

func AppendEventPush(push PushEvent) {
	Default.AppendEventPush(push)
}

func init() {
	Default = NewDialer()
}
