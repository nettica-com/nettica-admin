package core

import (
	"time"

	"github.com/patrickmn/go-cache"
)

var StatusCache *cache.Cache = cache.New((1 * time.Minute), (1 * time.Minute))

func FlushCache(id string) {
	StatusCache.Delete(id)
}

func GetCache(id string) (interface{}, bool) {
	return StatusCache.Get(id)
}

func SetCache(id string, status interface{}) {
	StatusCache.Set(id, status, cache.DefaultExpiration)
}
