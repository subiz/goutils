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
}

func New(prefix string, localsize int, redishosts []string, redispassword string,
	loadf func(string) ([]byte, error),
	removef func(string) error) *Cache {
	c := &Cache{
		Mutex:   &sync.Mutex{},
		loadf:   loadf,
		rclient: &goredis.Client{},
	}
	c.rclient.Connect(redishosts, redispassword)
	var err error
	c.lc, err = lru.New(localsize)
	if err != nil {
		panic(err)
	}
	return c
}

func (c *Cache) Get(key string) ([]byte, error) {
	c.Lock()
	defer c.Unlock()
	v, ok := c.lc.Get(key)
	if ok {
		return v.([]byte), nil
	}

	// local cache miss
	byts, err := c.rclient.Get(c.prefix + key, c.prefix + key)
	if err == nil {
		c.lc.Add(key, byts) // store back to client
		return byts, nil
	}

	// redis cache miss
	byts, err = c.loadf(key)
	if err != nil {
		return byts, err
	}

	// store back
	c.lc.Add(key, byts)
	c.rclient.Set(c.prefix + key, c.prefix + key, byts, 10*time.Minute) // ignore err
	return byts, nil
}

func (c *Cache) Set(key string, value []byte) error {
	c.Lock()
	defer c.Unlock()
	c.lc.Add(key, value)
	return c.rclient.Set(c.prefix + key, c.prefix + key, value, 10*time.Minute)
}

func (c *Cache) Remove(key string) error {
	c.Lock()
	defer c.Unlock()
	c.lc.Remove(key)
	return c.rclient.Expire(c.prefix + key, c.prefix + key, 0)
}
