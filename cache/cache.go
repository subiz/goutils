package cache

import (
	"git.subiz.net/goredis"
	lru "github.com/hashicorp/golang-lru"
	"github.com/prometheus/client_golang/prometheus"
	"sync"
	"time"
)

var LocalCacheHitCounter = prometheus.NewCounter(
	prometheus.CounterOpts{
		Subsystem: "global",
		Name:      "local_cache_hit",
		Help:      "Local cache hits",
	},
)

var GlobalCacheHitCounter = prometheus.NewCounter(
	prometheus.CounterOpts{
		Subsystem: "global",
		Name:      "global_cache_hit",
		Help:      "Global cache hits",
	},
)

var TotalCacheHitCounter = prometheus.NewCounter(
	prometheus.CounterOpts{
		Subsystem: "global",
		Name:      "total_cache_hit",
		Help:      "Total cache hits",
	},
)

type Cache struct {
	*sync.Mutex
	rclient *goredis.Client
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
	c.rclient, err = goredis.New(redishosts, redispassword)
	c.lc, err = lru.New(localsize)
	if err != nil {
		panic(err)
	}
	go c.clearExpire()
	return c
}

// GetGlobal loads data only from redis or loadf, it skip local cache
func (c *Cache) GetGlobal(key string) ([]byte, error) {
	TotalCacheHitCounter.Inc()

	byts, err := c.rclient.Get(c.prefix+key, c.prefix+key)
	c.Lock()
	if err == nil {
		c.expires[key] = time.Now()
		c.Unlock()
		GlobalCacheHitCounter.Inc()
		return byts, nil
	}
	c.Unlock()
	// redis cache miss
	byts, err = c.loadf(key)
	if err != nil {
		return byts, err
	}
	// store back
	c.Lock()
	c.lc.Add(key, byts)
	c.expires[key] = time.Now()
	c.Unlock()
	c.rclient.Set(c.prefix+key, c.prefix+key, byts, 10*c.exp) // ignore err
	return byts, err
}

// Get loads data from local cache, if miss, loads from redis, if also miss,
// call loadf to get the fresh data
func (c *Cache) Get(key string) ([]byte, error) {
	TotalCacheHitCounter.Inc()
	c.Lock()
	v, ok := c.lc.Get(key)
	if ok && time.Since(c.expires[key]) < c.exp {
		c.Unlock()
		LocalCacheHitCounter.Inc()
		return v.([]byte), nil
	}
	c.Unlock()
	// local cache miss
	byts, err := c.rclient.Get(c.prefix+key, c.prefix+key)
	if err == nil {
		GlobalCacheHitCounter.Inc()
		c.Lock()
		c.expires[key] = time.Now()
		c.lc.Add(key, byts) // store back to client
		c.Unlock()
		return byts, nil
	}
	// redis cache miss
	byts, err = c.loadf(key)
	if err != nil {
		return byts, err
	}

	c.Lock()
	// store back
	c.expires[key] = time.Now()
	c.lc.Add(key, byts)
	c.Unlock()

	c.rclient.Set(c.prefix+key, c.prefix+key, byts, 10*c.exp) // ignore err
	return byts, nil
}

func (c *Cache) Set(key string, value []byte) error {
	c.Lock()
	c.lc.Add(key, value)
	c.expires[key] = time.Now()
	c.Unlock()
	return c.rclient.Set(c.prefix+key, c.prefix+key, value, 10*c.exp)
}

func (c *Cache) Remove(key string) error {
	c.Lock()
	c.lc.Remove(key)
	delete(c.expires, key)
	c.Unlock()
	return c.rclient.Expire(c.prefix+key, c.prefix+key, 0)
}
