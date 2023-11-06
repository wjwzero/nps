package cloud

import (
	"ehang.io/nps/lib/crypt"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"
)

type CommonData struct {
	Code    int                    `json:"code"`
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data"`
	Msg     string                 `json:"msg"`
}

// GetNpsNodeExternalIp 从云平台获得节点外部IP地址
func GetNpsNodeExternalIp(cloudAddr string, deviceKey string) (ip string, err error) {
	var retryNum = 0
	var retryDuration = 10
retry:
	ip, err = getNpsNodeExternalIp(cloudAddr, deviceKey)
	if err != nil {
		retryNum++
		if retryNum < 3 {
			retryDuration = 100 + rand.Intn(30)
		} else if retryNum >= 3 && retryNum < 5 {
			retryDuration = 150 + rand.Intn(30)
		} else if retryNum >= 5 && retryNum < 10 {
			retryDuration = 300 + rand.Intn(30)
		} else {
			retryDuration = 9*60 + rand.Intn(60)
		}
		logs.Warn(fmt.Sprintf("从云平台获得NPS节点地址失败, %d s后重试 %s", retryDuration, err.Error()))
		time.Sleep(time.Duration(retryDuration) * time.Second)
		goto retry
	}
	return
}

// getNpsNodeExternalIp 从云平台获得节点外部IP地址 client 使用
func getNpsNodeExternalIp(cloudAddr string, deviceKey string) (ip string, err error) {
	var data CommonData
	resp, err := http.Get(cloudAddr + "/ip?deviceKey=" + deviceKey)
	if err != nil {
		return "", err
	}
	content, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	err = json.Unmarshal(content, &data)
	if data.Code != 200 {
		return "", errors.New(fmt.Sprintf("code：%d;%s", data.Code, data.Msg))
	}
	ip = data.Data["ip"].(string)
	return
}

// CheckPassword 从云平台验证password server 使用
func CheckPassword(cloudAddr string, vKey string, password string) (res bool, err error) {
	sign := crypt.CreateSign([]string{vKey, password})
	var data CommonData
	cloudAddr = beego.AppConfig.String("cloudAddr")
	resp, err := http.Get(cloudAddr + "/judgeNasPassword?password=" + password + "&deviceKey=" + vKey + "&authString=" + sign)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	content, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(content, &data)
	if err != nil {
		return false, err
	}
	if data.Code != 200 {
		return false, errors.New(fmt.Sprintf("data.Code %d data.Msg %s", data.Code, data.Msg))
	}
	res = data.Data["judge"].(bool)
	return
}

// CheckDeviceKey 从云平台验证vKey dk server 使用
func CheckDeviceKey(cloudAddr string, deviceKey string) (res bool, productKey string, err error) {
	sign := crypt.CreateSign([]string{deviceKey})
	var data CommonData
	resp, err := http.Get(cloudAddr + "/judgeDeviceExist?deviceKey=" + deviceKey + "&authString=" + sign)
	if err != nil {
		return false, "", err
	}
	defer resp.Body.Close()
	content, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(content, &data)
	if err != nil {
		return false, "", err
	}
	if data.Code != 200 {
		return false, "", errors.New(fmt.Sprintf("data.Code %d data.Msg %s", data.Code, data.Msg))
	}
	res = data.Data["judge"].(bool)
	if data.Data["productKey"] != nil {
		productKey = data.Data["productKey"].(string)
	}
	return
}
