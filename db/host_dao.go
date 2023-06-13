package db

import (
	"ehang.io/nps/lib/common"
	"ehang.io/nps/lib/file"
	. "ehang.io/nps/models"
	"errors"
	"github.com/astaxie/beego/logs"
	"net/http"
	"strings"
	"sync"
)

type HostDao struct {
	Hosts sync.Map
}

var hostUpdateCols = []string{"host", "location", "scheme", "target_str", "remark", "host_change", "header_change", "is_local_proxy", "key_file_path", "cert_file_path"}

func (hd *HostDao) LoadHostFromDb() {
	list, num := hd.GetHostAllList()
	if num == 0 {
		return
	}
	for _, v := range list {
		hd.Hosts.Store(v.Id, v)
	}
}

func (HostDao) DelHost(id int) error {
	h := new(NpsClientHostInfo)
	_, dbErr := DbEngine.ID(id).Delete(h)
	if dbErr != nil {
		logs.Error(dbErr)
	}
	return dbErr
}

func (hd HostDao) NewHost(h *NpsClientHostInfo) error {
	if h.Location == "" {
		h.Location = "/"
	}
	if hd.IsHostExist(h) {
		return errors.New("host has exist")
	}
	hd.Save(h)
	return nil
}

func (hd HostDao) IsHostExist(h *NpsClientHostInfo) bool {
	hi := hd.getHostByCond(h.Host, h.Location, h.Scheme)
	if hi != nil && hi.Id != h.Id {
		return true
	}
	return false
}

func (HostDao) Save(h *NpsClientHostInfo) error {
	_, dbErr := DbEngine.Insert(h)
	if dbErr != nil {
		logs.Error(dbErr)
		return dbErr
	}
	return dbErr
}

func (HostDao) SaveEdit(h *NpsClientHostInfo) error {
	_, dbErr := DbEngine.ID(h.Id).Cols(hostUpdateCols...).Update(h)
	if dbErr != nil {
		logs.Error(dbErr)
	}
	return dbErr
}

func (hd HostDao) getHostByCond(host string, location string, scheme string) (h *NpsClientHostInfo) {
	has, dbErr := DbEngine.Where("host = ? AND location = ？AND (scheme = 'all' OR scheme = ?)", host, location, scheme).Get(h)
	if !has {
		dbErr = errors.New("根据条件未找到主机")
	}
	if dbErr != nil {
		logs.Error(dbErr)
		return
	}
	return
}

func (HostDao) GetHostById(id int) (h *NpsClientHostInfo, err error) {
	tt := new(NpsClientHostInfo)
	has, err := DbEngine.Where("id = ?", id).Get(tt)
	if !has {
		err = errors.New("The host could not be parsed")
		return tt, err
	}
	return tt, nil
}

// get key by host from x
func (hd HostDao) GetInfoByHost(host string, r *http.Request) (h *NpsClientHostInfo, err error) {
	var hosts []*NpsClientHostInfo
	//Handling Ported Access
	host = common.GetIpByAddr(host)
	hd.LoadHostFromDb()
	hd.Hosts.Range(func(key, value interface{}) bool {
		v := value.(*NpsClientHostInfo)
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

func (HostDao) GetHostList(start, length int, clientId int, search string, hostId int, host string) ([]*NpsClientHostListInfo, int) {
	list := make([]*NpsClientHostListInfo, 0)
	if clientId != 0 {
		cnt, dbErr := DbEngine.Table("nps_client_host_info").
			Select("nps_client_host_info.id, `host`,`client_id`, `host_change`,`header_change`, `location`, `nps_client_host_info`.`no_store`, `is_close`, `scheme`, `target_str`, `nps_client_host_info`.`remark`, `is_local_proxy`, `nps_client_host_info`.`create_time`, `nps_client_host_info`.`update_time`,`nps_client_info`.`verify_key`, `nps_client_info`.`basic_auth_user`, `nps_client_info`.`basic_auth_pass`, `nps_client_info`.`is_compress`, `nps_client_info`.`is_crypt`, `nps_client_info`.`is_connect`").
			Join("INNER", "nps_client_info", "nps_client_info.Id = nps_client_host_info.client_id").
			Where("nps_client_host_info.client_id = ?", clientId).
			Limit(length, start).OrderBy("nps_client_host_info.id").FindAndCount(&list)
		if dbErr != nil {
			logs.Error(dbErr)
			return list, 0
		}
		for _, v := range list {
			v.Flow = new(file.Flow)
		}
		return list, int(cnt)
	} else {
		cnt, dbErr := DbEngine.Table("nps_client_host_info").
			Select("nps_client_host_info.id, `host`,`client_id`, `host_change`,`header_change`, `location`, `nps_client_host_info`.`no_store`, `is_close`, `scheme`, `target_str`, `nps_client_host_info`.`remark`, `is_local_proxy`, `nps_client_host_info`.`create_time`, `nps_client_host_info`.`update_time`,`nps_client_info`.`verify_key`, `nps_client_info`.`basic_auth_user`, `nps_client_info`.`basic_auth_pass`, `nps_client_info`.`is_compress`, `nps_client_info`.`is_crypt`, `nps_client_info`.`is_connect`").
			Join("INNER", "nps_client_info", "nps_client_info.Id = nps_client_host_info.client_id").
			Limit(length, start).OrderBy("nps_client_host_info.id").FindAndCount(&list)
		if dbErr != nil {
			logs.Error(dbErr)
			return list, 0
		}
		for _, v := range list {
			v.Flow = new(file.Flow)
		}
		return list, int(cnt)
	}
}

func (HostDao) GetHostAllList() ([]*NpsClientHostInfo, int) {
	list := make([]*NpsClientHostInfo, 0)
	cnt, dbErr := DbEngine.Table("nps_client_host_info").
		Select("nps_client_host_info.id, `host`,`client_id`, `host_change`, `location`, `nps_client_host_info`.`no_store`, `is_close`, `scheme`, `target_str`, `nps_client_host_info`.`remark`, `is_local_proxy`, `nps_client_host_info`.`create_time`, `nps_client_host_info`.`update_time`,`nps_client_info`.`verify_key`").
		Join("INNER", "nps_client_info", "nps_client_info.Id = nps_client_host_info.client_id").
		FindAndCount(&list)
	if dbErr != nil {
		logs.Error(dbErr)
		return list, 0
	}
	for _, v := range list {
		v.Flow = new(file.Flow)
		v.Client = new(NpsClientListInfo)
		v.Client.Id = v.ClientId
	}
	return list, int(cnt)
}

func (HostDao) GetHostAllListByCond(clientId int) ([]*NpsClientHostInfo, int) {
	list := make([]*NpsClientHostInfo, 0)
	cnt, dbErr := DbEngine.Table("nps_client_host_info").
		Select("nps_client_host_info.id, `host`,`client_id`, `host_change`, `location`, `nps_client_host_info`.`no_store`, `is_close`, `scheme`, `target_str`, `nps_client_host_info`.`remark`, `is_local_proxy`, `nps_client_host_info`.`create_time`, `nps_client_host_info`.`update_time`,`nps_client_info`.`verify_key`").
		Join("INNER", "nps_client_info", "nps_client_info.Id = nps_client_host_info.client_id").
		Where(" nps_client_host_info.client_id = ?", clientId).
		FindAndCount(&list)
	if dbErr != nil {
		logs.Error(dbErr)
		return list, 0
	}
	for _, v := range list {
		v.Flow = new(file.Flow)
		v.Client = new(NpsClientListInfo)
		v.Client.Id = v.ClientId
	}
	return list, int(cnt)
}

func (hd HostDao) HasHost(h *NpsClientHostInfo, clientId int) bool {
	var has bool
	_, num := hd.GetHostSByCond(clientId, h.Host, h.Location)
	if num > 0 {
		has = true
	}
	return has
}

func (HostDao) GetHostSByCond(clientId int, host string, location string) ([]*NpsClientHostInfo, int) {
	list := make([]*NpsClientHostInfo, 0)
	cnt, dbErr := DbEngine.Table("nps_client_host_info").
		Select("nps_client_host_info.id, `host`, `host_change`, `location`, `no_store`, `is_close`, `scheme`, `target_str`, `remark`, `is_local_proxy`, `nps_client_host_info`.`create_time`, `nps_client_host_info`.`update_time`,`nps_client_info`.'verify_key'").
		Join("INNER", "nps_client_info", "nps_client_info.Id = nps_client_host_info.client_id").
		Where(" nps_client_host_info.client_id = ? AND host = ？ AND location = ？", clientId, host, location).FindAndCount(&list)
	if dbErr != nil {
		logs.Error(dbErr)
		return list, 0
	}
	return list, int(cnt)
}
