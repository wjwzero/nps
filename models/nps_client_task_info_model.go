package models

import (
	"ehang.io/nps/lib/file"
	"github.com/pkg/errors"
	"strings"
	"sync"
	"time"
	"xorm.io/xorm"
)

// NpsClientTaskInfoModel represents a nps_client_task_info model.
type NpsClientTaskInfoModel struct {
	engine xorm.EngineInterface
}

// NpsClientTaskInfo represents a nps_client_task_info struct data.
type NpsClientTaskInfo struct {
	Id              int                `xorm:"pk autoincr 'id'" json:"TaskId"`       // 主键
	ClientId        int                `xorm:"'client_id'" json:"ClientId"`          // 客户端主键
	Mode            string             `xorm:"'mode'" json:"Mode"`                   // p2p/tcp/udp/httpProxy/socks5/secret/file
	Remark          string             `xorm:"'remark'" json:"TaskRemark"`           // 备注
	ServerIp        string             `xorm:"'server_ip'" json:"ServerIp"`          // 服务器ip
	Port            int                `xorm:"'port'" json:"Port"`                   // 端口
	Password        string             `xorm:"'password'" json:"Password"`           // 唯一标识密钥
	Ports           string             `xorm:"'ports'" json:"Ports"`                 // 端口集合
	Account         string             `xorm:"'account'" json:"Account"`             // socks5账号
	TargetAddr      string             `xorm:"'target_addr'" json:"TargetAddr"`      // 内网目标
	LocalPath       string             `xorm:"'local_path'" json:"LocalPath"`        // 本地文件目录
	StripPre        string             `xorm:"'strip_pre'" json:"StripPre"`          // 前缀
	NoStore         bool               `xorm:"'no_store'" json:"NoStore"`            // 是否不存储 1:是 0:否
	Status          bool               `xorm:"'status'" json:"TaskStatus"`           // 状态 1:开放 0:关闭
	RunStatus       bool               `xorm:"'run_status'" json:"RunStatus"`        // 运行状态 1:开放 0:关闭
	TargetStr       string             `xorm:"'target_str'" json:"TargetStr"`        // 目标 (IP:端口)
	IsLocalProxy    bool               `xorm:"'is_local_proxy'" json:"IsLocalProxy"` // 是否为本地代理 1:是 0:否
	CreateTime      time.Time          `xorm:"created" json:"createTime"`            // 创建日期
	UpdateTime      time.Time          `xorm:"updated" json:"updateTime"`            // 更新日期
	MultiAccount    *file.MultiAccount `xorm:"-"`
	Flow            *file.Flow         `xorm:"-"`
	Client          *NpsClientListInfo `xorm:"-"`
	TargetArr       []string           `xorm:"-"`
	nowIndex        int                `xorm:"-"`
	HealthRemoveArr []string           `xorm:"-"`
	sync.RWMutex    `xorm:"-"`
}

func (NpsClientTaskInfo) TableName() string {
	return "nps_client_task_info"
}

// NpsClientTaskListInfo represents a nps_client_task_info + nps_client_info.
type NpsClientTaskListInfo struct {
	NpsClientInfo     `xorm:"extends"`
	NpsClientTaskInfo `xorm:"extends"`
}

func (s NpsClientTaskInfo) GetRandomTarget() (string, error) {
	if s.TargetArr == nil {
		s.TargetArr = strings.Split(s.TargetStr, "\n")
	}
	if len(s.TargetArr) == 1 {
		return s.TargetArr[0], nil
	}
	if len(s.TargetArr) == 0 {
		return "", errors.New("all inward-bending targets are offline")
	}
	s.Lock()
	defer s.Unlock()
	if s.nowIndex >= len(s.TargetArr)-1 {
		s.nowIndex = -1
	}
	s.nowIndex++
	return s.TargetArr[s.nowIndex], nil
}
