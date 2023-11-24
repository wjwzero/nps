package callback

import "github.com/astaxie/beego/logs"

var Callback P2PCallback

type P2PCallback interface {
	OnConnectSuccess()
	OnP2PConnectSuccess()
}

func RegisterCallback(c P2PCallback) {
	Callback = c
}

const (
	OnConnectSuccess    = "OnConnectSuccess"
	OnP2PConnectSuccess = "OnP2PConnectSuccess"
)

func TriggerCallBack(callback string) {
	logs.Info("trigger npc callback !!!")
	if Callback == nil {
		logs.Error("npc callback is null !!!")
		return
	}
	switch callback {
	case OnConnectSuccess:
		Callback.OnConnectSuccess()
	case OnP2PConnectSuccess:
		Callback.OnP2PConnectSuccess()
	default:
		logs.Warning("Is not set callback : %s", callback)
	}
}
