package dbpool

import (
	logging "github.com/NGRsoftlab/ngr-logging"

	"errors"
	"github.com/jmoiron/sqlx"
	"sync"
	"time"
)

///////Safe db pool map with string in key///////////

type PoolItem struct {
	Expiration int64
	Duration   time.Duration
	Created    time.Time

	Db *sqlx.DB
}

type SafeDbMapCache struct {
	sync.RWMutex

	pool              map[string]PoolItem
	defaultExpiration time.Duration
	cleanupInterval   time.Duration
}

// New. Initializing a new memory cache
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

// Set setting a cache by key
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
		Db:         value,
		Expiration: expiration,
		Duration:   duration,
		Created:    time.Now(),
	}
}

// Get getting a cache by key
func (c *SafeDbMapCache) Get(key string) (*sqlx.DB, bool) {
	c.RLock()
	defer c.RUnlock()

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

	////TODO: set new timeout (?????? - think about it)
	var newExpiration int64
	if item.Duration > 0 {
		newExpiration = time.Now().Add(item.Duration).UnixNano()
	}

	c.pool[key] = PoolItem{
		Db:         item.Db,
		Expiration: newExpiration,
		Duration:   item.Duration,
		Created:    time.Now(),
	}

	return item.Db, true
}

// Delete cache by key
// Return false if key not found
func (c *SafeDbMapCache) Delete(key string) error {
	c.Lock()
	defer c.Unlock()

	connector, found := c.pool[key]

	if !found {
		return errors.New("key not found")
	}

	err := connector.Db.Close()
	if err != nil {
		logging.Logger.Warning("db connection close error: ", err)
	}

	delete(c.pool, key)

	return nil
}

// StartGC start Garbage Collection
func (c *SafeDbMapCache) StartGC() {
	go c.GC()
}

// GC Garbage Collection
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

// expiredKeys returns key list which are expired.
func (c *SafeDbMapCache) GetItems() (items []string) {
	c.RLock()
	defer c.RUnlock()

	for k, _ := range c.pool {
		items = append(items, k)
	}

	return
}

// expiredKeys returns key list which are expired.
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

// clearItems removes all the items which key in keys.
func (c *SafeDbMapCache) clearItems(keys []string) {
	c.Lock()
	defer c.Unlock()

	for _, k := range keys {
		connector, ok := c.pool[k]

		if ok {
			err := connector.Db.Close()
			if err != nil {
				logging.Logger.Warning("db connection close error: ", err)
			}
		}

		delete(c.pool, k)
	}
}

// ClearAll removes all the items which key in keys.
func (c *SafeDbMapCache) ClearAll() {
	c.Lock()
	defer c.Unlock()

	for k := range c.pool {
		connector, ok := c.pool[k]

		if ok {
			err := connector.Db.Close()
			if err != nil {
				logging.Logger.Warning("db connection close error: ", err)
			}
		}

		delete(c.pool, k)
	}
}
