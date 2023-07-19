package db

import (
	"ehang.io/nps/lib/common"
	"ehang.io/nps/lib/crypt"
	"ehang.io/nps/lib/rate"
	. "ehang.io/nps/models"
	"errors"
	"github.com/astaxie/beego/logs"
	"sync"
)

type ClientDao struct {
	Clients sync.Map
}

var client_cols = []string{"nps_client_info.id", "verify_key", "addr", "basic_auth_user", "basic_auth_pass", "device_key", "version", "status", "product_key", "remark", "is_connect", "is_config_conn_allow", "is_compress", "is_crypt", "no_display", "no_store", "max_channel_num", "max_connect_num", "rate_limit", "flow_limit", "web_user", "web_pass", "nps_client_info.create_time", "nps_client_info.update_time"}
var clientUpdateCols = []string{"verify_key", "product_key", "remark", "basic_auth_user", "basic_auth_pass", "is_config_conn_allow", "is_compress", "is_crypt"}

//func (cd *ClientDao) LoadClientFromDb() {
//	list, num := cd.GetClientAllListInfo()
//	if num == 0 {
//		return
//	}
//	for _, v := range list {
//		if v.RateLimit > 0 {
//			v.Rate = rate.NewRate(int64(v.RateLimit * 1024))
//		} else {
//			v.Rate = rate.NewRate(int64(2 << 23))
//		}
//		v.Rate.Start()
//		v.NowConnectNum = 0
//		cd.Clients.Store(v.Id, v)
//	}
//}

// query Client by id
func (ClientDao) GetClient(id int) (*NpsClientInfo, error) {
	cc := new(NpsClientInfo)
	has, err := DbEngine.Where("id = ?", id).Get(cc)
	if !has {
		err = errors.New("未找到客户端")
		return cc, err
	}
	return cc, nil
}

// query ClientId by vkey; vkey md5？？
func (ClientDao) GetClientIdByVkey(vkey string) (id int, err error) {
	c := new(NpsClientInfo)
	has, err := DbEngine.Where("verify_key = ?", vkey).Get(c)
	if !has {
		err = errors.New("未找到客户端")
	}
	id = int(c.Id)
	return
}

// query Client by deviceKey
func (ClientDao) GetClientByDeviceKey(deviceKey string) (client *NpsClientInfo) {
	has, dbErr := DbEngine.Where("device_key = ?", deviceKey).Get(client)
	if !has {
		dbErr = errors.New("未找到客户端")
	}
	if dbErr != nil {
		logs.Error(dbErr)
		return
	}
	return
}

// Delete is delete a Client
func (ClientDao) DelClient(id int) error {
	c := new(NpsClientInfo)
	_, err := DbEngine.ID(id).Delete(c)
	err = DelClientConnect(id)
	err = DelClientFlow(id)
	err = DelClientRate(id)
	return err
}
func DelClientConnect(clientId int) error {
	cc := new(NpsClientStatisticConnect)
	cc.ClientId = clientId
	_, err := DbEngine.Cols("client_id").Delete(cc)
	return err
}

func DelClientFlow(clientId int) error {
	cf := new(NpsClientStatisticFlow)
	cf.ClientId = clientId
	_, err := DbEngine.Cols("client_id").Delete(cf)
	return err
}

func DelClientRate(clientId int) error {
	cr := new(NpsClientStatisticRate)
	cr.ClientId = clientId
	_, err := DbEngine.Cols("client_id").Delete(cr)
	return err
}

// get client list search{Id,VerifyKey,Remark} sort{CreateTime DESC}
func (ClientDao) GetClientList(start, length int, search, sort, order string, clientId int) ([]*NpsClientInfo, int) {
	list := make([]*NpsClientInfo, 0)
	cnt, dbErr := DbEngine.Cols(client_cols...).Limit(length, start).FindAndCount(&list)
	if dbErr != nil {
		logs.Error(dbErr)
		return list, 0
	}
	return list, int(cnt)
}

// get client list search{Id,VerifyKey,Remark} sort{CreateTime DESC}
func (ClientDao) GetClientListInfo(start, length int, search, sort, order string, clientId int) ([]*NpsClientListInfo, int) {
	list := make([]*NpsClientListInfo, 0)
	cnt, dbErr := DbEngine.Table("nps_client_info").
		Select("nps_client_info.id, `verify_key`, `addr`, `basic_auth_user`, `basic_auth_pass`, `device_key`, `version`, `status`,`product_key`, `remark`, `is_connect`, `is_config_conn_allow`, `is_compress`, `is_crypt`, `no_display`, `no_store`, `max_channel_num`, `max_connect_num`, `rate_limit`, `flow_limit`, `web_user`, `web_pass`, `nps_client_info`.`create_time`, `nps_client_info`.`update_time`,nps_client_statistic_flow.flow_inlet,nps_client_statistic_flow.flow_export,nps_client_statistic_rate.rate_now,nps_client_statistic_connect.now_connect_num").
		Join("LEFT", "nps_client_statistic_flow", "nps_client_info.Id = nps_client_statistic_flow.client_id").
		Join("LEFT", "nps_client_statistic_rate", "nps_client_info.Id = nps_client_statistic_rate.client_id").
		Join("LEFT", "nps_client_statistic_connect", "nps_client_info.Id = nps_client_statistic_connect.client_id").
		Limit(length, start).OrderBy("nps_client_info.id").FindAndCount(&list)
	if dbErr != nil {
		logs.Error(dbErr)
		return list, 0
	}
	return list, int(cnt)
}

// query ClientInfo by id
func (ClientDao) GetClientInfo(id int) (*NpsClientListInfo, error) {
	if id == 0 {
		return nil, errors.New("未找到客户端")
	}
	cc := new(NpsClientListInfo)
	has, err := DbEngine.Table("nps_client_info").
		Select("nps_client_info.id, `verify_key`, `addr`, `basic_auth_user`, `basic_auth_pass`, `device_key`, `version`, `status`,`product_key`, `remark`, `is_connect`, `is_config_conn_allow`, `is_compress`, `is_crypt`, `no_display`, `no_store`, `max_channel_num`, `max_connect_num`, `rate_limit`, `flow_limit`, `web_user`, `web_pass`, `nps_client_info`.`create_time`, `nps_client_info`.`update_time`,nps_client_statistic_flow.flow_inlet,nps_client_statistic_flow.flow_export,nps_client_statistic_rate.rate_now,nps_client_statistic_connect.now_connect_num").
		Join("LEFT", "nps_client_statistic_flow", "nps_client_info.Id = nps_client_statistic_flow.client_id").
		Join("LEFT", "nps_client_statistic_rate", "nps_client_info.Id = nps_client_statistic_rate.client_id").
		Join("LEFT", "nps_client_statistic_connect", "nps_client_info.Id = nps_client_statistic_connect.client_id").
		Where("nps_client_info.id = ?", id).Get(cc)
	if !has {
		err = errors.New("未找到客户端")
		return cc, err
	}
	return cc, nil
}

func (ClientDao) GetClientAllListInfo() ([]*NpsClientListInfo, int) {
	list := make([]*NpsClientListInfo, 0)
	cnt, dbErr := DbEngine.Table("nps_client_info").
		Select("nps_client_info.id, `verify_key`, `addr`, `basic_auth_user`, `basic_auth_pass`, `device_key`, `version`, `status`,`product_key`, `remark`, `is_connect`, `is_config_conn_allow`, `is_compress`, `is_crypt`, `no_display`, `no_store`, `max_channel_num`, `max_connect_num`, `rate_limit`, `flow_limit`, `web_user`, `web_pass`, `nps_client_info`.`create_time`, `nps_client_info`.`update_time`,nps_client_statistic_flow.flow_inlet,nps_client_statistic_flow.flow_export,nps_client_statistic_rate.rate_now,nps_client_statistic_connect.now_connect_num").
		Join("LEFT", "nps_client_statistic_flow", "nps_client_info.Id = nps_client_statistic_flow.client_id").
		Join("LEFT", "nps_client_statistic_rate", "nps_client_info.Id = nps_client_statistic_rate.client_id").
		Join("LEFT", "nps_client_statistic_connect", "nps_client_info.Id = nps_client_statistic_connect.client_id").
		FindAndCount(&list)
	if dbErr != nil {
		logs.Error(dbErr)
		return list, 0
	}
	return list, int(cnt)
}

func (cd ClientDao) NewClient(c *NpsClientInfo) error {
	var isNotSet bool
	if c.WebUser != "" && !cd.VerifyUserName(c.WebUser, c.Id) {
		return errors.New("web login username duplicate, please reset")
	}
reset:
	if c.VerifyKey == "" || isNotSet {
		isNotSet = true
		c.VerifyKey = crypt.GetRandomString(16)
	}
	if c.RateLimit == 0 {
		c.Rate = rate.NewRate(int64(2 << 23))
	} else if c.Rate == nil {
		c.Rate = rate.NewRate(int64(c.RateLimit * 1024))
	}
	c.Rate.Start()
	if !cd.VerifyVkey(c.VerifyKey, c.Id) {
		if isNotSet {
			goto reset
		}
		return errors.New("Vkey duplicate, please reset")
	}
	cd.Save(c)
	return nil
}

func (ClientDao) VerifyVkey(vkey string, id int) (res bool) {
	res = true
	c := new(NpsClientInfo)
	result, dbErr := DbEngine.Where("verify_key = ?", vkey).Get(c)
	if dbErr != nil {
		logs.Error(dbErr)
		return false
	}
	if !result {
		logs.Info("Client not found")
		return true
	}
	if c.Id != id {
		res = false
	}
	return res
}

func (ClientDao) Save(c *NpsClientInfo) error {
	if c.NoStore {
		return nil
	}
	_, dbErr := DbEngine.Insert(c)
	if dbErr != nil {
		logs.Error(dbErr)
		return dbErr
	}
	//dbErr = SaveClientConnect(c.Id)
	//dbErr = SaveClientFlow(c.Id)
	//dbErr = SaveClientRate(c.Id)
	return dbErr
}

func SaveClientConnect(clientId int) error {
	cc := new(NpsClientStatisticConnect)
	cc.ClientId = clientId
	_, dbErr := DbEngine.Insert(cc)
	if dbErr != nil {
		logs.Error(dbErr)
		return dbErr
	}
	return dbErr
}

func SaveClientFlow(clientId int) error {
	cf := new(NpsClientStatisticFlow)
	cf.ClientId = clientId
	_, dbErr := DbEngine.Insert(cf)
	if dbErr != nil {
		logs.Error(dbErr)
		return dbErr
	}
	return dbErr
}

func SaveClientRate(clientId int) error {
	cr := new(NpsClientStatisticRate)
	cr.ClientId = clientId
	_, dbErr := DbEngine.Insert(cr)
	if dbErr != nil {
		logs.Error(dbErr)
		return dbErr
	}
	return dbErr
}

func (ClientDao) SaveEdit(c *NpsClientInfo) error {
	_, dbErr := DbEngine.ID(c.Id).Cols(clientUpdateCols...).Update(c)
	if dbErr != nil {
		logs.Error(dbErr)
	}
	return dbErr
}

func (ClientDao) VerifyUserName(username string, id int) (res bool) {
	res = true
	c := new(NpsClientInfo)
	result, dbErr := DbEngine.Where("web_user = ?", username).Get(&c)
	if dbErr != nil {
		logs.Error(dbErr)
		return false
	}
	if !result {
		logs.Info("Client not found")
		return true
	}
	if c.Id != id {
		res = false
	}
	return res
}

func (cd ClientDao) IsPubClient(id int) bool {
	client, err := cd.GetClient(id)
	if err == nil {
		return client.NoDisplay
	}
	return false
}

func (cd ClientDao) UpdateStatus(c *NpsClientInfo) error {
	_, dbErr := DbEngine.ID(c.Id).Cols("status").Update(c)
	if dbErr != nil {
		logs.Error(dbErr)
	}
	return dbErr
}

func (cd ClientDao) UpdateIsConnect(c *NpsClientInfo) error {
	_, dbErr := DbEngine.ID(c.Id).Cols("is_connect").Update(c)
	if dbErr != nil {
		logs.Error(dbErr)
	}
	return dbErr
}

func (ClientDao) GetIdByVerifyKey(vKey string, addr string) (id int, status bool, err error) {
	c := new(NpsClientInfo)
	has, err := DbEngine.Where("verify_key = ?", vKey).Get(c)
	if !has {
		return 0, status, errors.New("not found")
	}
	status = c.Status
	if c.Status {
		c.Addr = common.GetIpByAddr(addr)
		id = c.Id
		return
	}
	return
}

func (cd ClientDao) PreCreateVerifyKeyClient(verifyKey string, productKey string) (id int, err error) {
	t := &NpsClientInfo{
		VerifyKey:         verifyKey,
		Remark:            verifyKey,
		ProductKey:        productKey,
		DeviceKey:         verifyKey,
		Status:            true,
		IsCompress:        false,
		IsCrypt:           false,
		IsConfigConnAllow: true,
		BasicAuthUser:     "",
		BasicAuthPass:     "",
		RateLimit:         0,
		MaxChannelNum:     0,
		MaxConnectNum:     0,
		WebUser:           "",
		WebPass:           "",
		FlowLimit:         0,
	}
	if err := cd.NewClient(t); err != nil {
		err = errors.New("创建verifyKey Client失败")
	}
	id = t.Id
	return
}

func (cd ClientDao) UpdateAddressOnline(id int, addr string) error {
	c := new(NpsClientInfo)
	c.Id = id
	c.Addr = common.GetIpByAddr(addr)
	c.IsConnect = true
	_, dbErr := DbEngine.ID(c.Id).Cols("addr", "is_connect").Update(c)
	if dbErr != nil {
		logs.Error(dbErr)
	}
	return dbErr
}

func (cd ClientDao) UpdateStatusOffline(id int) error {
	c := new(NpsClientInfo)
	c.Id = id
	c.IsConnect = false
	return cd.UpdateIsConnect(c)
}

func (cd ClientDao) UpdateVersion(id int, version string) error {
	c := new(NpsClientInfo)
	c.Id = id
	c.Version = version
	_, dbErr := DbEngine.ID(c.Id).Cols("version").Update(c)
	if dbErr != nil {
		logs.Error(dbErr)
	}
	return dbErr
}

func (cd ClientDao) UpdateVersionAndIsConnect(id int, version string, isConnect bool) error {
	c := new(NpsClientInfo)
	c.Id = id
	c.IsConnect = isConnect
	c.Version = version
	_, dbErr := DbEngine.ID(c.Id).Cols("is_connect", "version").Update(c)
	if dbErr != nil {
		logs.Error(dbErr)
	}
	return dbErr
}
