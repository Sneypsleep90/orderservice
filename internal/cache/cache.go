package cache

import (
	"myapp/internal/model"
	"sync"
	"time"
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
