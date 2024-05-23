package dbpool

import (
	"errors"
	"sync"
	"time"

	. "github.com/NGRsoftlab/ngr-logging"

	"github.com/jmoiron/sqlx"
)

////////////////////////////////////////////////////////// Safe db pool map with string in key

type PoolItem struct {
	Expiration int64
	Duration   time.Duration
	Created    time.Time

	DB *sqlx.DB
}

type SafeDbMapCache struct {
	sync.RWMutex

	pool              map[string]PoolItem
	defaultExpiration time.Duration
	cleanupInterval   time.Duration
}

// New - initializing a new SafeDbMapCache cache
func New(defaultExpiration, cleanupInterval time.Duration) *SafeDbMapCache {
	items := make(map[string]PoolItem)

	// cache item
	cache := SafeDbMapCache{
		pool:              items,
		defaultExpiration: defaultExpiration,
		cleanupInterval:   cleanupInterval,
	}

	if cleanupInterval > 0 {
		cache.StartGC()
	}

	return &cache
}

// closePoolDB - close connector.DB wrapper
func (p *PoolItem) closePoolDB() {
	defer panicPoolRecover()
	err := p.DB.Close()
	if err != nil {
		Logger.Warningf("db connection close error: %s", err.Error())
	}
}

// nullDBRecover - recover
func panicPoolRecover() {
	if r := recover(); r != nil {
		Logger.Warning("Recovered in dbpool function: ", r)
	}
}

////////////////////////////////////////////////////////// Get-Set

// Get - getting *sqlx.DB value by key
func (c *SafeDbMapCache) Get(key string) (*sqlx.DB, bool) {
	// changed from RLock to Lock because of line 99 operation (updating creation time)
	c.Lock()
	defer c.Unlock()

	item, found := c.pool[key]

	// cache not found
	if !found {
		return nil, false
	}

	if item.Expiration > 0 {
		// cache expired
		if time.Now().UnixNano() > item.Expiration {
			return nil, false
		}
	}

	// TODO: set new timeout (?????? - think about it)
	var newExpiration int64
	if item.Duration > 0 {
		newExpiration = time.Now().Add(item.Duration).UnixNano()
	}

	c.pool[key] = PoolItem{
		DB:         item.DB,
		Expiration: newExpiration,
		Duration:   item.Duration,
		Created:    time.Now(),
	}

	return item.DB, true
}

// Set - setting *sqlx.DB value by key
func (c *SafeDbMapCache) Set(key string, value *sqlx.DB, duration time.Duration) {
	var expiration int64

	if duration == 0 {
		duration = c.defaultExpiration
	}

	if duration > 0 {
		expiration = time.Now().Add(duration).UnixNano()
	}

	c.Lock()

	defer c.Unlock()

	c.pool[key] = PoolItem{
		DB:         value,
		Expiration: expiration,
		Duration:   duration,
		Created:    time.Now(),
	}
}

////////////////////////////////////////////////////////// Items

// GetItems - returns item list.
func (c *SafeDbMapCache) GetItems() (items []string) {
	c.RLock()
	defer c.RUnlock()

	for k := range c.pool {
		items = append(items, k)
	}

	return
}

// ExpiredKeys - returns list of expired keys.
func (c *SafeDbMapCache) ExpiredKeys() (keys []string) {
	c.RLock()
	defer c.RUnlock()

	for k, i := range c.pool {
		if time.Now().UnixNano() > i.Expiration && i.Expiration > 0 {
			keys = append(keys, k)
		}
	}

	return
}

// clearItems - removes all the items with key in keys.
func (c *SafeDbMapCache) clearItems(keys []string) {
	c.Lock()
	defer c.Unlock()

	for _, k := range keys {
		connector, ok := c.pool[k]

		if ok {
			connector.closePoolDB()
		}

		delete(c.pool, k)
	}
}

////////////////////////////////////////////////////////// Cleaning

// StartGC - start Garbage Collection
func (c *SafeDbMapCache) StartGC() {
	go c.GC()
}

// GC - Garbage Collection cycle
func (c *SafeDbMapCache) GC() {
	for {
		<-time.After(c.cleanupInterval)

		if c.pool == nil {
			return
		}

		if keys := c.ExpiredKeys(); len(keys) != 0 {
			c.clearItems(keys)
		}
	}
}

// Delete - delete *sqlx.DB value by key. Return false if key not found
func (c *SafeDbMapCache) Delete(key string) error {
	c.Lock()
	defer c.Unlock()

	connector, found := c.pool[key]

	if !found {
		return errors.New("key not found")
	}

	connector.closePoolDB()

	delete(c.pool, key)

	return nil
}

// ClearAll - remove all items.
func (c *SafeDbMapCache) ClearAll() {
	c.Lock()
	defer c.Unlock()

	for k := range c.pool {
		connector, ok := c.pool[k]

		if ok {
			connector.closePoolDB()
		}

		delete(c.pool, k)
	}
}
