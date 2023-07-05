package proxy

import (
	"ehang.io/nps/db"
	"ehang.io/nps/models"
	"errors"
	"net"
	"net/http"
	"sync"

	"ehang.io/nps/bridge"
	"ehang.io/nps/lib/common"
	"ehang.io/nps/lib/conn"
	"ehang.io/nps/lib/file"
	"github.com/astaxie/beego/logs"
)

type Service interface {
	Start() error
	Close() error
}

type NetBridge interface {
	SendLinkInfo(clientId int, link *conn.Link, t *models.NpsClientTaskInfo) (target net.Conn, err error)
}

// BaseServer struct
type BaseServer struct {
	id           int
	bridge       NetBridge
	task         *models.NpsClientTaskInfo
	errorContent []byte
	sync.Mutex
	db.ClientStatisticDao
}

func NewBaseServer(bridge *bridge.Bridge, task *models.NpsClientTaskInfo) *BaseServer {
	return &BaseServer{
		bridge:       bridge,
		task:         task,
		errorContent: nil,
		Mutex:        sync.Mutex{},
	}
}

// add the flow
func (s *BaseServer) FlowAdd(in, out int64) {
	s.Lock()
	defer s.Unlock()
	s.task.Flow.ExportFlow += out
	s.task.Flow.InletFlow += in
}

// change the flow
func (s *BaseServer) FlowAddHost(host *file.Host, in, out int64) {
	s.Lock()
	defer s.Unlock()
	host.Flow.ExportFlow += out
	host.Flow.InletFlow += in
}

// write fail bytes to the connection
func (s *BaseServer) writeConnFail(c net.Conn) {
	c.Write([]byte(common.ConnectionFailBytes))
	c.Write(s.errorContent)
}

// auth check
func (s *BaseServer) auth(r *http.Request, c *conn.Conn, u, p string) error {
	if u != "" && p != "" && !common.CheckAuth(r, u, p) {
		c.Write([]byte(common.UnauthorizedBytes))
		c.Close()
		return errors.New("401 Unauthorized")
	}
	return nil
}

// check flow limit of the client ,and decrease the allow num of client
func (s *BaseServer) CheckFlowAndConnNum(client *models.NpsClientListInfo) error {
	if client.FlowLimit > 0 && (client.FlowLimit<<20) < (client.FlowExport+client.FlowInlet) {
		return errors.New("Traffic exceeded")
	}
	if !client.GetConn() {
		return errors.New("Connections exceed the current client limit")
	}
	s.UpdateConnectNum(client.Id, client.NowConnectNum)
	return nil
}

// create a new connection and start bytes copying
func (s *BaseServer) DealClient(c *conn.Conn, client *models.NpsClientListInfo, addr string, rb []byte, tp string, f func(), flow *file.Flow, localProxy bool) error {
	link := conn.NewLink(tp, addr, client.IsCrypt, client.IsCompress, c.Conn.RemoteAddr().String(), localProxy)
	if target, err := s.bridge.SendLinkInfo(client.Id, link, s.task); err != nil {
		logs.Warn("get connection from client id %d  error %s", client.Id, err.Error())
		c.Close()
		return err
	} else {
		if f != nil {
			f()
		}
		conn.CopyWaitGroup(target, c.Conn, link.Crypt, link.Compress, client.Rate, flow, true, rb)
	}
	return nil
}
