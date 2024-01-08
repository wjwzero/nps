package file

import (
	"errors"
	"fmt"
	"github.com/astaxie/beego/logs"
	"net/http"
	"sort"
	"strings"
	"sync"

	"ehang.io/nps/lib/common"
	"ehang.io/nps/lib/crypt"
	"ehang.io/nps/lib/rate"
)

type DbUtils struct {
	JsonDb *JsonDb
}

var (
	Db   *DbUtils
	once sync.Once
)

//init csv from file
func GetDb() *DbUtils {
	once.Do(func() {
		jsonDb := NewJsonDb(common.GetRunPath())
		// 配置优先加载
		jsonDb.LoadDynamicConfigFromJsonFile()
		jsonDb.LoadClientFromJsonFile()
		jsonDb.LoadTaskFromJsonFile()
		jsonDb.LoadHostFromJsonFile()
		jsonDb.LoadTunnelTypesProductRelationFromJsonFile()
		Db = &DbUtils{JsonDb: jsonDb}
	})
	return Db
}

func GetMapKeys(m sync.Map, isSort bool, sortKey, order string) (keys []int) {
	if sortKey != "" && isSort {
		return sortClientByKey(m, sortKey, order)
	}
	m.Range(func(key, value interface{}) bool {
		keys = append(keys, key.(int))
		return true
	})
	sort.Ints(keys)
	return
}

func (s *DbUtils) GetClientList(start, length int, search, sort, order string, clientId int) ([]*Client, int) {
	list := make([]*Client, 0)
	var cnt int
	keys := GetMapKeys(s.JsonDb.Clients, true, sort, order)
	for _, key := range keys {
		if value, ok := s.JsonDb.Clients.Load(key); ok {
			v := value.(*Client)
			if v.NoDisplay {
				continue
			}
			if clientId != 0 && clientId != v.Id {
				continue
			}
			if search != "" && !(v.Id == common.GetIntNoErrByStr(search) || strings.Contains(v.VerifyKey, search) || strings.Contains(v.Remark, search)) {
				continue
			}
			cnt++
			if start--; start < 0 {
				if length--; length >= 0 {
					list = append(list, v)
				}
			}
		}
	}
	return list, cnt
}

func (s *DbUtils) GetIdByVerifyKey(vKey string, addr string) (id int, err error) {
	var exist bool
	s.JsonDb.Clients.Range(func(key, value interface{}) bool {
		v := value.(*Client)
		if v.VerifyKey == vKey && v.Status {
			v.Addr = common.GetIpByAddr(addr)
			id = v.Id
			exist = true
			return false
		}
		return true
	})
	if exist {
		return
	}
	return 0, errors.New("not found")
}

func (s *DbUtils) NewTask(t *Tunnel) (err error) {
	s.JsonDb.Tasks.Range(func(key, value interface{}) bool {
		v := value.(*Tunnel)
		if (v.Mode == "secret" || v.Mode == "p2p") && v.Password == t.Password {
			err = errors.New(fmt.Sprintf("secret mode keys %s must be unique", t.Password))
			return false
		}
		return true
	})
	if err != nil {
		return
	}
	t.Flow = new(Flow)
	s.JsonDb.Tasks.Store(t.Id, t)
	s.JsonDb.StoreTasksToJsonFile()
	return
}

// 根据password 删除Task
func (s *DbUtils) DelTaskByPassword(password string) (err error) {
	var taskId int
	s.JsonDb.Tasks.Range(func(key, value interface{}) bool {
		v := value.(*Tunnel)
		if v.Password == password {
			taskId = v.Id
			return false
		}
		return true
	})
	s.JsonDb.Tasks.Delete(taskId)
	s.JsonDb.StoreTasksToJsonFile()
	return
}

func (s *DbUtils) UpdateTask(t *Tunnel) error {
	s.JsonDb.Tasks.Store(t.Id, t)
	s.JsonDb.StoreTasksToJsonFile()
	return nil
}

func (s *DbUtils) DelTask(id int) error {
	s.JsonDb.Tasks.Delete(id)
	s.JsonDb.StoreTasksToJsonFile()
	return nil
}

//md5 password
func (s *DbUtils) GetTaskByMd5Password(p string) (t *Tunnel) {
	s.JsonDb.Tasks.Range(func(key, value interface{}) bool {
		if crypt.Md5(value.(*Tunnel).Password) == p {
			t = value.(*Tunnel)
			return false
		}
		return true
	})
	return
}

func (s *DbUtils) GetTaskByPassword(p string) (t *Tunnel) {
	s.JsonDb.Tasks.Range(func(key, value interface{}) bool {
		if value.(*Tunnel).Password == p {
			t = value.(*Tunnel)
			return false
		}
		return true
	})
	return
}

func (s *DbUtils) GetTask(id int) (t *Tunnel, err error) {
	if v, ok := s.JsonDb.Tasks.Load(id); ok {
		t = v.(*Tunnel)
		return
	}
	err = errors.New("not found")
	return
}

func (s *DbUtils) DelHost(id int) error {
	s.JsonDb.Hosts.Delete(id)
	s.JsonDb.StoreHostToJsonFile()
	return nil
}

func (s *DbUtils) IsHostExist(h *Host) bool {
	var exist bool
	s.JsonDb.Hosts.Range(func(key, value interface{}) bool {
		v := value.(*Host)
		if v.Id != h.Id && v.Host == h.Host && h.Location == v.Location && (v.Scheme == "all" || v.Scheme == h.Scheme) {
			exist = true
			return false
		}
		return true
	})
	return exist
}

func (s *DbUtils) NewHost(t *Host) error {
	if t.Location == "" {
		t.Location = "/"
	}
	if s.IsHostExist(t) {
		return errors.New("host has exist")
	}
	t.Flow = new(Flow)
	s.JsonDb.Hosts.Store(t.Id, t)
	s.JsonDb.StoreHostToJsonFile()
	return nil
}

func (s *DbUtils) GetHost(start, length int, id int, search string) ([]*Host, int) {
	list := make([]*Host, 0)
	var cnt int
	keys := GetMapKeys(s.JsonDb.Hosts, false, "", "")
	for _, key := range keys {
		if value, ok := s.JsonDb.Hosts.Load(key); ok {
			v := value.(*Host)
			if search != "" && !(v.Id == common.GetIntNoErrByStr(search) || strings.Contains(v.Host, search) || strings.Contains(v.Remark, search)) {
				continue
			}
			if id == 0 || v.Client.Id == id {
				cnt++
				if start--; start < 0 {
					if length--; length >= 0 {
						list = append(list, v)
					}
				}
			}
		}
	}
	return list, cnt
}

func (s *DbUtils) DelClient(id int) error {
	s.JsonDb.Clients.Delete(id)
	s.JsonDb.StoreClientsToJsonFile()
	return nil
}

func (s *DbUtils) NewClient(c *Client) error {
	var isNotSet bool
	if c.WebUserName != "" && !s.VerifyUserName(c.WebUserName, c.Id) {
		return errors.New("web login username duplicate, please reset")
	}
reset:
	if c.VerifyKey == "" || isNotSet {
		isNotSet = true
		c.VerifyKey = crypt.GetRandomString(16)
	}
	if c.RateLimit == 0 {
		var rateMaxErr error
		var rateLimit int64
		if rateLimit, rateMaxErr = s.JsonDb.GetCommonRateLimitMax(); rateMaxErr != nil {
			logs.Error("获取动态配置失败 GetCommonRateLimitMax 默认 %s kb", rateLimit, rateMaxErr.Error())
		}
		c.Rate = rate.NewRate(int64(rateLimit * 1024))
	} else if c.Rate == nil {
		c.Rate = rate.NewRate(int64(c.RateLimit * 1024))
	}
	c.Rate.Start()
	if !s.VerifyVkey(c.VerifyKey, c.Id) {
		if isNotSet {
			goto reset
		}
		return errors.New("Vkey duplicate, please reset")
	}
	if c.Id == 0 {
		c.Id = int(s.JsonDb.GetClientId())
	}
	if c.Flow == nil {
		c.Flow = new(Flow)
	}
	s.JsonDb.Clients.Store(c.Id, c)
	s.JsonDb.StoreClientsToJsonFile()
	return nil
}

func (s *DbUtils) VerifyVkey(vkey string, id int) (res bool) {
	res = true
	s.JsonDb.Clients.Range(func(key, value interface{}) bool {
		v := value.(*Client)
		if v.VerifyKey == vkey && v.Id != id {
			res = false
			return false
		}
		return true
	})
	return res
}

func (s *DbUtils) VerifyUserName(username string, id int) (res bool) {
	res = true
	s.JsonDb.Clients.Range(func(key, value interface{}) bool {
		v := value.(*Client)
		if v.WebUserName == username && v.Id != id {
			res = false
			return false
		}
		return true
	})
	return res
}

func (s *DbUtils) UpdateClient(t *Client) error {
	s.JsonDb.Clients.Store(t.Id, t)
	if t.RateLimit == 0 {
		var rateMaxErr error
		var rateLimit int64
		if rateLimit, rateMaxErr = s.JsonDb.GetCommonRateLimitMax(); rateMaxErr != nil {
			logs.Error("获取动态配置失败 GetCommonRateLimitMax 默认 %s kb", rateLimit, rateMaxErr.Error())
		}
		t.Rate = rate.NewRate(int64(rateLimit * 1024))
		t.Rate.Start()
	}
	return nil
}

func (s *DbUtils) IsPubClient(id int) bool {
	client, err := s.GetClient(id)
	if err == nil {
		return client.NoDisplay
	}
	return false
}

func (s *DbUtils) GetClient(id int) (c *Client, err error) {
	if v, ok := s.JsonDb.Clients.Load(id); ok {
		c = v.(*Client)
		return
	}
	err = errors.New("未找到客户端")
	return
}

func (s *DbUtils) GetClientIdByVkey(vkey string) (id int, err error) {
	var exist bool
	s.JsonDb.Clients.Range(func(key, value interface{}) bool {
		v := value.(*Client)
		if crypt.Md5(v.VerifyKey) == vkey {
			exist = true
			id = v.Id
			return false
		}
		return true
	})
	if exist {
		return
	}
	err = errors.New("未找到客户端")
	return
}

func (s *DbUtils) GetClientByDeviceKey(deviceKey string) (client *Client) {
	var exist bool
	s.JsonDb.Clients.Range(func(key, value interface{}) bool {
		clientObj := value.(*Client)
		if clientObj.DeviceKey == deviceKey {
			exist = true
			client = clientObj
			return false
		}
		return true
	})
	if exist {
		return
	}
	return
}

func (s *DbUtils) GetHostById(id int) (h *Host, err error) {
	if v, ok := s.JsonDb.Hosts.Load(id); ok {
		h = v.(*Host)
		return
	}
	err = errors.New("The host could not be parsed")
	return
}

//get key by host from x
func (s *DbUtils) GetInfoByHost(host string, r *http.Request) (h *Host, err error) {
	var hosts []*Host
	//Handling Ported Access
	host = common.GetIpByAddr(host)
	s.JsonDb.Hosts.Range(func(key, value interface{}) bool {
		v := value.(*Host)
		if v.IsClose {
			return true
		}
		//Remove http(s) http(s)://a.proxy.com
		//*.proxy.com *.a.proxy.com  Do some pan-parsing
		if v.Scheme != "all" && v.Scheme != r.URL.Scheme {
			return true
		}
		tmpHost := v.Host
		if strings.Contains(tmpHost, "*") {
			tmpHost = strings.Replace(tmpHost, "*", "", -1)
			if strings.Contains(host, tmpHost) {
				hosts = append(hosts, v)
			}
		} else if v.Host == host {
			hosts = append(hosts, v)
		}
		return true
	})

	for _, v := range hosts {
		//If not set, default matches all
		if v.Location == "" {
			v.Location = "/"
		}
		if strings.Index(r.RequestURI, v.Location) == 0 {
			if h == nil || (len(v.Location) > len(h.Location)) {
				h = v
			}
		}
	}
	if h != nil {
		return
	}
	err = errors.New("The host could not be parsed")
	return
}

func (s *DbUtils) PreCreateVerifyKeyClient(verifyKey string, productKey string) (id int, err error) {
	clientId := int(s.JsonDb.GetClientId())
	t := &Client{
		VerifyKey: verifyKey,
		Id:        clientId,
		Status:    true,
		Remark:    verifyKey,
		Cnf: &Config{
			U:        "",
			P:        "",
			Compress: false,
			Crypt:    false,
		},
		ConfigConnAllow: true,
		RateLimit:       0,
		MaxConn:         0,
		WebUserName:     "",
		WebPassword:     "",
		MaxTunnelNum:    0,
		Flow: &Flow{
			ExportFlow: 0,
			InletFlow:  0,
			FlowLimit:  0,
		},
		DeviceKey:  verifyKey,
		ProductKey: productKey,
	}
	if err := s.NewClient(t); err != nil {
		err = errors.New("创建verifyKey Client失败")
	}
	return
}

func (s *DbUtils) GetTunnelType(productKey string) (tunnelType *TunnelTypesProductRelation, err error) {
	if v, ok := s.JsonDb.TunnelTypes.Load(productKey); ok {
		tunnelType = v.(*TunnelTypesProductRelation)
		return
	} else {
		return nil, errors.New(fmt.Sprintf("未获取到 %s 对应tunnelTypes", productKey))
	}
}

func (s *DbUtils) GetTunnelList() ([]*TunnelTypesProductRelation, int) {
	var cnt int
	list := make([]*TunnelTypesProductRelation, 0)
	s.JsonDb.TunnelTypes.Range(func(key, value interface{}) bool {
		tempTunnel := value.(*TunnelTypesProductRelation)
		list = append(list, tempTunnel)
		cnt++
		return true
	})
	return list, cnt
}

func (s *DbUtils) NewTunnelType(t *TunnelTypesProductRelation) error {
	s.JsonDb.TunnelTypes.Store(t.ProductKey, t)
	s.JsonDb.StoreTunnelTypeToJsonFile()
	return nil
}

func (s *DbUtils) DelTunnelType(pk string) error {
	s.JsonDb.TunnelTypes.Delete(pk)
	s.JsonDb.StoreTunnelTypeToJsonFile()
	return nil
}

func (s *DbUtils) NewDynamicConfig(t *DynamicConfig) error {
	s.JsonDb.DynamicConfig.Store(t.Key, t)
	s.JsonDb.StoreTunnelTypeToJsonFile()
	return nil
}

func (s *DbUtils) GetConfig(key string) (tunnelType *DynamicConfig, err error) {
	if v, ok := s.JsonDb.DynamicConfig.Load(key); ok {
		tunnelType = v.(*DynamicConfig)
		return
	} else {
		return nil, errors.New(fmt.Sprintf("未获取到 %s 对应DynamicConfig", key))
	}
}

func (s *DbUtils) GetDynamicConfigList() ([]*DynamicConfig, int) {
	var cnt int
	list := make([]*DynamicConfig, 0)
	s.JsonDb.DynamicConfig.Range(func(key, value interface{}) bool {
		dynamicConfig := value.(*DynamicConfig)
		list = append(list, dynamicConfig)
		cnt++
		return true
	})
	return list, cnt
}

func (s *DbUtils) UpdateDynamicConfig(t *DynamicConfig) error {
	s.JsonDb.DynamicConfig.Store(t.Key, t)
	s.JsonDb.StoreDynamicConfigToJsonFile()
	logs.Info("=============更新 动态配置 Key:%s, Value:%s =================", t.Key, t.Value)
	return nil
}
