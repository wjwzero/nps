package models

import (
	"time"

	"xorm.io/xorm"
)

// NpsClientStatisticConnectModel represents a nps_client_statistic_connect model.
type NpsClientStatisticConnectModel struct {
	engine xorm.EngineInterface
}

// NpsClientStatisticConnect represents a nps_client_statistic_connect struct data.
type NpsClientStatisticConnect struct {
	ConnectId     int       `xorm:"pk autoincr " json:"ConnectId"`          // 主键
	ClientId      int       `xorm:"'client_id'" json:"ClientId"`            // 客户端主键
	NowConnectNum int32     `xorm:"'now_connect_num'" json:"NowConnectNum"` // 当前连接数
	CreateTime    time.Time `xorm:"created" json:"createTime"`              // 创建日期
	UpdateTime    time.Time `xorm:"updated" json:"updateTime"`              // 更新日期
}

func (NpsClientStatisticConnect) TableName() string {
	return "nps_client_statistic_connect"
}
