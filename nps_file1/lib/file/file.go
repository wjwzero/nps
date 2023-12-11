package file

import (
	"encoding/json"
	"errors"
	"github.com/astaxie/beego/logs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"ehang.io/nps/lib/common"
	"ehang.io/nps/lib/rate"
)

func NewJsonDb(runPath string) *JsonDb {
	return &JsonDb{
		RunPath:               runPath,
		TaskFilePath:          filepath.Join(runPath, "conf", "tasks.json"),
		HostFilePath:          filepath.Join(runPath, "conf", "hosts.json"),
		ClientFilePath:        filepath.Join(runPath, "conf", "clients.json"),
		TunnelTypesFilePath:   filepath.Join(runPath, "conf", "tunnelTypes.json"),
		DynamicConfigFilePath: filepath.Join(runPath, "conf", "dynamicConfig.json"),
	}
}

type JsonDb struct {
	Tasks                 sync.Map
	Hosts                 sync.Map
	HostsTmp              sync.Map
	Clients               sync.Map
	TunnelTypes           sync.Map
	DynamicConfig         sync.Map
	RunPath               string
	ClientIncreaseId      int32  //client increased id
	TaskIncreaseId        int32  //task increased id
	HostIncreaseId        int32  //host increased id
	TaskFilePath          string //task file path
	HostFilePath          string //host file path
	ClientFilePath        string //client file path
	TunnelTypesFilePath   string //tunnelTypes file path
	DynamicConfigFilePath string //tunnelTypes file path
}

func (s *JsonDb) LoadTaskFromJsonFile() {
	loadSyncMapFromFile(s.TaskFilePath, func(v string) {
		var err error
		post := new(Tunnel)
		if json.Unmarshal([]byte(v), &post) != nil {
			return
		}
		if post.Client, err = s.GetClient(post.Client.Id); err != nil {
			return
		}
		s.Tasks.Store(post.Id, post)
		if post.Id > int(s.TaskIncreaseId) {
			s.TaskIncreaseId = int32(post.Id)
		}
	})
}

func (s *JsonDb) LoadClientFromJsonFile() {
	loadSyncMapFromFile(s.ClientFilePath, func(v string) {
		post := new(Client)
		if json.Unmarshal([]byte(v), &post) != nil {
			return
		}
		rateFlag := common.GetRateFlag()
		if post.RateLimit > 0 && rateFlag {
			post.Rate = rate.NewRate(int64(post.RateLimit * 1024))
		} else {
			var rateMaxErr error
			var rateLimit int64
			if rateLimit, rateMaxErr = s.GetCommonRateLimitMax(); rateMaxErr != nil {
				logs.Error("获取动态配置失败 GetCommonRateLimitMax 默认 %s kb", rateLimit, rateMaxErr.Error())
			}
			post.Rate = rate.NewRate(int64(rateLimit * 1024))
		}
		post.Rate.Start()
		post.NowConn = 0
		s.Clients.Store(post.Id, post)
		if post.Id > int(s.ClientIncreaseId) {
			s.ClientIncreaseId = int32(post.Id)
		}
	})
}

func (s *JsonDb) LoadHostFromJsonFile() {
	loadSyncMapFromFile(s.HostFilePath, func(v string) {
		var err error
		post := new(Host)
		if json.Unmarshal([]byte(v), &post) != nil {
			return
		}
		if post.Client, err = s.GetClient(post.Client.Id); err != nil {
			return
		}
		s.Hosts.Store(post.Id, post)
		if post.Id > int(s.HostIncreaseId) {
			s.HostIncreaseId = int32(post.Id)
		}
	})
}

func (s *JsonDb) LoadTunnelTypesProductRelationFromJsonFile() {
	loadSyncMapFromFile(s.TunnelTypesFilePath, func(v string) {
		tunnelTypes := new(TunnelTypesProductRelation)
		if json.Unmarshal([]byte(v), &tunnelTypes) != nil {
			return
		}
		s.TunnelTypes.Store(tunnelTypes.ProductKey, tunnelTypes)
	})
}

func (s *JsonDb) LoadDynamicConfigFromJsonFile() {
	loadSyncMapFromFile(s.DynamicConfigFilePath, func(v string) {
		dynamicConfig := new(DynamicConfig)
		if json.Unmarshal([]byte(v), &dynamicConfig) != nil {
			return
		}
		s.DynamicConfig.Store(dynamicConfig.Key, dynamicConfig)
		logs.Info("=============初始化 动态配置 Key:%s, Value:%s =================", dynamicConfig.Key, dynamicConfig.Value)
	})
}

func (s *JsonDb) GetClient(id int) (c *Client, err error) {
	if v, ok := s.Clients.Load(id); ok {
		c = v.(*Client)
		return
	}
	err = errors.New("未找到客户端")
	return
}

var hostLock sync.Mutex

func (s *JsonDb) StoreHostToJsonFile() {
	hostLock.Lock()
	storeSyncMapToFile(s.Hosts, s.HostFilePath)
	hostLock.Unlock()
}

var taskLock sync.Mutex

func (s *JsonDb) StoreTasksToJsonFile() {
	taskLock.Lock()
	storeSyncMapToFile(s.Tasks, s.TaskFilePath)
	taskLock.Unlock()
}

var clientLock sync.Mutex

func (s *JsonDb) StoreClientsToJsonFile() {
	clientLock.Lock()
	storeSyncMapToFile(s.Clients, s.ClientFilePath)
	clientLock.Unlock()
}

var tunnelTypeLock sync.Mutex

func (s *JsonDb) StoreTunnelTypeToJsonFile() {
	tunnelTypeLock.Lock()
	storeSyncMapToFile(s.TunnelTypes, s.TunnelTypesFilePath)
	tunnelTypeLock.Unlock()
}

var dynamicConfigLock sync.Mutex

func (s *JsonDb) StoreDynamicConfigToJsonFile() {
	dynamicConfigLock.Lock()
	storeSyncMapToFile(s.DynamicConfig, s.DynamicConfigFilePath)
	dynamicConfigLock.Unlock()
}

func (s *JsonDb) GetClientId() int32 {
	return atomic.AddInt32(&s.ClientIncreaseId, 1)
}

func (s *JsonDb) GetTaskId() int32 {
	return atomic.AddInt32(&s.TaskIncreaseId, 1)
}

func (s *JsonDb) GetHostId() int32 {
	return atomic.AddInt32(&s.HostIncreaseId, 1)
}

func loadSyncMapFromFile(filePath string, f func(value string)) {
	b, err := common.ReadAllFromFile(filePath)
	if err != nil {
		panic(err)
	}
	for _, v := range strings.Split(string(b), "\n"+common.CONN_DATA_SEQ) {
		f(v)
	}
}

func storeSyncMapToFile(m sync.Map, filePath string) {
	file, err := os.Create(filePath + ".tmp")
	// first create a temporary file to store
	if err != nil {
		panic(err)
	}
	m.Range(func(key, value interface{}) bool {
		var b []byte
		var err error
		switch value.(type) {
		case *Tunnel:
			obj := value.(*Tunnel)
			if obj.NoStore {
				return true
			}
			b, err = json.Marshal(obj)
		case *Host:
			obj := value.(*Host)
			if obj.NoStore {
				return true
			}
			b, err = json.Marshal(obj)
		case *Client:
			obj := value.(*Client)
			if obj.NoStore {
				return true
			}
			b, err = json.Marshal(obj)
		case *TunnelTypesProductRelation:
			obj := value.(*TunnelTypesProductRelation)
			b, err = json.Marshal(obj)
		default:
			return true
		}
		if err != nil {
			return true
		}
		_, err = file.Write(b)
		if err != nil {
			panic(err)
		}
		_, err = file.Write([]byte("\n" + common.CONN_DATA_SEQ))
		if err != nil {
			panic(err)
		}
		return true
	})
	_ = file.Sync()
	_ = file.Close()
	// must close file first, then rename it
	err = os.Rename(filePath+".tmp", filePath)
	if err != nil {
		logs.Error(err, "store to file err, data will lost")
	}
	// replace the file, maybe provides atomic operation
}

func (s *JsonDb) GetDynamicConfig(key string) (c *DynamicConfig, err error) {
	if v, ok := s.DynamicConfig.Load(key); ok {
		c = v.(*DynamicConfig)
		return
	}
	err = errors.New("未找到配置")
	return
}

func (s *JsonDb) CheckRateLimitFlag() (rateLimit bool, err error) {
	var d *DynamicConfig
	if d, err = s.GetDynamicConfig("RateLimit"); err == nil {
		rateLimitStr := d.Value
		rateLimit, err = strconv.ParseBool(rateLimitStr)
		return
	}
	return
}

func (s *JsonDb) GetCommonRateLimitMax() (rateLimitMax int64, err error) {
	var d *DynamicConfig
	rateLimitMax = 50000
	if d, err = s.GetDynamicConfig("RateLimitMax"); err == nil {
		rateLimitMaxStr := d.Value
		rateLimitMax, err = strconv.ParseInt(rateLimitMaxStr, 10, 64)
		return
	}
	return
}
