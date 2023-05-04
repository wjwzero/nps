package client

import (
	"ehang.io/nps-mux"
	"ehang.io/nps/lib/crypt"
	"errors"
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
	LocalServer   []*net.TCPListener
	udpConn       net.Conn
	muxSession    *nps_mux.Mux
	fileServer    []*http.Server
	p2pNetBridge  *p2pBridge
	lock          sync.RWMutex
	udpConnStatus bool
	connStatus    string
)

type p2pBridge struct {
}

func (p2pBridge *p2pBridge) SendLinkInfo(clientId int, link *conn.Link, t *file.Tunnel) (target net.Conn, err error) {
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

func startLocalFileServer(config *config.CommonConfig, t *file.Tunnel, vkey string) {
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
	if l.Type != "secret" {
		go handleUdpMonitor(config, l)
	}
	task := &file.Tunnel{
		Port:     l.Port,
		ServerIp: "0.0.0.0",
		Status:   true,
		Client: &file.Client{
			Cnf: &file.Config{
				U:        "",
				P:        "",
				Compress: config.Client.Cnf.Compress,
			},
			Status:    true,
			RateLimit: 0,
			Flow:      &file.Flow{},
		},
		Flow:   &file.Flow{},
		Target: &file.Target{},
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
		conn.Accept(listener, func(c net.Conn) {
			logs.Trace("new %s connection", l.Type)
			// 进行局域网
			if err = handleRemoteLocalAddr(c, config, l); err != nil {
				logs.Info("LAN connect error %s", err.Error())
				if l.Type == "secret" {
					handleSecret(c, config, l)
				} else if l.Type == "p2p" {
					handleP2PVisitor(c, config, l)
				}
			}
			logs.Info("now connect Type ->>>>>>:", l.ConnStatus)
		})
	}
	return nil
}

func handleUdpMonitor(config *config.CommonConfig, l *config.LocalServer) {
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if !udpConnStatus {
				udpConn = nil
				tmpConn, err := common.GetLocalUdpAddr()
				if err != nil {
					logs.Error(err)
					return
				}
				for i := 0; i < 10; i++ {
					logs.Notice("try to connect to the server after 10s ", i+1)
					newUdpConn(tmpConn.LocalAddr().String(), config, l)
					if udpConn != nil {
						udpConnStatus = true
						break
					}
					time.Sleep(time.Duration(10) * time.Second)
				}
			}
		}
	}
}

func handleRemoteLocalAddr(localTcpConn net.Conn, config *config.CommonConfig, l *config.LocalServer) (err error) {
	err, localIp := common.GetIntranetIp()
	if err != nil {
		logs.Error(err)
		return
	}
	l.ConnStatus = "LAN"
	err = newRemoteLocalConn(localTcpConn, localIp, config, l)
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

func handleP2PVisitor(localTcpConn net.Conn, config *config.CommonConfig, l *config.LocalServer) {
	if udpConn == nil {
		logs.Notice("new conn, P2P can not penetrate successfully, traffic will be transferred through the server")
		l.ConnStatus = "relay"
		logs.Info("now connect Type :", l.ConnStatus)
		handleSecret(localTcpConn, config, l)
		return
	}
	logs.Trace("start trying to connect with the server")
	//TODO just support compress now because there is not tls file in client packages
	link := conn.NewLink(common.CONN_TCP, l.Target, false, config.Client.Cnf.Compress, localTcpConn.LocalAddr().String(), false)
	if target, err := p2pNetBridge.SendLinkInfo(0, link, nil); err != nil {
		logs.Error(err)
		udpConnStatus = false
		return
	} else {
		l.ConnStatus = "P2P"
		logs.Info("now connect Type :", l.ConnStatus)
		conn.CopyWaitGroup(target, localTcpConn, false, config.Client.Cnf.Compress, nil, nil, false, nil)
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

func newRemoteLocalConn(localTcpConn net.Conn, localAddr string, config *config.CommonConfig, l *config.LocalServer) (err error) {
	var remoteConn *conn.Conn
	remoteConn, err = NewConn(config.Tp, config.VKey, config.Server, common.WORK_P2P, config.ProxyUrl)
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
