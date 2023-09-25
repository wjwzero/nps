package controllers

import (
	"ehang.io/nps/lib/common"
	"ehang.io/nps/lib/file"
	"ehang.io/nps/server"
	"strings"
)

type TunnelController struct {
	BaseController
}

//修改Tunnel
func (s *TunnelController) Edit() {
	pk := s.getEscapeString("productKey")
	if s.Ctx.Request.Method == "GET" {
		s.Data["menu"] = "tunnel"
		if c, err := file.GetDb().GetTunnelType(pk); err != nil {
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

func (s *TunnelController) Tunnellist() {
	list, cnt := server.GetTunnelList()
	cmd := make(map[string]interface{})
	//ip := s.Ctx.Request.Host
	//cmd["ip"] = common.GetIpByAddr(ip)
	//cmd["bridgeType"] = beego.AppConfig.String("bridge_type")
	//cmd["bridgePort"] = server.Bridge.TunnelPort
	s.AjaxTable(list, cnt, cnt, cmd)
}

//添加客户端
func (s *TunnelController) Add() {
	if s.Ctx.Request.Method == "GET" {
		s.Data["menu"] = "client"
		s.SetInfo("add client")
		s.display()
	} else {
		s.Save()
	}
}

func (s *TunnelController) Save() {
	tunnelTypeStr := s.getEscapeString("tunnelType")
	tunnelTypeArr := strings.Split(tunnelTypeStr, ",")
	for _, tunnelType := range tunnelTypeArr {
		switch tunnelType {
		case common.LAN_TUNNEL_TYPE, common.P2P_TUNNEL_TYPE, common.RELAY_TUNNEL_TYPE:
			break
		default:
			s.AjaxErr("must in LAN,P2P,Relay")
			return
		}
	}
	t := &file.TunnelTypesProductRelation{
		ProductKey:  s.getEscapeString("productKey"),
		TunnelTypes: tunnelTypeStr,
	}
	if err := file.GetDb().NewTunnelType(t); err != nil {
		s.AjaxErr(err.Error())
	}
	s.AjaxOk("add success")
}

//删除Tunnel
func (s *TunnelController) Del() {
	pk := s.getEscapeString("productKey")
	file.GetDb().DelTunnelType(pk)
	s.AjaxOk("delete success")
}
