package models

import (
	"time"

	"xorm.io/xorm"
)

// NpsClientStatisticRateModel represents a nps_client_statistic_rate model.
type NpsClientStatisticRateModel struct {
	engine xorm.EngineInterface
}

// NpsClientStatisticRate represents a nps_client_statistic_rate struct data.
type NpsClientStatisticRate struct {
	RateId     int       `xorm:"pk autoincr " json:"RateId"`  // 主键
	ClientId   int       `xorm:"'client_id'" json:"ClientId"` // 客户端主键
	RateNow    int64     `xorm:"'rate_now'" json:"RateNow"`   // 网速 B/S
	CreateTime time.Time `xorm:"created" json:"createTime"`   // 创建日期
	UpdateTime time.Time `xorm:"updated" json:"updateTime"`   // 更新日期
}

func (NpsClientStatisticRate) TableName() string {
	return "nps_client_statistic_rate"
}
