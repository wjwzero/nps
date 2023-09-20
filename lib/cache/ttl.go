package cache

import (
	"fmt"
	"github.com/patrickmn/go-cache"
	"time"
)

type ttlCache struct {
	Cache *cache.Cache
}

var ttlCacheObj *ttlCache

func init() {
	fmt.Println("init in main.go ")
	ttlCacheObj = new(ttlCache)
	// 初始化cache 默认过期时间设置为5*time.Minute，扫描过期key的间隔时间10*time.Minute
	ttlCacheObj.Cache = cache.New(5*time.Minute, 10*time.Minute)
}

func GetTtlInstance() *ttlCache {
	return ttlCacheObj
}
