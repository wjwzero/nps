package client

import (
	"ehang.io/nps-mux"
	"ehang.io/nps/lib/crypt"
	"ehang.io/nps/models"
	"errors"
	"math/rand"
	"net"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	"ehang.io/nps/lib/common"
	"ehang.io/nps/lib/config"
	"ehang.io/nps/lib/conn"
	"ehang.io/nps/lib/file"
	"ehang.io/nps/server/proxy"
	"github.com/astaxie/beego/logs"
	"github.com/xtaci/kcp-go"
)

var (
	LocalServer     []*net.TCPListener
	udpConn         net.Conn
	muxSession      *nps_mux.Mux
	fileServer      []*http.Server
	p2pNetBridge    *p2pBridge
	lock            sync.RWMutex
	udpConnStatus   bool
	connStatus      string
	tunneltypesTemp string
)

type p2pBridge struct {
}

func (p2pBridge *p2pBridge) SendLinkInfo(clientId int, link *conn.Link, t *models.NpsClientTaskInfo) (target net.Conn, err error) {
	for i := 0; muxSession == nil; i++ {
		if i >= 20 {
			err = errors.New("p2pBridge:too many times to get muxSession")
			logs.Error(err)
			return
		}
		runtime.Gosched() // waiting for another goroutine establish the mux connection
	}
	nowConn, err := muxSession.NewConn()
	if err != nil {
		udpConn = nil
		return nil, err
	}
	if _, err := conn.NewConn(nowConn).SendInfo(link, ""); err != nil {
		udpConnStatus = false
		return nil, err
	}
	return nowConn, nil
}

func CloseLocalServer() {
	for _, v := range LocalServer {
		v.Close()
	}
	for _, v := range fileServer {
		v.Close()
	}
}

func startLocalFileServer(config *config.CommonConfig, t *models.NpsClientTaskInfo, vkey string) {
	remoteConn, err := NewConn(config.Tp, vkey, config.Server, common.WORK_FILE, config.ProxyUrl)
	if err != nil {
		logs.Error("Local connection server failed ", err.Error())
		return
	}
	srv := &http.Server{
		Handler: http.StripPrefix(t.StripPre, http.FileServer(http.Dir(t.LocalPath))),
	}
	logs.Info("start local file system, local path %s, strip prefix %s ,remote port %s ", t.LocalPath, t.StripPre, t.Ports)
	fileServer = append(fileServer, srv)
	listener := nps_mux.NewMux(remoteConn.Conn, common.CONN_TCP, config.DisconnectTime)
	logs.Error(srv.Serve(listener))
}

func StartLocalServer(l *config.LocalServer, config *config.CommonConfig) error {
	var tunnelTypes string
	var typeErr error
	// 获取隧道类型
	if tunnelTypes, typeErr = getTunnelTypes(config, l); typeErr != nil {
		tunnelTypes = common.DEFULT_TUNNEL_TYPE
		logs.Warn("未从平台获取到 tunnelTypes 采用默认 %s", tunnelTypes)
	}
	logs.Info("从平台获取到 tunnelTypes 采用 %s", tunnelTypes)
	if strings.Contains(tunnelTypes, common.P2P_TUNNEL_TYPE) {
		logs.Info("TunnelTypes includes P2P, which is used to attempt UDP interaction.")
		go handleUdpMonitor(config, l)
	}
	task := &models.NpsClientTaskInfo{
		Port:     l.Port,
		ServerIp: "0.0.0.0",
		Status:   true,
		Client:   &models.NpsClientListInfo{},
		Flow:     &file.Flow{},
	}
	switch l.Type {
	case "p2ps":
		logs.Info("successful start-up of local socks5 monitoring, port", l.Port)
		return proxy.NewSock5ModeServer(p2pNetBridge, task).Start()
	case "p2pt":
		logs.Info("successful start-up of local tcp trans monitoring, port", l.Port)
		return proxy.NewTunnelModeServer(proxy.HandleTrans, p2pNetBridge, task).Start()
	case "p2p", "secret":
		listener, err := net.ListenTCP("tcp", &net.TCPAddr{net.ParseIP("0.0.0.0"), l.Port, ""})
		if err != nil {
			logs.Error("local listener startup failed port %d, error %s", l.Port, err.Error())
			return err
		}
		LocalServer = append(LocalServer, listener)
		logs.Info("successful start-up of local tcp monitoring, port", l.Port)
	retry:
		conn.Accept(listener, func(c net.Conn) {
			// 获取隧道类型
			logs.Trace("new %s connection", tunnelTypes)
			var p2perr error
			// 隧道类型包含 P2P
			if strings.Contains(tunnelTypes, common.P2P_TUNNEL_TYPE) {
				l.ConnStatus = common.P2P_TUNNEL_TYPE
				p2perr = handleP2PVisitor(c, config, l)
			}
			// 隧道包含Realy 或 隧道类型包含Relay 且 P2P失败
			if strings.Contains(tunnelTypes, common.RELAY_TUNNEL_TYPE) || (strings.Contains(tunnelTypes, common.RELAY_TUNNEL_TYPE) && p2perr != nil) {
				l.ConnStatus = common.RELAY_TUNNEL_TYPE
				handleSecret(c, config, l)
			}
			logs.Info("now connect Type ->>>>>>:", l.ConnStatus)
		})
		goto retry
	}
	return nil
}

func switchTunnel(c net.Conn, tunnelTypeArr []string, l *config.LocalServer, config *config.CommonConfig) bool {
	for _, tunnelType := range tunnelTypeArr {
		switch tunnelType {
		case common.LAN_TUNNEL_TYPE:
			l.ConnStatus = common.LAN_TUNNEL_TYPE
			logs.Info("start connect Type :", l.ConnStatus)
			if err := handleRemoteLocalAddr(c, config, l); err != nil {
				logs.Info("end connect Type :%s e: %s", l.ConnStatus, err.Error())
				return true
			}
		case common.P2P_TUNNEL_TYPE:
			l.ConnStatus = common.P2P_TUNNEL_TYPE
			logs.Info("start connect Type :", l.ConnStatus)
			if err := handleP2PVisitor(c, config, l); err != nil {
				logs.Info("end connect Type :%s e: %s", l.ConnStatus, err.Error())
				return true
			}
		case common.RELAY_TUNNEL_TYPE:
			l.ConnStatus = common.RELAY_TUNNEL_TYPE
			logs.Info("start connect Type :", l.ConnStatus)
			handleSecret(c, config, l)
		}
	}
	return false
}

func handleUdpMonitor(config *config.CommonConfig, l *config.LocalServer) {
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()
	for {
		var retryNum = 0
		select {
		case <-ticker.C:
			if !udpConnStatus {
				udpConn = nil
				tmpConn, err := common.GetLocalUdpAddr()
				if err != nil {
					logs.Error(err)
					return
				}
				var retryDuration = 10
				for i := 0; i < 10; i++ {
					retryNum++
					if retryNum < 3 {
						retryDuration = 10 + rand.Intn(5)
					} else if retryNum >= 3 && retryNum < 5 {
						retryDuration = 15 + rand.Intn(5)
					} else if retryNum >= 5 && retryNum < 10 {
						retryDuration = 30 + rand.Intn(5)
					} else {
						retryDuration = 9*6 + rand.Intn(6)
					}
					logs.Notice("try to connect to the server after %d s, time %d ", retryDuration, i+1)
					newUdpConn(tmpConn.LocalAddr().String(), config, l)
					if udpConn != nil {
						udpConnStatus = true
						break
					}
					time.Sleep(time.Duration(retryDuration) * time.Second)
				}
			}
		}
	}
}

func handleRemoteLocalAddr(localTcpConn net.Conn, config *config.CommonConfig, l *config.LocalServer) (err error) {
	tmpConn, err := common.GetLocalUdpAddr()
	if err != nil {
		logs.Error(err)
		return
	}
	//l.ConnStatus = "LAN"
	err = newRemoteLocalConn(localTcpConn, tmpConn.LocalAddr().String(), config, l)
	return
}

func handleSecret(localTcpConn net.Conn, config *config.CommonConfig, l *config.LocalServer) {
	remoteConn, err := NewConn(config.Tp, config.VKey, config.Server, common.WORK_SECRET, config.ProxyUrl)
	if err != nil {
		logs.Error("Local connection server failed ", err.Error())
		return
	}
	// common.WORK_SECRET 单纯校验MD5
	if err := remoteConn.WriteLenContent([]byte(crypt.Md5(l.Password[0 : len(l.Password)-4]))); err != nil {
		logs.Error("Local connection server failed ", err.Error())
		return
	}
	conn.CopyWaitGroup(remoteConn.Conn, localTcpConn, false, false, nil, nil, false, nil)
}

func handleP2PVisitor(localTcpConn net.Conn, config *config.CommonConfig, l *config.LocalServer) (err error) {
	if udpConn == nil {
		logs.Notice("new conn, P2P can not penetrate successfully, traffic will be transferred through the server")
		//l.ConnStatus = "relay"
		err = errors.New("P2P 未打通, 如配置 relay 将尝试relay")
		logs.Warn(err.Error())
		//handleSecret(localTcpConn, config, l)
		return
	}
	logs.Trace("start trying to connect with the server")
	//TODO just support compress now because there is not tls file in client packages
	link := conn.NewLink(common.CONN_TCP, l.Target, false, config.Client.IsCompress, localTcpConn.LocalAddr().String(), false)
	if target, bErr := p2pNetBridge.SendLinkInfo(0, link, nil); bErr != nil {
		logs.Error(bErr)
		udpConnStatus = false
		return bErr
	} else {
		//l.ConnStatus = "P2P"
		conn.CopyWaitGroup(target, localTcpConn, false, config.Client.IsCompress, nil, nil, false, nil)
		return nil
	}
}

func newUdpConn(localAddr string, config *config.CommonConfig, l *config.LocalServer) {
	lock.Lock()
	defer lock.Unlock()
	remoteConn, err := NewConn(config.Tp, config.VKey, config.Server, common.WORK_P2P, config.ProxyUrl)
	if err != nil {
		logs.Error("Local connection server failed ", err.Error())
		return
	}
	// common.WORK_P2P 因预创建，需要对称加密
	if err := remoteConn.WriteLenContent([]byte(common.GetAesEnVerifyval(l.Password))); err != nil {
		logs.Error("Local connection server failed ", err.Error())
		return
	}
	var rAddr []byte
	//读取服务端地址、密钥 继续做处理
	if rAddr, err = remoteConn.GetShortLenContent(); err != nil {
		logs.Error(err)
		return
	}
	if strings.HasPrefix(string(rAddr), "{[checked]}") {
		logs.Error(string(rAddr))
		return
	}
	var localConn net.PacketConn
	var remoteAddress string
	if remoteAddress, localConn, err = handleP2PUdp(localAddr, string(rAddr), common.GetAesEnVerifyval(l.Password), common.WORK_P2P_VISITOR); err != nil {
		logs.Error(err)
		return
	}
	udpTunnel, err := kcp.NewConn(remoteAddress, nil, 150, 3, localConn)
	if err != nil || udpTunnel == nil {
		logs.Warn(err)
		return
	}
	logs.Trace("successful create a connection with server", remoteAddress)
	conn.SetUdpSession(udpTunnel)
	udpConn = udpTunnel
	muxSession = nps_mux.NewMux(udpConn, "kcp", config.DisconnectTime)
	p2pNetBridge = &p2pBridge{}
}

func getTunnelTypes(config *config.CommonConfig, l *config.LocalServer) (tunnelTypes string, err error) {
	if tunneltypesTemp != "" {
		logs.Debug("use tunnelTypes_temp : %s", tunneltypesTemp)
		tunnelTypes = tunneltypesTemp
		return
	}
	var remoteConn *conn.Conn
	remoteConn, err = NewConn(config.Tp, config.VKey, config.Server, common.WORK_LAN, config.ProxyUrl)
	if err != nil {
		logs.Error("Local connection server failed ", err.Error())
		return
	}
	// common.WORK_P2P 因预创建，需要对称加密
	if err = remoteConn.WriteLenContent([]byte(common.GetAesEnVerifyval(l.Password))); err != nil {
		logs.Error("Local connection server failed ", err.Error())
		return
	}
	var rAddr []byte
	//读取服务端地址、密钥 继续做处理
	if rAddr, err = remoteConn.GetShortLenContent(); err != nil {
		logs.Error(err)
		return
	}
	if strings.HasPrefix(string(rAddr), "{[checked]}") {
		err = errors.New(string(rAddr))
		logs.Error(string(rAddr))
		return
	}
	// 获得产品信息pk
	var tunnelByte []byte
	if tunnelByte, err = remoteConn.GetShortLenContent(); err != nil {
		logs.Error(err)
		return
	}
	tunnelTypes = string(tunnelByte)
	tunneltypesTemp = tunnelTypes
	return
}

func newRemoteLocalConn(localTcpConn net.Conn, localAddr string, config *config.CommonConfig, l *config.LocalServer) (err error) {
	var remoteConn *conn.Conn
	remoteConn, err = NewConn(config.Tp, config.VKey, config.Server, common.WORK_LAN, config.ProxyUrl)
	if err != nil {
		logs.Error("Local connection server failed ", err.Error())
		return
	}
	// common.WORK_P2P 因预创建，需要对称加密
	if err = remoteConn.WriteLenContent([]byte(common.GetAesEnVerifyval(l.Password))); err != nil {
		logs.Error("Local connection server failed ", err.Error())
		return
	}
	var rAddr []byte
	//读取服务端地址、密钥 继续做处理
	if rAddr, err = remoteConn.GetShortLenContent(); err != nil {
		logs.Error(err)
		return
	}
	if strings.HasPrefix(string(rAddr), "{[checked]}") {
		logs.Error(string(rAddr))
		return
	}
	var remoteAddress string
	if remoteAddress, err = handleRemoteLocalTcp(localAddr, string(rAddr), common.GetAesEnVerifyval(l.Password), common.WORK_P2P_VISITOR, l.Target); err != nil {
		logs.Error(err)
		return
	}
	arr := strings.Split(remoteAddress, ":")
	logs.Info("局域网 Remote Local connection server ", remoteAddress)
	remoteLocalConn, err := NewRemoteLocalConn(arr[0] + ":" + l.Target)
	if err != nil {
		logs.Error("Local connection server failed ", err.Error())
		return
	}
	conn.CopyWaitGroup(remoteLocalConn.Conn, localTcpConn, false, false, nil, nil, false, nil)
	return
}

func UdpConnStatus() bool {
	return udpConnStatus
}

func ConnStatus() string {
	return connStatus
}

func GetLanAddr(l *config.LocalServer, config *config.CommonConfig) (localIp string, err error) {
	tmpConn, err := common.GetLocalUdpAddr()
	if err != nil {
		logs.Error(err)
		return
	}
	localAddr := tmpConn.LocalAddr().String()
	var remoteConn *conn.Conn
	remoteConn, err = NewConn(config.Tp, config.VKey, config.Server, common.WORK_LAN, config.ProxyUrl)
	if err != nil {
		logs.Error("Local connection server failed ", err.Error())
		return
	}
	// common.WORK_P2P 因预创建，需要对称加密
	if err = remoteConn.WriteLenContent([]byte(common.GetAesEnVerifyval(l.Password))); err != nil {
		logs.Error("Local connection server failed ", err.Error())
		return
	}
	var rAddr []byte
	//读取服务端地址、密钥 继续做处理
	if rAddr, err = remoteConn.GetShortLenContent(); err != nil {
		logs.Error(err)
		return
	}
	if strings.HasPrefix(string(rAddr), "{[checked]}") {
		logs.Error(string(rAddr))
		err = errors.New("device offline")
		return
	}
	if localIp, err = handleRemoteLocalTcp(localAddr, string(rAddr), common.GetAesEnVerifyval(l.Password), common.WORK_P2P_VISITOR, l.Target); err != nil {
		logs.Error(err)
		return
	}
	logs.Info("局域网 Remote Local connection server ", localIp)
	if err != nil {
		logs.Error("Local connection server failed ", err.Error())
		return
	}
	return
}
