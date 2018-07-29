package cache

import (
	lru "github.com/hashicorp/golang-lru"
	"sync"
	"time"
)

type Cache struct {
	*sync.Mutex
	lc      *lru.Cache
	loadf   func(key string) ([]byte, error)
	prefix  string
	expires map[string]time.Time
	exp     time.Duration
}

// clearExpire delete expire data from local cache (trigger 5mins)
func (c *Cache) clearExpire() {
	for {
		c.Lock()
		for k, v := range c.expires {
			if time.Since(v) < c.exp {
				continue
			}
			c.lc.Remove(k)
			delete(c.expires, k)
		}
		c.Unlock()
		time.Sleep(5 * time.Minute)
	}
}

func New(prefix string, localsize int, redishosts []string, redispassword string,
	loadf func(string) ([]byte, error), expire time.Duration) *Cache {
	c := &Cache{
		Mutex:   &sync.Mutex{},
		prefix:  prefix,
		expires: make(map[string]time.Time),
		exp:     expire,
		loadf:   loadf,
	}
	var err error
	c.lc, err = lru.New(localsize)
	if err != nil {
		panic(err)
	}
	go c.clearExpire()
	return c
}

// GetGlobal loads data only from redis or loadf, it skip local cache
func (c *Cache) GetGlobal(key string) ([]byte, error) {
	return c.loadf(key)
}

// Get loads data from local cache, if miss, loads from redis, if also miss,
// call loadf to get the fresh data
func (c *Cache) Get(key string) ([]byte, error) {
	c.Lock()
	v, ok := c.lc.Get(key)
	if ok && time.Since(c.expires[key]) < c.exp {
		c.Unlock()
		return v.([]byte), nil
	}
	c.Unlock()

	// cache miss
	byts, err := c.loadf(key)
	if err != nil {
		return byts, err
	}

	c.Lock()
	// store back
	c.expires[key] = time.Now()
	c.lc.Add(key, byts)
	c.Unlock()

	return byts, nil
}

func (c *Cache) Set(key string, value []byte) error {
	c.Lock()
	c.lc.Add(key, value)
	c.expires[key] = time.Now()
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
