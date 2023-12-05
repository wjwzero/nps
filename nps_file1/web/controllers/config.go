package controllers

import (
	"ehang.io/nps/lib/file"
	"ehang.io/nps/server"
)

type ConfigController struct {
	BaseController
}

//修改Tunnel
func (s *ConfigController) Edit() {
	key := s.getEscapeString("key")
	if s.Ctx.Request.Method == "GET" {
		s.Data["menu"] = "config"
		if c, err := file.GetDb().GetConfig(key); err != nil {
			s.error()
		} else {
			s.Data["c"] = c
		}
		s.SetInfo("edit client")
		s.display()
	} else {
		s.Save()
		s.AjaxOk("save success")
	}
}

func (s *ConfigController) Configlist() {
	list, cnt := server.GetConfigList()
	cmd := make(map[string]interface{})
	//ip := s.Ctx.Request.Host
	//cmd["ip"] = common.GetIpByAddr(ip)
	//cmd["bridgeType"] = beego.AppConfig.String("bridge_type")
	//cmd["bridgePort"] = server.Bridge.TunnelPort
	s.AjaxTable(list, cnt, cnt, cmd)
}

//添加客户端
func (s *ConfigController) Add() {
	if s.Ctx.Request.Method == "GET" {
		s.Data["menu"] = "config"
		s.SetInfo("add config")
		s.display()
	} else {
		s.Save()
	}
}

func (s *ConfigController) Save() {
	key := s.getEscapeString("key")
	t := &file.DynamicConfig{
		Key:   key,
		Value: s.getEscapeString("value"),
	}
	if err := file.GetDb().NewDynamicConfig(t); err != nil {
		s.AjaxErr(err.Error())
	}
	s.AjaxOk("edit success")
}
