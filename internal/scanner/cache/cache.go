package cache

import (
	"encoding/json"
	"github.com/allegro/bigcache"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"log"
	"time"
)

// MemCache represents in-memory cache.
type MemCache struct {
	cache *bigcache.BigCache
}

// New creates a new in-memory bridge instance.
func New() *MemCache {
	c, err := bigcache.NewBigCache(bigcache.Config{
		Shards:             2048,
		LifeWindow:         5 * time.Minute,
		CleanWindow:        5 * time.Minute,
		MaxEntriesInWindow: 1500 * 10 * 60,
		MaxEntrySize:       2048,
		Verbose:            false,
		HardMaxCacheSize:   300,
		Logger:             log.Default(),
	})
	if err != nil {
		log.Fatalf("can not create cache; %s", err.Error())
	}
	return &MemCache{cache: c}
}

func (c *MemCache) Transaction(tx common.Hash, loader func(tx common.Hash)(*types.Transaction, error)) (trx *types.Transaction, err error) {
	key := "trx" + tx.String()

	data, err := c.cache.Get(key)
	if err == nil {
		if err := json.Unmarshal(data, &trx); err != nil {
			return nil, err
		}
		return trx, nil // HIT
	}

	trx, err = loader(tx) // load data from primary source

	data, err = json.Marshal(trx)
	if err != nil {
		log.Fatalf("can not encode trx into cache; %s", err)
	}
	err = c.cache.Set(key, data)
	if err != nil {
		log.Fatalf("can not store trx in cache; %s", err)
	}
	return trx, nil // MIS
}
