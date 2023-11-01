package npc

import (
	"fmt"
	"testing"
	"time"
)

//
func TestStartP2PClient(t *testing.T) {
	StartP2PClient("", "", "123123", 52000)
	time.Sleep(time.Duration(10) * time.Second)
	// SDK 中无效
	StopP2P()
	time.Sleep(time.Duration(100) * time.Second)
}

func TestChanGo(t *testing.T) {
	lanAddr := GetLanAddr("", "", "")
	fmt.Println(fmt.Sprintf("%s !!!!!!!!!!!!1xxxx", lanAddr))
	//StartP2PClient("", "", "", 11212)
	time.Sleep(1000 * time.Second)
}

func TestCloseGoroutine(t *testing.T) {
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				fmt.Println("exit goroutine 01")
				return
			default:
				fmt.Println("watch 01 ....")
				time.Sleep(1 * time.Second)
			}
		}
	}()

	go func() {
		for res := range done {
			fmt.Println(res) //没有消息则是阻塞状态 //chan 关闭则for循环结束
		}
		fmt.Println("exit goroutine 03")
	}()
	go func() {
		for {
			select {
			case <-done:
				fmt.Println("exit goroutine 02")
				return
			default:
				fmt.Println("watch 02.。。")
				time.Sleep(1 * time.Second)
			}

		}
	}()

	time.Sleep(3 * time.Second)
	close(done)
	time.Sleep(5 * time.Second)
	fmt.Println("退出程序")

}
