package preclient

import (
	"ehang.io/nps/db"
	"ehang.io/nps/lib/file"
	"ehang.io/nps/models"
	"ehang.io/nps/server/tool"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
)

var (
	ClientDao db.ClientDao
	TaskDao   db.TaskDao
	HostDao   db.HostDao
)

func P2pClient(clientId int, deviceKey string, password string) (t *models.NpsClientTaskInfo, err error) {
	// P2P 端代理目标端口号
	p2pListenPort := beego.AppConfig.String("p2p_listen_port")
	t = &models.NpsClientTaskInfo{
		RunStatus:    true,
		Port:         0,
		ServerIp:     "",
		Mode:         "p2p",
		TargetStr:    p2pListenPort,
		IsLocalProxy: false,
		Status:       true,
		Remark:       deviceKey,
		Password:     password,
		LocalPath:    "",
		StripPre:     "",
		Flow:         &file.Flow{},
	}
	if !tool.TestServerPort(t.Port, t.Mode) {
		return nil, errors.New("the port cannot be opened because it may has been occupied or is no longer allowed")
	}
	if t.Client, err = ClientDao.GetClientInfo(clientId); err != nil {
		return nil, err
	}
	if t.Client.MaxChannelNum != 0 && TaskDao.GetTunnelNum(t.Client.Id) >= t.Client.MaxChannelNum {
		return nil, err
	}
	if err := TaskDao.NewTask(t); err != nil {
		return nil, errors.New(fmt.Sprintf("the number of tunnels exceeds the limit %s", err.Error()))
	}
	return
}
