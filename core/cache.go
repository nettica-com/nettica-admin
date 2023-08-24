package core

import (
	"time"

	"github.com/patrickmn/go-cache"
)

var StatusCache *cache.Cache = cache.New((60 * time.Minute), (10 * time.Minute))

func flushCache(id string) {
	StatusCache.Delete(id)
}

func getCache(id string) (interface{}, bool) {
	return StatusCache.Get(id)
}

func setCache(id string, status interface{}) {
	StatusCache.Set(id, status, cache.DefaultExpiration)
}
