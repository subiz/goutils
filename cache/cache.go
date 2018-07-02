package cache

import (
	"bitbucket.org/subiz/goredis"
	lru "github.com/hashicorp/golang-lru"
	"sync"
	"time"
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
		rclient: &goredis.Client{},
	}
	c.rclient.Connect(redishosts, redispassword)
	var err error
	c.lc, err = lru.New(localsize)
	if err != nil {
		panic(err)
	}
	go c.clearExpire()
	return c
}

func (c *Cache) Get(key string) ([]byte, error) {
	c.Lock()
	defer c.Unlock()

	v, ok := c.lc.Get(key)
	if ok && time.Since(c.expires[key]) < c.exp {
		return v.([]byte), nil
	}

	// local cache miss
	byts, err := c.rclient.Get(c.prefix+key, c.prefix+key)
	if err == nil {
		c.expires[key] = time.Now()
		c.lc.Add(key, byts) // store back to client
		return byts, nil
	}

	// redis cache miss
	byts, err = c.loadf(key)
	if err != nil {
		return byts, err
	}

	// store back
	c.expires[key] = time.Now()
	c.lc.Add(key, byts)
	c.rclient.Set(c.prefix+key, c.prefix+key, byts, 10*c.exp) // ignore err
	return byts, nil
}

func (c *Cache) Set(key string, value []byte) error {
	c.Lock()
	defer c.Unlock()
	c.lc.Add(key, value)
	c.expires[key] = time.Now()
	return c.rclient.Set(c.prefix+key, c.prefix+key, value, 10*c.exp)
}

func (c *Cache) remove(key string) error {
	c.lc.Remove(key)
	delete(c.expires, key)
	return c.rclient.Expire(c.prefix+key, c.prefix+key, 0)
}

func (c *Cache) Remove(key string) error {
	c.Lock()
	defer c.Unlock()
	return c.remove(key)
}
