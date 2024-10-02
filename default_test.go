package netc

import (
	"fmt"
	"sync"
	"testing"
)

func TestDialCtl(t *testing.T) {
	err := Default.SetLocalIP("172.16.95.1")
	if err != nil {
		panic(err)
	}
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		_, err = Default.Dial("tcp", "172.16.95.132:7001")
		if err == nil {
			fmt.Println("7001连接成功")
		}
		wg.Done()
	}()
	go func() {
		_, err = Default.Dial("tcp", "172.16.95.132:7002")
		if err == nil {
			fmt.Println("7002连接成功")
		}
		wg.Done()
	}()
	wg.Wait()
}
