package cache

import (
	"encoding/binary"
	"fmt"
	"github.com/allegro/bigcache"
	"github.com/ethereum/go-ethereum/common"
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

// BlockTime returns cached time of the block by its number.
func (c *MemCache) BlockTime(bn uint64, load func(uint64) (uint64, error)) (uint64, error) {
	key := fmt.Sprintf("blk%x", bn)

	data, err := c.cache.Get(key)
	if err == nil {
		return binary.LittleEndian.Uint64(data), nil
	}

	// non cached - take the slow path
	v, err := load(bn)
	if err != nil {
		return 0, err
	}

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, v)
	if err := c.cache.Set(key, b); err != nil {
		log.Printf("unable to store; %s", err.Error())
	}

	return v, nil
}

// TrxRecipient provides cached recipient of a transaction by its hash.
func (c *MemCache) TrxRecipient(tx common.Hash, load func(common.Hash) (common.Address, error)) (common.Address, error) {
	// do we have the address in cache?
	data, err := c.cache.Get(tx.String())
	if err == nil {
		return common.BytesToAddress(data), nil
	}

	a, err := load(tx)
	if err != nil {
		log.Fatalf("recipient not available; %s", err.Error())
		return common.Address{}, err
	}

	if err := c.cache.Set(tx.String(), a.Bytes()); err != nil {
		log.Printf("can not cache; %s", err.Error())
	}

	return a, nil
}
