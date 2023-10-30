package db

import (
	"ehang.io/nps/lib/file"
	. "ehang.io/nps/models"
	"errors"
	"fmt"
	"github.com/astaxie/beego/logs"
	"sync"
)

type TaskDao struct {
	Tasks sync.Map
}

var taskUpdateCols = []string{"client_id", "remark", "mode", "server_ip", "port", "password", "target_addr", "local_path", "strip_pre", "target_str", "is_local_proxy"}

//func (td *TaskDao) LoadTaskFromDB() {
//	list, num := td.GetTunnelAllList()
//	if num == 0 {
//		return
//	}
//	for _, v := range list {
//		td.Tasks.Store(v.Id, v)
//	}
//}

func (td TaskDao) NewTask(t *NpsClientTaskInfo) (err error) {
	value := td.GetTaskByPassword(t.Password)
	if value != nil && (value.Mode == "secret" || value.Mode == "p2p") {
		errorStr := fmt.Sprintf("secret mode keys %s must be unique", t.Password)
		logs.Error(errorStr)
		return errors.New(errorStr)
	}
	td.Save(t)
	return
}

func (TaskDao) GetTaskByPassword(password string) (task *NpsClientTaskInfo) {
	t := new(NpsClientTaskInfo)
	has, dbErr := DbEngine.Where(" password = ?", password).Get(t)
	if !has {
		dbErr = errors.New("根据唯一标识密钥未找到隧道")
	}
	if dbErr != nil {
		logs.Warn(dbErr)
		return
	}
	t.Client = new(NpsClientListInfo)
	t.Client.Id = t.ClientId
	return t
}

// md5 password
func (TaskDao) GetTaskByMd5Password(p string) (task *NpsClientTaskInfo) {
	t := new(NpsClientTaskInfo)
	has, dbErr := DbEngine.Where("MD5(password) = ?", p).Get(t)
	if !has {
		dbErr = errors.New("根据MD5唯一标识密钥未找到隧道")
	}
	if dbErr != nil {
		logs.Error(dbErr)
		return
	}
	t.Client = new(NpsClientListInfo)
	t.Client.Id = t.ClientId
	return t
}

func (TaskDao) Save(t *NpsClientTaskInfo) error {
	if t.Client != nil {
		t.ClientId = t.Client.Id
	}
	_, dbErr := DbEngine.Insert(t)
	if dbErr != nil {
		logs.Error(dbErr)
		return dbErr
	}
	return dbErr
}

// 根据password 删除Task
func (td TaskDao) DelTaskByPassword(password string) (err error) {
	value := td.GetTaskByPassword(password)
	if value != nil {
		td.DelTask(value.Id)
	}
	return
}

func (TaskDao) UpdateTask(t *NpsClientTaskInfo) error {
	_, dbErr := DbEngine.ID(t.Id).Cols(taskUpdateCols...).Update(t)
	if dbErr != nil {
		logs.Error(dbErr)
	}
	return dbErr
}

func (TaskDao) DelTask(id int) error {
	t := new(NpsClientTaskInfo)
	_, dbErr := DbEngine.ID(id).Delete(t)
	if dbErr != nil {
		logs.Error(dbErr)
	}
	return dbErr
}

func (TaskDao) GetTask(id int) (t *NpsClientTaskInfo, err error) {
	tt := new(NpsClientTaskInfo)
	has, err := DbEngine.Where("id = ?", id).Get(tt)
	if !has {
		err = errors.New("not found")
		return tt, err
	}
	return tt, nil
}

func (td TaskDao) HasTunnel(t *NpsClientTaskInfo, clientId int) (exist bool) {
	ti, err := td.GetTunnelByCond(clientId, t.Port)
	if err == nil && ti != nil {
		return true
	}
	return false
}

func (td TaskDao) GetTunnelByCond(clientId int, port int) (t *NpsClientTaskInfo, err error) {
	has, dbErr := DbEngine.Where("client_id = ? AND port !=0 AND port = ?", clientId, port).Get(t)
	if !has {
		dbErr = errors.New("根据条件未找到隧道")
	}
	if dbErr != nil {
		logs.Error(dbErr)
		return
	}
	return
}

func (td TaskDao) GetTunnelNum(clientId int) (num int) {
	_, num = td.GetTunnelListByCond(clientId)
	return num
}

func (td TaskDao) GetTunnel(clientId int) (tunnel *NpsClientTaskInfo, exist bool) {
	list, num := td.GetTunnelListByCond(clientId)
	if num > 0 {
		tunnel = list[0]
		exist = true
	} else {
		tunnel = nil
		exist = false
	}
	return tunnel, exist
}

func (TaskDao) GetTunnelListByCond(clientId int) ([]*NpsClientTaskInfo, int) {
	list := make([]*NpsClientTaskInfo, 0)
	cnt, dbErr := DbEngine.Table("nps_client_task_info").
		Select("nps_client_task_info.id, `mode`, `nps_client_task_info`.`remark`, `server_ip`, `port`, `password`, `ports`, `account`, `target_addr`, `local_path`, `strip_pre`,  `nps_client_task_info`.`status`, `run_status`, `target_str`, `nps_client_task_info`.`create_time`, `nps_client_task_info`.`update_time`,`nps_client_info`.`verify_key`").
		Join("INNER", "nps_client_info", "nps_client_info.Id = `nps_client_task_info`.client_id").
		Where("  `nps_client_task_info`.client_id = ? ", clientId).OrderBy("nps_client_task_info.id").
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

func (TaskDao) GetTunnelAllList() ([]*NpsClientTaskInfo, int) {
	list := make([]*NpsClientTaskInfo, 0)
	cnt, dbErr := DbEngine.Table("nps_client_task_info").
		Select("nps_client_task_info.id, `mode`, `nps_client_task_info`.`remark`, `client_id`, `server_ip`, `port`, `password`, `ports`, `account`, `target_addr`, `local_path`, `strip_pre`, `nps_client_task_info`.`status`, `run_status`, `target_str`, `nps_client_task_info`.`create_time`, `nps_client_task_info`.`update_time`,`nps_client_info`.`verify_key`").
		Join("INNER", "nps_client_info", "nps_client_info.Id = nps_client_task_info.client_id").OrderBy("nps_client_task_info.id").
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

func (TaskDao) GetTunnelList(start, length int, typeVal string, clientId int, search string) ([]*NpsClientTaskListInfo, int) {
	list := make([]*NpsClientTaskListInfo, 0)
	if typeVal == "" {
		if clientId != 0 {
			cnt, dbErr := DbEngine.Table("nps_client_task_info").
				Select("nps_client_task_info.id, `mode`, `nps_client_task_info`.`remark`,`client_id`, `server_ip`, `port`, `password`, `ports`, `account`, `target_addr`, `local_path`, `strip_pre`, `nps_client_task_info`.`status`, `run_status`, `target_str`, `nps_client_task_info`.`create_time`, `nps_client_task_info`.`update_time`,`nps_client_info`.`verify_key`, `nps_client_info`.`basic_auth_user`, `nps_client_info`.`basic_auth_pass`, `nps_client_info`.`is_compress`, `nps_client_info`.`is_crypt`, `nps_client_info`.`is_connect`").
				Join("INNER", "nps_client_info", "nps_client_info.Id = nps_client_task_info.client_id").
				Where(" nps_client_task_info.client_id = ? ", clientId).
				Limit(length, start).
				OrderBy("nps_client_task_info.id").
				FindAndCount(&list)
			if dbErr != nil {
				logs.Error(dbErr)
				return list, 0
			}
			for _, v := range list {
				v.Flow = new(file.Flow)
			}
			return list, int(cnt)
		} else {
			cnt, dbErr := DbEngine.Table("nps_client_task_info").
				Select("nps_client_task_info.id, `mode`, `nps_client_task_info`.`remark`,`client_id`, `server_ip`, `port`, `password`, `ports`, `account`, `target_addr`, `local_path`, `strip_pre`, `nps_client_task_info`.`status`, `run_status`, `target_str`, `nps_client_task_info`.`create_time`, `nps_client_task_info`.`update_time`,`nps_client_info`.`verify_key`, `nps_client_info`.`basic_auth_user`, `nps_client_info`.`basic_auth_pass`, `nps_client_info`.`is_compress`, `nps_client_info`.`is_crypt`, `nps_client_info`.`is_connect`").
				Join("INNER", "nps_client_info", "nps_client_info.Id = nps_client_task_info.client_id").
				Limit(length, start).
				OrderBy("nps_client_task_info.id").
				FindAndCount(&list)
			if dbErr != nil {
				logs.Error(dbErr)
				return list, 0
			}
			for _, v := range list {
				v.Flow = new(file.Flow)
			}
			return list, int(cnt)
		}
	} else {
		if clientId != 0 {
			cnt, dbErr := DbEngine.Table("nps_client_task_info").
				Select("nps_client_task_info.id, `mode`, `nps_client_task_info`.`remark`,`client_id`, `server_ip`, `port`, `password`, `ports`, `account`, `target_addr`, `local_path`, `strip_pre`, `nps_client_task_info`.`status`, `run_status`, `target_str`, `nps_client_task_info`.`create_time`, `nps_client_task_info`.`update_time`,`nps_client_info`.`verify_key`, `nps_client_info`.`basic_auth_user`, `nps_client_info`.`basic_auth_pass`, `nps_client_info`.`is_compress`, `nps_client_info`.`is_crypt`, `nps_client_info`.`is_connect`").
				Join("INNER", "nps_client_info", "nps_client_info.Id = nps_client_task_info.client_id").
				Where("mode =? AND nps_client_task_info.client_id = ? ", typeVal, clientId).
				Limit(length, start).
				OrderBy("nps_client_task_info.id").
				FindAndCount(&list)
			if dbErr != nil {
				logs.Error(dbErr)
				return list, 0
			}
			for _, v := range list {
				v.Flow = new(file.Flow)
			}
			return list, int(cnt)
		} else {
			cnt, dbErr := DbEngine.Table("nps_client_task_info").
				Select("nps_client_task_info.id, `mode`, `nps_client_task_info`.`remark`,`client_id`, `server_ip`, `port`, `password`, `ports`, `account`, `target_addr`, `local_path`, `strip_pre`, `nps_client_task_info`.`status`, `run_status`, `target_str`, `nps_client_task_info`.`create_time`, `nps_client_task_info`.`update_time`,`nps_client_info`.`verify_key`, `nps_client_info`.`basic_auth_user`, `nps_client_info`.`basic_auth_pass`, `nps_client_info`.`is_compress`, `nps_client_info`.`is_crypt`, `nps_client_info`.`is_connect`").
				Join("INNER", "nps_client_info", "nps_client_info.Id = nps_client_task_info.client_id").
				Where("mode =? ", typeVal).
				Limit(length, start).
				OrderBy("nps_client_task_info.id").
				FindAndCount(&list)
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

	return list, 0
}
func (td TaskDao) UpdateTaskStatus(t *NpsClientTaskInfo) error {
	_, dbErr := DbEngine.ID(t.Id).Cols("status").Update(t)
	if dbErr != nil {
		logs.Error(dbErr)
	}
	return dbErr
}

func (td TaskDao) UpdateTaskRunStatus(t *NpsClientTaskInfo) error {
	_, dbErr := DbEngine.ID(t.Id).Cols("run_status").Update(t)
	if dbErr != nil {
		logs.Error(dbErr)
	}
	return dbErr
}
