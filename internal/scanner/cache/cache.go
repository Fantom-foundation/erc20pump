package cache

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/patrickmn/go-cache"
	"strconv"
	"time"
)

// MemCache represents in-memory cache.
type MemCache struct {
	cache *cache.Cache
}

// New creates a new in-memory bridge instance.
func New() *MemCache {
	c := cache.New(5 * time.Minute, 10 * time.Minute)
	return &MemCache{cache: c}
}

func (c *MemCache) Transaction(tx common.Hash, loader func(tx common.Hash)(*types.Transaction, error)) (*types.Transaction, error) {
	key := "t" + tx.String()

	hit, found := c.cache.Get(key)
	if found {
		return hit.(*types.Transaction), nil
	}

	trx, err := loader(tx) // load data from primary source
	if err != nil {
		return trx, err
	}

	c.cache.Set(key, trx, cache.DefaultExpiration)
	return trx, nil
}

func (c *MemCache) Block(blockNumber uint64, loader func(blockNumber uint64)(*types.Block, error)) (block *types.Block, err error) {
	key := "b" + strconv.FormatUint(blockNumber, 16)

	hit, found := c.cache.Get(key)
	if found {
		return hit.(*types.Block), nil
	}

	block, err = loader(blockNumber) // load data from primary source
	if err != nil {
		return block, err
	}

	c.cache.Set(key, block, cache.DefaultExpiration)
	return block, nil
}
