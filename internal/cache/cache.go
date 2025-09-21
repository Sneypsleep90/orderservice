package cache

import (
	"myapp/internal/model"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
)

type Cache interface {
	Set(orderUID string, order *model.Order)
	Get(orderUID string) (*model.Order, bool)
	Delete(orderUID string)
	GetAll() map[string]*model.Order
	Clear()
	Size() int
}

type InMemoryCache struct {
	mu    sync.RWMutex
	items map[string]*model.Order
}

func NewInMemoryCache() Cache {
	return &InMemoryCache{
		items: make(map[string]*model.Order),
	}
}

func (c *InMemoryCache) Set(orderUID string, order *model.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[orderUID] = order
}

func (c *InMemoryCache) Get(orderUID string) (*model.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	order, exists := c.items[orderUID]
	return order, exists
}

func (c *InMemoryCache) Delete(orderUID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, orderUID)
}

func (c *InMemoryCache) GetAll() map[string]*model.Order {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]*model.Order)
	for k, v := range c.items {
		result[k] = v
	}
	return result
}

func (c *InMemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*model.Order)
}

func (c *InMemoryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

type LRUCache struct {
	cache *lru.Cache[string, *model.Order]
}

func NewLRUCache(capacity int) (Cache, error) {
	c, err := lru.New[string, *model.Order](capacity)
	if err != nil {
		return nil, err
	}
	return &LRUCache{cache: c}, nil
}

func (c *LRUCache) Set(orderUID string, order *model.Order) {
	c.cache.Add(orderUID, order)
}

func (c *LRUCache) Get(orderUID string) (*model.Order, bool) {
	return c.cache.Get(orderUID)
}

func (c *LRUCache) Delete(orderUID string) {
	c.cache.Remove(orderUID)
}

func (c *LRUCache) GetAll() map[string]*model.Order {
	result := make(map[string]*model.Order)
	for _, key := range c.cache.Keys() {
		if val, ok := c.cache.Peek(key); ok {
			result[key] = val
		}
	}
	return result
}

func (c *LRUCache) Clear() {
	c.cache.Purge()
}

func (c *LRUCache) Size() int {
	return c.cache.Len()
}

type CacheStats struct {
	Size      int           `json:"size"`
	HitRate   float64       `json:"hit_rate"`
	MissRate  float64       `json:"miss_rate"`
	TotalHits int64         `json:"total_hits"`
	TotalMiss int64         `json:"total_miss"`
	Uptime    time.Duration `json:"uptime"`
}

type StatsCache struct {
	Cache
	stats     CacheStats
	startTime time.Time
	mu        sync.RWMutex
}

func NewStatsCache(cache Cache) *StatsCache {
	return &StatsCache{
		Cache:     cache,
		startTime: time.Now(),
	}
}

func (sc *StatsCache) Get(orderUID string) (*model.Order, bool) {
	order, exists := sc.Cache.Get(orderUID)

	sc.mu.Lock()
	if exists {
		sc.stats.TotalHits++
	} else {
		sc.stats.TotalMiss++
	}
	sc.updateRates()
	sc.mu.Unlock()

	return order, exists
}

func (sc *StatsCache) GetStats() CacheStats {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	stats := sc.stats
	stats.Size = sc.Cache.Size()
	stats.Uptime = time.Since(sc.startTime)

	return stats
}

func (sc *StatsCache) updateRates() {
	total := sc.stats.TotalHits + sc.stats.TotalMiss
	if total > 0 {
		sc.stats.HitRate = float64(sc.stats.TotalHits) / float64(total)
		sc.stats.MissRate = float64(sc.stats.TotalMiss) / float64(total)
	}
}
