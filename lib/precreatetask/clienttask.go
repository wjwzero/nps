package precreatetask

import (
	preclient "ehang.io/nps/lib/precreate"
	"ehang.io/nps/server"
)

func P2pClient(clientId int, deviceKey string, password string) (err error) {
	t, err := preclient.P2pClient(clientId, deviceKey, password)
	if err != nil {
		return err
	}
	if err := server.AddTask(t); err != nil {
		return err
	}
	return
}
