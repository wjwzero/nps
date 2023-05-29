package preclient

import (
	"ehang.io/nps/lib/file"
	"ehang.io/nps/server/tool"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
)

func P2pClient(clientId int, deviceKey string, password string) (t *file.Tunnel, err error) {
	// P2P 端代理目标端口号
	p2pListenPort := beego.AppConfig.String("p2p_listen_port")
	t = &file.Tunnel{
		Port:      0,
		ServerIp:  "",
		Mode:      "p2p",
		Target:    &file.Target{TargetStr: p2pListenPort, LocalProxy: false},
		Id:        int(file.GetDb().JsonDb.GetTaskId()),
		Status:    true,
		Remark:    deviceKey,
		Password:  password,
		LocalPath: "",
		StripPre:  "",
		Flow:      &file.Flow{},
	}
	if !tool.TestServerPort(t.Port, t.Mode) {
		return nil, errors.New("the port cannot be opened because it may has been occupied or is no longer allowed")
	}
	if t.Client, err = file.GetDb().GetClient(clientId); err != nil {
		return nil, err
	}
	if t.Client.MaxTunnelNum != 0 && t.Client.GetTunnelNum() >= t.Client.MaxTunnelNum {
		return nil, err
	}
	if err := file.GetDb().NewTask(t); err != nil {
		return nil, errors.New(fmt.Sprintf("the number of tunnels exceeds the limit %s", err.Error()))
	}
	return
}
