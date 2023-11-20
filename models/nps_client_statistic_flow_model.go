package models

import (
	"time"

	"xorm.io/xorm"
)

// NpsClientStatisticFlowModel represents a nps_client_statistic_flow model.
type NpsClientStatisticFlowModel struct {
	engine xorm.EngineInterface
}

// NpsClientStatisticFlow represents a nps_client_statistic_flow struct data.
type NpsClientStatisticFlow struct {
	FlowId     int       `xorm:"pk autoincr " json:"FlowId"`      // 主键
	ClientId   int       `xorm:"'client_id'" json:"ClientId"`     // 客户端主键
	FlowInlet  int64     `xorm:"'flow_inlet'" json:"FlowInlet"`   // 入口流量 B
	FlowExport int64     `xorm:"'flow_export'" json:"FlowExport"` // 出口流量 B
	CreateTime time.Time `xorm:"created" json:"createTime"`       // 创建日期
	UpdateTime time.Time `xorm:"updated" json:"updateTime"`       // 更新日期
}

func (NpsClientStatisticFlow) TableName() string {
	return "nps_client_statistic_flow"
}
