package npc

import (
	"C"
	"ehang.io/nps/client"
	"ehang.io/nps/lib/cloud"
	"ehang.io/nps/lib/common"
	"ehang.io/nps/lib/crypt"
	"ehang.io/nps/lib/version"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/validation"
	"github.com/ccding/go-stun/stun"
)

var cl *client.TRPClient

var npclient *npc

var serverIp string

//export StartClientByVerifyKey
func StartClientByVerifyKey(serverAddr, verifyKey, connType, proxyUrl *C.char) int {
	_ = logs.SetLogger("store")
	if cl != nil {
		cl.Close()
	}
	cl = client.NewRPClient(C.GoString(serverAddr), C.GoString(verifyKey), C.GoString(connType), C.GoString(proxyUrl), nil, 60)
	cl.Start()
	return 1
}

//export StartP2PClient
func StartP2PClient(cloudAddr string, verifyKey string, password string, port int) int {
	npclient = &npc{
		exit: make(chan struct{}),
	}
	valid := validation.Validation{}

	if v := valid.Required(verifyKey, "verifyKey"); !v.Ok {
		logs.Error(v.Error.Key, v.Error.Message)
		return 0
	}
	if v := valid.Required(password, "password"); !v.Ok {
		logs.Error(v.Error.Key, v.Error.Message)
		return 0
	}
	oriVerifyKey, err := crypt.VerifySign(verifyKey)
	if err != nil {
		logs.Error("deviceKey sign code error ", err)
		return 0
	}
	// 因SDK输入verifyKey 与 password是加签名后的，所以用源vkey查找IP
	var serverAddrStr string
	if serverIp == "" {
		serverIp, err = cloud.GetNpsNodeExternalIp(cloudAddr, oriVerifyKey)
		logs.Debug("begin get addr")
	}
	serverAddrStr = serverIp + ":8024"
	logs.Info("client连接地址:", serverIp)
	err = npclient.StartP2P(serverAddrStr, verifyKey, password, "5212", "p2p", port)
	if err != nil {
		logs.Error("p2p client connect error", err)
	}
	return 1
}

//export GetLanAddr
func GetLanAddr(cloudAddr string, verifyKey string, password string) string {
	npclient = &npc{
		exit: make(chan struct{}),
	}
	valid := validation.Validation{}

	if v := valid.Required(verifyKey, "verifyKey"); !v.Ok {
		logs.Error(v.Error.Key, v.Error.Message)
		return "0"
	}
	if v := valid.Required(password, "password"); !v.Ok {
		logs.Error(v.Error.Key, v.Error.Message)
		return "0"
	}
	oriVerifyKey, err := crypt.VerifySign(verifyKey)
	if err != nil {
		logs.Error("deviceKey sign code error ", err)
		return "0"
	}
	// 因SDK输入verifyKey 与 password是加签名后的，所以用源vkey查找IP
	var serverAddrStr string
	if serverIp == "" {
		serverIp, err = cloud.GetNpsNodeExternalIp(cloudAddr, oriVerifyKey)
		logs.Debug("begin get addr")
	}
	serverAddrStr = serverIp + ":8024"
	logs.Info("client连接地址:", serverIp)
	var lanAddr string
	lanAddr, err = npclient.GetLan(serverAddrStr, verifyKey, password, "5212", "p2p")
	if err != nil {
		logs.Error("client LAN connect error: ", err)
		return "0"
	}
	return lanAddr
}

//export StopP2P
func StopP2P() int {
	err := npclient.Stop(nil)
	if err != nil {
		return 0
	}
	return 1
}

//export P2PStatus
func P2PStatus() bool {
	return client.UdpConnStatus()
}

//export ConnectStatus
func ConnectStatus() string {
	return npclient.connStatus()
}

//export NatInfo
func NatInfo(stunAddr string) string {
	valid := validation.Validation{}
	if v := valid.Required(stunAddr, "stunAddr"); !v.Ok {
		logs.Info("no stunAddr, use default stunAddr")
		stunAddr = "stun.stunprotocol.org:3478"
	}
	c := stun.NewClient()
	c.SetServerAddr(stunAddr)
	nat, host, err := c.Discover()
	if err != nil || host == nil {
		logs.Error("get nat type error", err)
		return "get nat type error" + err.Error()
	}
	return fmt.Sprintf("nat type: %s \npublic address: %s\n", nat.String(), host.String())
}

//export GetClientStatus
func GetClientStatus() int {
	return client.NowStatus
}

//export CloseClient
func CloseClient() {
	if cl != nil {
		cl.Close()
	}
}

//export Version
func Version() *C.char {
	return C.CString(version.VERSION)
}

//export Logs
func Logs() *C.char {
	return C.CString(common.GetLogMsg())
}

//export Greetings
func Greetings(name string) string {
	return fmt.Sprintf("Hello %s", name)
}

//export Ctest
func Ctest() *C.char {
	logs.Info("Hello %s", "C")
	return C.CString("a")
}

func main() {
	// Need a main function to make CGO compile package as C shared library
}
