package models

import (
	"ehang.io/nps/lib/file"
	"github.com/pkg/errors"
	"strings"
	"sync"
	"time"
	"xorm.io/xorm"
)

// NpsClientHostInfoModel represents a nps_client_host_info model.
type NpsClientHostInfoModel struct {
	engine xorm.EngineInterface
}

// NpsClientHostInfo represents a nps_client_host_info struct data.
type NpsClientHostInfo struct {
	Id              int                `xorm:"pk autoincr 'id'" json:"HostId"`       // 主键
	ClientId        int                `xorm:"'client_id'" json:"ClientId"`          // 客户端主键
	Remark          string             `xorm:"'remark'" json:"HostRemark"`           // 备注
	Host            string             `xorm:"'host'" json:"Host"`                   // 主机
	HostChange      string             `xorm:"'host_change'" json:"HostChange"`      // 请求主机信息修改
	HeaderChange    string             `xorm:"'header_change'" json:"HeaderChange"`  // 请求头部信息修改;多个冒号分割
	KeyFilePath     string             `xorm:"'key_file_path'" json:"KeyFilePath"`   // 密钥文件路径
	CertFilePath    string             `xorm:"'cert_file_path'" json:"CertFilePath"` // 证书文件路径
	Location        string             `xorm:"'location'" json:"Location"`           // URL 路由
	NoStore         bool               `xorm:"'no_store'" json:"NoStore"`            // 是否不存储 1:是 0:否
	IsClose         bool               `xorm:"'is_close'" json:"IsClose"`            // 是否关闭 1:是 0:否
	Scheme          string             `xorm:"'scheme'" json:"Scheme"`               // 模式
	TargetStr       string             `xorm:"'target_str'" json:"TargetStr"`        // 目标 (IP:端口)
	IsLocalProxy    bool               `xorm:"'is_local_proxy'" json:"IsLocalProxy"` // 是否为本地代理 1:是 0:否
	CreateTime      time.Time          `xorm:"created" json:"createTime"`            // 创建日期
	UpdateTime      time.Time          `xorm:"updated" json:"updateTime"`            // 更新日期
	Flow            *file.Flow         `xorm:"-"`
	Client          *NpsClientListInfo `xorm:"-"`
	TargetArr       []string           `xorm:"-"`
	nowIndex        int                `xorm:"-"`
	HealthRemoveArr []string           `xorm:"-"`
	sync.RWMutex    `xorm:"-"`
}

// NpsClientHostListInfo represents a nps_client_host_info + nps_client_info.
type NpsClientHostListInfo struct {
	NpsClientInfo     `xorm:"extends"`
	NpsClientHostInfo `xorm:"extends"`
}

func (NpsClientHostInfo) TableName() string {
	return "nps_client_host_info"
}

func (s NpsClientHostInfo) GetRandomTarget() (string, error) {
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
