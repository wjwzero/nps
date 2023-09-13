package npc

import (
	"ehang.io/nps/client"
	"ehang.io/nps/lib/common"
	"ehang.io/nps/lib/config"
	"ehang.io/nps/models"
	"flag"
	"github.com/astaxie/beego/logs"
	"github.com/kardianos/service"
	"os"
	"runtime"
)

var (
	serverAddr     = flag.String("server", "", "Server addr (ip:port)")
	configPath     = flag.String("config", "", "Configuration file path")
	verifyKey      = flag.String("vkey", "", "Authentication key")
	logType        = flag.String("log", "stdout", "Log output mode（stdout|file）")
	connType       = flag.String("type", "tcp", "Connection type with the server（kcp|tcp）")
	proxyUrl       = flag.String("proxy", "", "proxy socks5 url(eg:socks5://111:222@127.0.0.1:9007)")
	logLevel       = flag.String("log_level", "7", "log level 0~7")
	registerTime   = flag.Int("time", 2, "register time long /h")
	localPort      = flag.Int("local_port", 52000, "p2p local port")
	password       = flag.String("password", "", "p2p password flag")
	target         = flag.String("target", "", "p2p target")
	localType      = flag.String("local_type", "p2p", "p2p target")
	logPath        = flag.String("log_path", "", "npc log path")
	debug          = flag.Bool("debug", true, "npc debug")
	pprofAddr      = flag.String("pprof", "", "PProf debug addr (ip:port)")
	stunAddr       = flag.String("stun_addr", "stun.stunprotocol.org:3478", "stun server address (eg:stun.stunprotocol.org:3478)")
	ver            = flag.Bool("version", false, "show current version")
	disconnectTime = flag.Int("disconnect_timeout", 60, "not receiving check packet times, until timeout will disconnect the client")
	localServer    *config.LocalServer
)

type npc struct {
	exit chan struct{}
}

func (p *npc) StartP2P(serverAddrParam string, verifyKeyParam string, passwordParam string, targetParam string, localTypeParam string, localPortParam int) error {
	*serverAddr = serverAddrParam
	*verifyKey = verifyKeyParam
	*password = passwordParam
	*target = targetParam
	*localType = localTypeParam
	*localPort = localPortParam
	go p.run()
	return nil
}

func (p *npc) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *npc) Stop(s service.Service) error {
	close(p.exit)
	if service.Interactive() {
		os.Exit(0)
	}
	return nil
}

func (p *npc) run() error {
	defer func() {
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			logs.Warning("npc: panic serving %v: %v\n%s", err, string(buf))
		}
	}()
	run()
	select {
	case <-p.exit:
		logs.Warning("stop...")
	}
	return nil
}

func (p *npc) connStatus() string {
	return localServer.ConnStatus
}

func run() {
	common.InitPProfFromArg(*pprofAddr)
	//p2p or secret command
	commonConfig := new(config.CommonConfig)
	commonConfig.Server = *serverAddr
	commonConfig.VKey = *verifyKey
	commonConfig.Tp = *connType
	localServer = new(config.LocalServer)
	localServer.Type = *localType
	localServer.Password = *password
	localServer.Target = *target
	localServer.Port = *localPort
	commonConfig.Client = new(models.NpsClientInfo)
	go client.StartLocalServer(localServer, commonConfig)
	return
}

func (p *npc) GetLan(serverAddrParam string, verifyKeyParam string, passwordParam string, targetParam string, localTypeParam string) (localIp string, err error) {
	*serverAddr = serverAddrParam
	*verifyKey = verifyKeyParam
	*password = passwordParam
	*target = targetParam
	*localType = localTypeParam
	//p2p or secret command
	commonConfig := new(config.CommonConfig)
	commonConfig.Server = *serverAddr
	commonConfig.VKey = *verifyKey
	commonConfig.Tp = *connType
	localServer = new(config.LocalServer)
	localServer.Type = *localType
	localServer.Password = *password
	localServer.Target = *target
	localServer.Port = *localPort
	commonConfig.Client = new(models.NpsClientInfo)
	localIp, err = client.GetLanAddr(localServer, commonConfig)
	return
}
