package server

import (
	"ehang.io/nps/db"
	"ehang.io/nps/lib/version"
	"ehang.io/nps/models"
	"errors"
	"math"
	"os"
	"strconv"
	"sync"
	"time"

	"ehang.io/nps/bridge"
	"ehang.io/nps/lib/common"
	"ehang.io/nps/lib/file"
	"ehang.io/nps/server/proxy"
	"ehang.io/nps/server/tool"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

var (
	Bridge    *bridge.Bridge
	RunList   sync.Map //map[int]interface{}
	ClientDao db.ClientDao
	TaskDao   db.TaskDao
	HostDao   db.HostDao
)

func init() {
	RunList = sync.Map{}
}

// init task from db
func InitFromCsv() {
	//Add a public password
	//if vkey := beego.AppConfig.String("public_vkey"); vkey != "" {
	//	c := models.NewClientInit(vkey, true, true)
	//	ClientDao.NewClient(c)
	//	RunList.Store(c.Id, nil)
	//	//RunList[c.Id] = nil
	//}
	//Initialize services in server-side files
	//TaskDao.LoadTaskFromDB()
	//TaskDao.Tasks.Range(func(key, value interface{}) bool {
	//	if value.(*models.NpsClientTaskInfo).Status {
	//		AddTask(value.(*models.NpsClientTaskInfo))
	//	}
	//	return true
	//})
}

// get bridge command
func DealBridgeTask() {
	for {
		select {
		case t := <-Bridge.OpenTask:
			AddTask(t)
		case t := <-Bridge.CloseTask:
			StopServer(t.Id)
		case id := <-Bridge.CloseClient:
			ClientDao.UpdateStatusOffline(id)
		case tunnel := <-Bridge.OpenTask:
			StartTask(tunnel.Id)
		case s := <-Bridge.SecretChan:
			logs.Trace("New secret connection, addr", s.Conn.Conn.RemoteAddr())
			if t := TaskDao.GetTaskByMd5Password(s.Password); t != nil {
				if t.Status {
					go proxy.NewBaseServer(Bridge, t).DealClient(s.Conn, t.Client, t.TargetStr, nil, common.CONN_TCP, nil, t.Flow, t.IsLocalProxy)
				} else {
					s.Conn.Close()
					logs.Trace("This key %s cannot be processed,status is close", s.Password)
				}
			} else {
				logs.Trace("This key %s cannot be processed", s.Password)
				s.Conn.Close()
			}
		}
	}
}

// start a new server
func StartNewServer(bridgePort int, cnf *models.NpsClientTaskInfo, bridgeType string, bridgeDisconnect int) {
	Bridge = bridge.NewTunnel(bridgePort, bridgeType, common.GetBoolByStr(beego.AppConfig.String("ip_limit")), RunList, bridgeDisconnect)
	go func() {
		if err := Bridge.StartTunnel(); err != nil {
			logs.Error("start server bridge error", err)
			os.Exit(0)
		}
	}()
	if p, err := beego.AppConfig.Int("p2p_port"); err == nil {
		go proxy.NewP2PServer(p).Start()
		go proxy.NewP2PServer(p + 1).Start()
		go proxy.NewP2PServer(p + 2).Start()
	}
	go DealBridgeTask()
	go dealClientFlow()
	if svr := NewMode(Bridge, cnf); svr != nil {
		if err := svr.Start(); err != nil {
			logs.Error(err)
		}
		StartRunStatus(cnf)
		RunList.Store(cnf.Id, svr)
		//RunList[cnf.Id] = svr
	} else {
		logs.Error("Incorrect startup mode %s", cnf.Mode)
	}
}

func dealClientFlow() {
	// 客户端处理，去除定制逻辑
	/*ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			dealClientData()
		}
	}*/
}

// new a server by mode name
func NewMode(Bridge *bridge.Bridge, c *models.NpsClientTaskInfo) proxy.Service {
	var service proxy.Service
	switch c.Mode {
	case "tcp", "file":
		service = proxy.NewTunnelModeServer(proxy.ProcessTunnel, Bridge, c)
	case "socks5":
		service = proxy.NewSock5ModeServer(Bridge, c)
	case "httpProxy":
		service = proxy.NewTunnelModeServer(proxy.ProcessHttp, Bridge, c)
	case "tcpTrans":
		service = proxy.NewTunnelModeServer(proxy.HandleTrans, Bridge, c)
	case "udp":
		service = proxy.NewUdpModeServer(Bridge, c)
	case "webServer":
		InitFromCsv()
		t := &models.NpsClientTaskInfo{
			Port:   0,
			Mode:   "httpHostServer",
			Status: true,
		}
		AddTask(t)
		service = proxy.NewWebServer(Bridge)
	case "httpHostServer":
		httpPort, _ := beego.AppConfig.Int("http_proxy_port")
		httpsPort, _ := beego.AppConfig.Int("https_proxy_port")
		useCache, _ := beego.AppConfig.Bool("http_cache")
		cacheLen, _ := beego.AppConfig.Int("http_cache_length")
		addOrigin, _ := beego.AppConfig.Bool("http_add_origin_header")
		service = proxy.NewHttp(Bridge, c, httpPort, httpsPort, useCache, cacheLen, addOrigin)
	}
	return service
}

// stop server
func StopServer(id int) error {
	//if v, ok := RunList[id]; ok {
	if v, ok := RunList.Load(id); ok {
		if svr, ok := v.(proxy.Service); ok {
			if err := svr.Close(); err != nil {
				return err
			}
			logs.Info("stop server id %d", id)
		} else {
			logs.Warn("stop server id %d error", id)
		}
		if t, err := TaskDao.GetTask(id); err != nil {
			return err
		} else {
			t.Status = false
			TaskDao.UpdateTaskStatus(t)
			StopRunStatus(t)
		}
		//delete(RunList, id)
		RunList.Delete(id)
		return nil
	}
	return errors.New("task is not running")
}

func StopRunStatus(t *models.NpsClientTaskInfo) {
	t.RunStatus = false
	TaskDao.UpdateTaskRunStatus(t)
}

func StartRunStatus(t *models.NpsClientTaskInfo) {
	t.RunStatus = true
	TaskDao.UpdateTaskRunStatus(t)
}

// add task
func AddTask(t *models.NpsClientTaskInfo) error {
	if t.Mode == "secret" || t.Mode == "p2p" {
		logs.Info("secret task %s start id %d", t.Remark, t.Id)
		//RunList[t.Id] = nil
		StartRunStatus(t)
		RunList.Store(t.Id, nil)
		return nil
	}
	if b := tool.TestServerPort(t.Port, t.Mode); !b && t.Mode != "httpHostServer" {
		logs.Error("taskId %d start error port %d open failed", t.Id, t.Port)
		return errors.New("the port open error")
	}
	if svr := NewMode(Bridge, t); svr != nil {
		logs.Info("tunnel task %s start mode：%s port %d", t.Remark, t.Mode, t.Port)
		//RunList[t.Id] = svr
		StartRunStatus(t)
		RunList.Store(t.Id, svr)
		go func() {
			if err := svr.Start(); err != nil {
				logs.Error("clientId %d taskId %d start error %s", t.Client.Id, t.Id, err)
				//delete(RunList, t.Id)
				StopRunStatus(t)
				RunList.Delete(t.Id)
				return
			}
		}()
	} else {
		return errors.New("the mode is not correct")
	}
	return nil
}

// start task
func StartTask(id int) error {
	if t, err := TaskDao.GetTask(id); err != nil {
		return err
	} else {
		AddTask(t)
		t.Status = true
		TaskDao.UpdateTaskStatus(t)
	}
	return nil
}

// delete task
func DelTask(id int) error {
	//if _, ok := RunList[id]; ok {
	if _, ok := RunList.Load(id); ok {
		if err := StopServer(id); err != nil {
			return err
		}
	}
	return TaskDao.DelTask(id)
}

func dealClientData() {
	// 因遍历数据而禁用
	// 此处处理删除，减少查询与数据变更
	//ClientDao.LoadClientFromDb()
	//ClientDao.Clients.Range(func(key, value interface{}) bool {
	//	v := value.(*models.NpsClientListInfo)
	//	if vv, ok := Bridge.Client.Load(v.Id); ok {
	//		v.IsConnect = true
	//		v.Version = vv.(*bridge.Client).Version
	//	} else {
	//		v.IsConnect = false
	//	}
	//	return true
	//})
	//return
}

// close the client
func DelClientConnect(clientId int) {
	Bridge.DelClient(clientId)
}

func GetDashboardData() map[string]interface{} {
	data := make(map[string]interface{})
	data["version"] = version.VERSION
	//HostDao.LoadHostFromDb()
	data["hostCount"] = common.GeSynctMapLen(HostDao.Hosts)
	//ClientDao.LoadClientFromDb()
	data["clientCount"] = common.GeSynctMapLen(ClientDao.Clients)
	dealClientData()
	c := 0
	var in, out int64
	//ClientDao.LoadClientFromDb()
	//ClientDao.Clients.Range(func(key, value interface{}) bool {
	//	v := value.(*models.NpsClientListInfo)
	//	if v.IsConnect {
	//		c += 1
	//	}
	//	in += v.FlowInlet
	//	out += v.FlowExport
	//	return true
	//})
	data["clientOnlineCount"] = c
	data["inletFlowCount"] = int(in)
	data["exportFlowCount"] = int(out)
	var tcp, udp, secret, socks5, p2p, http int
	//TaskDao.LoadTaskFromDB()
	//TaskDao.Tasks.Range(func(key, value interface{}) bool {
	//	switch value.(*models.NpsClientTaskInfo).Mode {
	//	case "tcp":
	//		tcp += 1
	//	case "socks5":
	//		socks5 += 1
	//	case "httpProxy":
	//		http += 1
	//	case "udp":
	//		udp += 1
	//	case "p2p":
	//		p2p += 1
	//	case "secret":
	//		secret += 1
	//	}
	//	return true
	//})

	data["tcpC"] = tcp
	data["udpCount"] = udp
	data["socks5Count"] = socks5
	data["httpProxyCount"] = http
	data["secretCount"] = secret
	data["p2pCount"] = p2p
	data["bridgeType"] = beego.AppConfig.String("bridge_type")
	data["httpProxyPort"] = beego.AppConfig.String("http_proxy_port")
	data["httpsProxyPort"] = beego.AppConfig.String("https_proxy_port")
	data["ipLimit"] = beego.AppConfig.String("ip_limit")
	data["flowStoreInterval"] = beego.AppConfig.String("flow_store_interval")
	data["serverIp"] = beego.AppConfig.String("p2p_ip")
	data["p2pPort"] = beego.AppConfig.String("p2p_port")
	data["logLevel"] = beego.AppConfig.String("log_level")
	tcpCount := 0
	//ClientDao.LoadClientFromDb()
	//ClientDao.Clients.Range(func(key, value interface{}) bool {
	//	tcpCount += int(value.(*models.NpsClientListInfo).NowConnectNum)
	//	return true
	//})
	data["tcpCount"] = tcpCount
	cpuPercet, _ := cpu.Percent(0, true)
	var cpuAll float64
	for _, v := range cpuPercet {
		cpuAll += v
	}
	loads, _ := load.Avg()
	data["load"] = loads.String()
	data["cpu"] = math.Round(cpuAll / float64(len(cpuPercet)))
	swap, _ := mem.SwapMemory()
	data["swap_mem"] = math.Round(swap.UsedPercent)
	vir, _ := mem.VirtualMemory()
	data["virtual_mem"] = math.Round(vir.UsedPercent)
	conn, _ := net.ProtoCounters(nil)
	io1, _ := net.IOCounters(false)
	time.Sleep(time.Millisecond * 500)
	io2, _ := net.IOCounters(false)
	if len(io2) > 0 && len(io1) > 0 {
		data["io_send"] = (io2[0].BytesSent - io1[0].BytesSent) * 2
		data["io_recv"] = (io2[0].BytesRecv - io1[0].BytesRecv) * 2
	}
	for _, v := range conn {
		data[v.Protocol] = v.Stats["CurrEstab"]
	}
	//chart
	var fg int
	if len(tool.ServerStatus) >= 10 {
		fg = len(tool.ServerStatus) / 10
		for i := 0; i <= 9; i++ {
			data["sys"+strconv.Itoa(i+1)] = tool.ServerStatus[i*fg]
		}
	}
	return data
}

// get tunnel list
func GetTunnelList() (list []*file.TunnelTypesProductRelation, cnt int) {
	list, cnt = file.GetDb().GetTunnelList()
	return
}

func GetConfigList() (list []*file.DynamicConfig, cnt int) {
	list, cnt = file.GetDb().GetDynamicConfigList()
	return
}
