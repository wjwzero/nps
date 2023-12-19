package controllers

import (
	"ehang.io/nps/db"
	"ehang.io/nps/lib/common"
	"ehang.io/nps/lib/crypt"
	"ehang.io/nps/lib/file"
	"ehang.io/nps/lib/rate"
	. "ehang.io/nps/models"
	"ehang.io/nps/server"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

type ClientController struct {
	BaseController
	ClientDao db.ClientDao
}

func (s *ClientController) List() {
	if s.Ctx.Request.Method == "GET" {
		s.Data["menu"] = "client"
		s.SetInfo("client")
		s.display("client/list")
		return
	}
	start, length := s.GetAjaxParams()
	clientIdSession := s.GetSession("clientId")
	var clientId int
	if clientIdSession == nil {
		clientId = 0
	} else {
		clientId = clientIdSession.(int)
	}
	list, cnt := s.ClientDao.GetClientListInfo(start, length, s.getEscapeString("search"), s.getEscapeString("sort"), s.getEscapeString("order"), clientId)
	cmd := make(map[string]interface{})
	ip := s.Ctx.Request.Host
	cmd["ip"] = common.GetIpByAddr(ip)
	cmd["bridgeType"] = beego.AppConfig.String("bridge_type")
	cmd["bridgePort"] = server.Bridge.TunnelPort
	s.AjaxTable(list, cnt, cnt, cmd)
}

// 添加客户端
func (s *ClientController) Add() {
	if s.Ctx.Request.Method == "GET" {
		s.Data["menu"] = "client"
		s.SetInfo("add client")
		s.display()
	} else {
		t := &NpsClientInfo{
			VerifyKey:         s.getEscapeString("vkey"),
			ProductKey:        s.getEscapeString("productKey"),
			Remark:            s.getEscapeString("remark"),
			Status:            true,
			IsCompress:        common.GetBoolByStr(s.getEscapeString("compress")),
			IsCrypt:           s.GetBoolNoErr("crypt"),
			IsConfigConnAllow: s.GetBoolNoErr("config_conn_allow"),
			BasicAuthUser:     s.getEscapeString("u"),
			BasicAuthPass:     s.getEscapeString("p"),
			RateLimit:         s.GetIntNoErr("rate_limit"),
			MaxChannelNum:     s.GetIntNoErr("max_tunnel"),
			MaxConnectNum:     s.GetIntNoErr("max_conn"),
			WebUser:           s.getEscapeString("web_username"),
			WebPass:           s.getEscapeString("web_password"),
			FlowLimit:         int64(s.GetIntNoErr("flow_limit")),
		}
		if err := s.ClientDao.NewClient(t); err != nil {
			s.AjaxErr(err.Error())
		}
		s.AjaxOk("add success")
	}
}
func (s *ClientController) GetClient() {
	if s.Ctx.Request.Method == "POST" {
		id := s.GetIntNoErr("id")
		data := make(map[string]interface{})
		if c, err := s.ClientDao.GetClient(id); err != nil {
			data["code"] = 0
		} else {
			data["code"] = 1
			data["data"] = c
		}
		s.Data["json"] = data
		s.ServeJSON()
	}
}

// 修改客户端
func (s *ClientController) Edit() {
	id := s.GetIntNoErr("id")
	if s.Ctx.Request.Method == "GET" {
		s.Data["menu"] = "client"
		if c, err := s.ClientDao.GetClient(id); err != nil {
			s.error()
		} else {
			s.Data["c"] = c
		}
		s.SetInfo("edit client")
		s.display()
	} else {
		if c, err := s.ClientDao.GetClient(id); err != nil {
			s.error()
			s.AjaxErr("client ID not found")
			return
		} else {
			if s.getEscapeString("web_username") != "" {
				if s.getEscapeString("web_username") == beego.AppConfig.String("web_username") || !s.ClientDao.VerifyUserName(s.getEscapeString("web_username"), c.Id) {
					s.AjaxErr("web login username duplicate, please reset")
					return
				}
			}
			if s.GetSession("isAdmin").(bool) {
				c.VerifyKey = s.getEscapeString("vkey")
				if c.VerifyKey == "" {
					c.VerifyKey = crypt.GetRandomString(16)
				}
				if !s.ClientDao.VerifyVkey(c.VerifyKey, c.Id) {
					s.AjaxErr("Vkey duplicate, please reset")
					return
				}
				c.FlowLimit = int64(s.GetIntNoErr("flow_limit"))
				c.RateLimit = s.GetIntNoErr("rate_limit")
				c.MaxConnectNum = s.GetIntNoErr("max_conn")
				c.MaxChannelNum = s.GetIntNoErr("max_tunnel")
			}
			c.ProductKey = s.getEscapeString("productKey")
			c.Remark = s.getEscapeString("remark")
			c.BasicAuthPass = s.getEscapeString("u")
			c.BasicAuthPass = s.getEscapeString("p")
			c.IsCompress = common.GetBoolByStr(s.getEscapeString("compress"))
			c.IsCrypt = s.GetBoolNoErr("crypt")
			b, err := beego.AppConfig.Bool("allow_user_change_username")
			if s.GetSession("isAdmin").(bool) || (err == nil && b) {
				c.WebUser = s.getEscapeString("web_username")
			}
			c.WebPass = s.getEscapeString("web_password")
			c.IsConfigConnAllow = s.GetBoolNoErr("config_conn_allow")
			if c.Rate != nil {
				c.Rate.Stop()
			}
			rateFlag := common.GetRateFlag()
			if c.RateLimit > 0 && rateFlag {
				c.Rate = rate.NewRate(int64(c.RateLimit * 1024))
				c.Rate.Start()
			} else {
				var rateMaxErr error
				var rateLimit int64
				if rateLimit, rateMaxErr = file.GetDb().JsonDb.GetCommonRateLimitMax(); rateMaxErr != nil {
					logs.Error("获取动态配置失败 GetCommonRateLimitMax 默认 %s kb", rateLimit, rateMaxErr.Error())
					s.AjaxErr("commonRateLimit 获取失败, please reset")
					return
				}
				c.Rate = rate.NewRate(int64(rateLimit * 1024))
				c.Rate.Start()
			}
			s.ClientDao.SaveEdit(c)
		}
		s.AjaxOk("save success")
	}
}

// 更改状态
func (s *ClientController) ChangeStatus() {
	id := s.GetIntNoErr("id")
	if client, err := s.ClientDao.GetClient(id); err == nil {
		client.Status = s.GetBoolNoErr("status")
		err = s.ClientDao.UpdateStatus(client)
		if err == nil {
			if client.Status == false {
				server.DelClientConnect(client.Id)
			}
			s.AjaxOk("modified success")
		}
	}
	s.AjaxErr("modified fail")
}

// 删除客户端
func (s *ClientController) Del() {
	id := s.GetIntNoErr("id")
	server.DelClientConnect(id)
	if err := s.ClientDao.DelClient(id); err != nil {
		s.AjaxErr("delete error")
	}
	s.AjaxOk("delete success")
}
