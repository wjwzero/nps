package controllers

import (
	"ehang.io/nps/db"
	"ehang.io/nps/lib/crypt"
	"ehang.io/nps/lib/precreatetask"
	"ehang.io/nps/models"
	"fmt"
	"github.com/astaxie/beego/validation"
)

type ReportController struct {
	BaseController
	ClientDao db.ClientDao
	TaskDao   db.TaskDao
}

type ReportBean struct {
	DeviceKey string `json:"deviceKey"`
	Account   string `json:"account"`
	Sign      string `json:"sign"`
}

type CreateNpsClientBean struct {
	DeviceKey  string `json:"deviceKey"`
	Password   string `json:"password"`
	Sign       string `json:"sign"`
	ProductKey string `json:"productKey"`
}

// CreateNpsClient 云平台创建连接客户端
func (s *ReportController) CreateNpsClient() {
	var bean CreateNpsClientBean
	valid := validation.Validation{}
	err := s.getRequestBodyBean(&bean)
	if err != nil {
		s.AjaxCloudErr("request body error", 400)
	}
	if s.Ctx.Request.Method != "POST" {
		s.AjaxCloudErr("request method Post", 400)
	}
	deviceKey := bean.DeviceKey
	if v := valid.Required(deviceKey, "deviceKey"); !v.Ok {
		s.AjaxCloudErr("deviceKey 不应为空", 400)
	}
	password := bean.Password
	if v := valid.Required(password, "password"); !v.Ok {
		s.AjaxCloudErr("password 不应为空", 400)
	}
	sign := bean.Sign
	trueSign := crypt.CreateSign([]string{deviceKey, password})
	if trueSign != sign {
		s.AjaxCloudErr("sign 不正确", 400)
	}
	productKey := bean.ProductKey
	// 检测此deviceKey是否已注册
	client := s.ClientDao.GetClientByDeviceKey(deviceKey)
	if client != nil {
		deviceKey = client.VerifyKey
		if err = precreatetask.P2pClient(client.Id, deviceKey, password); err != nil {
			s.AjaxCloudErr(err.Error(), 1400)
		}
	} else {
		t := &models.NpsClientInfo{
			VerifyKey:         deviceKey,
			Status:            true,
			Remark:            deviceKey,
			IsCrypt:           false,
			IsCompress:        false,
			BasicAuthPass:     "",
			BasicAuthUser:     "",
			IsConfigConnAllow: true,
			RateLimit:         0,
			MaxConnectNum:     0,
			WebUser:           "",
			WebPass:           "",
			MaxChannelNum:     0,
			FlowLimit:         0,
			DeviceKey:         deviceKey,
			ProductKey:        productKey,
		}
		if err := s.ClientDao.NewClient(t); err != nil {
			s.AjaxCloudErr("创建被代理客户端数据失败， "+err.Error(), 1401)
		}
		if err = precreatetask.P2pClient(t.Id, deviceKey, password); err != nil {
			s.AjaxCloudErr("创建代理客户端数据失败， "+err.Error(), 1402)
		}
	}
	s.AjaxCloudOk("create npc success")
}

type DeleteNpsClientBean struct {
	DeviceKey string `json:"deviceKey"`
	Password  string `json:"password"`
	Sign      string `json:"sign"`
}

// DeleteNpsClient 云平台解除绑定分享连接客户端
func (s *ReportController) DeleteNpsClient() {
	var bean DeleteNpsClientBean
	valid := validation.Validation{}
	err := s.getRequestBodyBean(&bean)
	if err != nil {
		s.AjaxCloudErr("request body error", 400)
	}
	if s.Ctx.Request.Method != "POST" {
		s.AjaxCloudErr("request method Post", 400)
	}
	deviceKey := bean.DeviceKey
	if v := valid.Required(deviceKey, "deviceKey"); !v.Ok {
		s.AjaxCloudErr("deviceKey 不应为空", 400)
	}
	password := bean.Password
	if v := valid.Required(password, "password"); !v.Ok {
		s.AjaxCloudErr("password 不应为空", 400)
	}
	sign := bean.Sign
	trueSign := crypt.CreateSign([]string{deviceKey, password})
	if trueSign != sign {
		s.AjaxCloudErr("sign 不正确", 400)
	}
	// 检测此deviceKey是否已注册
	client := s.ClientDao.GetClientByDeviceKey(deviceKey)
	if client != nil {
		if err := s.TaskDao.DelTaskByPassword(password); err != nil {
			s.AjaxCloudErr(fmt.Sprintf("根据deviceKey %s， password %s 删除task失败,err: %s", deviceKey, password, err.Error()), 1501)
		}
		s.AjaxCloudOk("delete npc success")
	} else {
		s.AjaxCloudErr(fmt.Sprintf("根据deviceKey %s， password %s 删除task失败, 未查到deviceKey对应数据,err: %s", deviceKey, password, err.Error()), 1502)
	}
}
