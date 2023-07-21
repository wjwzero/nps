package db

import (
	"ehang.io/nps/models"
	"github.com/astaxie/beego/logs"
)

type ClientStatisticDao struct {
}

func (cd ClientStatisticDao) UpdateConnectNum(clientId int, NowConnectNum int32) {
	npsCsc := new(models.NpsClientStatisticConnect)
	npsCsc.ClientId = clientId
	npsCsc.NowConnectNum = NowConnectNum
	_, dbErr := DbEngine.Where("client_id = ?", npsCsc.ClientId).Cols("now_connect_num").Update(npsCsc)
	if dbErr != nil {
		logs.Error(dbErr)
	}
}

func (cd ClientStatisticDao) UpdateFlow(clientId int, inlet int64, export int64) {
	npsCsf := new(models.NpsClientStatisticFlow)
	npsCsf.ClientId = clientId
	npsCsf.FlowInlet = inlet
	npsCsf.FlowExport = export
	_, dbErr := DbEngine.Where("client_id = ?", npsCsf.ClientId).Cols("flow_inlet", "flow_export").Update(npsCsf)
	if dbErr != nil {
		logs.Error(dbErr)
	}
}

func (cd ClientStatisticDao) UpdateRate(clientId int, nowRate int64) {
	npsCsr := new(models.NpsClientStatisticRate)
	npsCsr.ClientId = clientId
	npsCsr.RateNow = nowRate
	_, dbErr := DbEngine.Where("client_id = ?", npsCsr.ClientId).Cols("rate_now").Update(npsCsr)
	if dbErr != nil {
		logs.Error(dbErr)
	}
}
