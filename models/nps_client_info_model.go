package models

import (
	"ehang.io/nps/lib/rate"
	"sync/atomic"
	"time"

	"xorm.io/xorm"
)

// NpsClientInfoModel represents a nps_client_info model.
type NpsClientInfoModel struct {
	engine xorm.EngineInterface
}

// NpsClientInfo represents a nps_client_info struct data.
type NpsClientInfo struct {
	Id                int        `xorm:"pk autoincr 'id'" json:"Id"`                      // 主键
	VerifyKey         string     `xorm:"'verify_key'" json:"VerifyKey"`                   // 唯一验证key
	Addr              string     `xorm:"'addr'" json:"Addr"`                              // ip地址
	BasicAuthUser     string     `xorm:"'basic_auth_user'" json:"BasicAuthUser"`          // Basic 认证用户名
	BasicAuthPass     string     `xorm:"'basic_auth_pass'" json:"BasicAuthPass"`          // Basic 认证密码
	ProductKey        string     `xorm:"'product_key'" json:"ProductKey"`                 //产品Key
	DeviceKey         string     `xorm:"'device_key'" json:"DeviceKey"`                   // 设备Key
	Version           string     `xorm:"'version'" json:"Version"`                        // 版本
	Status            bool       `xorm:"'status'" json:"Status"`                          // 状态 1:开放 0:关闭
	Remark            string     `xorm:"'remark'" json:"Remark"`                          // 备注
	IsConnect         bool       `xorm:"'is_connect'" json:"IsConnect"`                   // 是否连接 1:在线 0:离线
	IsConfigConnAllow bool       `xorm:"'is_config_conn_allow'" json:"IsConfigConnAllow"` // 是否允许客户端通过配置文件连接 1:是 0:否
	IsCompress        bool       `xorm:"'is_compress'" json:"IsCompress"`                 // 是否压缩 1:是 0:否
	IsCrypt           bool       `xorm:"'is_crypt'" json:"IsCrypt"`                       // 是否加密 1:是 0:否
	NoDisplay         bool       `xorm:"'no_display'" json:"NoDisplay"`                   // 在web页面是否不显示 1:是 0:否
	NoStore           bool       `xorm:"'no_store'" json:"NoStore"`                       // 是否不存储 1:是 0:否
	MaxChannelNum     int        `xorm:"'max_channel_num'" json:"MaxChannelNum"`          // 最大隧道数
	MaxConnectNum     int        `xorm:"'max_connect_num'" json:"MaxConnectNum"`          // 最大连接数
	RateLimit         int        `xorm:"'rate_limit'" json:"RateLimit"`                   // 带宽限制 kb/s
	FlowLimit         int64      `xorm:"'flow_limit'" json:"FlowLimit"`                   // 流量限制 B
	WebUser           string     `xorm:"'web_user'" json:"WebUser"`                       // web 登陆用户名
	WebPass           string     `xorm:"'web_pass'" json:"WebPass"`                       // web 登陆密码
	CreateTime        time.Time  `xorm:"created" json:"createTime"`                       // 创建日期
	UpdateTime        time.Time  `xorm:"updated" json:"updateTime"`                       // 更新日期
	Rate              *rate.Rate `xorm:"-"`                                               //rate limit
}

// NpsClientListInfo represents a nps_client_info + other struct data.
type NpsClientListInfo struct {
	NpsClientInfo             `xorm:"extends"`
	NpsClientStatisticFlow    `xorm:"extends"`
	NpsClientStatisticRate    `xorm:"extends"`
	NpsClientStatisticConnect `xorm:"extends"`
}

func (s *NpsClientListInfo) CutConn() {
	atomic.AddInt32(&s.NowConnectNum, 1)
}

func (s *NpsClientListInfo) AddConn() {
	atomic.AddInt32(&s.NowConnectNum, -1)
}

func (s *NpsClientListInfo) GetConn() bool {
	if s.MaxConnectNum == 0 || int(s.NowConnectNum) < s.MaxConnectNum {
		s.CutConn()
		return true
	}
	return false
}

func NewClientInit(vKey string, noStore bool, noDisplay bool) *NpsClientInfo {
	return &NpsClientInfo{
		VerifyKey: vKey,
		Addr:      "",
		Remark:    "",
		Status:    true,
		IsConnect: false,
		RateLimit: 0,
		Rate:      nil,
		NoStore:   noStore,
		NoDisplay: noDisplay,
	}
}

func (NpsClientInfo) TableName() string {
	return "nps_client_info"
}
