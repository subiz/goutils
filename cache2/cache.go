package cache

import (
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
)

type Cache struct {
	*sync.RWMutex
	lc      *lru.Cache
	loadf   func(key string) (interface{}, error)
	prefix  string
	expires map[string]int64
	exp     time.Duration
}

// clearExpire delete expire data from local cache (trigger 5mins)
func (c *Cache) clearExpire() {
	for {
		c.Lock()
		for k, v := range c.expires {
			if time.Now().UnixNano()-v < c.exp.Nanoseconds() {
				continue
			}
			c.lc.Remove(k)
			delete(c.expires, k)
		}
		c.Unlock()
		time.Sleep(5 * time.Minute)
	}
}

func New(prefix string, size int,
	loadf func(string) (interface{}, error), expire time.Duration) *Cache {
	c := &Cache{
		RWMutex: &sync.RWMutex{},
		prefix:  prefix,
		expires: make(map[string]int64),
		exp:     expire,
		loadf:   loadf,
	}
	var err error
	c.lc, err = lru.New(size)
	if err != nil {
		panic(err)
	}
	go c.clearExpire()
	return c
}

// GetGlobal loads data only from loadf, it skip local cache
func (c *Cache) GetGlobal(key string) (interface{}, error) {
	return c.loadf(key)
}

// Get loads data from local cache, if miss, call loadf to get the fresh data
func (c *Cache) Get(key string) (interface{}, error) {
	c.RLock()
	v, ok := c.lc.Get(key)
	if ok && time.Now().UnixNano()-c.expires[key] < c.exp.Nanoseconds() {
		c.RUnlock()
		return v, nil
	}
	c.RUnlock()

	// cache miss
	byts, err := c.loadf(key)
	if err != nil {
		return nil, err
	}

	// store back
	c.Lock()
	c.expires[key] = time.Now().UnixNano()
	c.lc.Add(key, byts)
	c.Unlock()

	return byts, nil
}

func (c *Cache) Set(key string, value interface{}) error {
	c.Lock()
	c.lc.Add(key, value)
	c.expires[key] = time.Now().UnixNano()
	c.Unlock()
	return nil
}

func (c *Cache) Remove(key string) error {
	c.Lock()
	c.lc.Remove(key)
	delete(c.expires, key)
	c.Unlock()
	return nil
}
