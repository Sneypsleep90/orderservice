package service

import (
	"fmt"
	"log"
	"myapp/internal/cache"
	"myapp/internal/model"
	"myapp/internal/repository"
	"time"
)

type Service interface {
	ProcessOrder(order *model.Order) error
	GetOrderByUID(orderUID string) (*model.Order, error)
	GetAllOrders() ([]*model.Order, error)
	UpdateOrder(order *model.Order) error
	DeleteOrder(orderUID string) error
	GetCacheStats() cache.CacheStats
	WarmupCache() error
}

type OrderService struct {
	repo  repository.Repository
	cache cache.Cache
}

func NewOrderService(repo repository.Repository, cache cache.Cache) Service {
	return &OrderService{
		repo:  repo,
		cache: cache,
	}
}

func (s *OrderService) ProcessOrder(order *model.Order) error {
	log.Printf("Creating order: %s", order.OrderUID)

	if err := s.validateOrder(order); err != nil {
		return fmt.Errorf("order validation failed: %w", err)
	}

	if order.DateCreated.IsZero() {
		order.DateCreated = time.Now()
	}

	log.Printf("Saving order %s to database", order.OrderUID)
	if err := s.repo.CreateOrder(order); err != nil {
		return fmt.Errorf("failed to save order to database: %w", err)
	}

	log.Printf("Adding order %s to cache", order.OrderUID)
	s.cache.Set(order.OrderUID, order)

	log.Printf("Order %s processed successfully", order.OrderUID)
	return nil
}

func (s *OrderService) GetOrderByUID(orderUID string) (*model.Order, error) {
	if order, exists := s.cache.Get(orderUID); exists {
		log.Printf("Order %s found in cache", orderUID)
		return order, nil
	}

	order, err := s.repo.GetOrderByUID(orderUID)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}

	s.cache.Set(orderUID, order)
	log.Printf("Order %s retrieved from database and cached", orderUID)

	return order, nil
}

func (s *OrderService) GetAllOrders() ([]*model.Order, error) {
	orders, err := s.repo.GetAllOrders()
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	for _, order := range orders {
		s.cache.Set(order.OrderUID, order)
	}

	return orders, nil
}

func (s *OrderService) UpdateOrder(order *model.Order) error {

	if err := s.validateOrder(order); err != nil {
		return fmt.Errorf("order validation failed: %w", err)
	}

	if err := s.repo.UpdateOrder(order); err != nil {
		return fmt.Errorf("failed to update order in database: %w", err)
	}

	s.cache.Set(order.OrderUID, order)

	log.Printf("Order %s updated successfully", order.OrderUID)
	return nil
}

func (s *OrderService) DeleteOrder(orderUID string) error {
	if err := s.repo.DeleteOrder(orderUID); err != nil {
		return fmt.Errorf("failed to delete order from database: %w", err)
	}

	s.cache.Delete(orderUID)

	log.Printf("Order %s deleted successfully", orderUID)
	return nil
}

func (s *OrderService) GetCacheStats() cache.CacheStats {
	if statsCache, ok := s.cache.(*cache.StatsCache); ok {
		return statsCache.GetStats()
	}
	return cache.CacheStats{}
}

func (s *OrderService) WarmupCache() error {
	log.Println("Starting cache warmup...")

	orders, err := s.repo.GetAllOrders()
	if err != nil {
		return fmt.Errorf("failed to get orders for cache warmup: %w", err)
	}

	s.cache.Clear()

	for _, order := range orders {
		s.cache.Set(order.OrderUID, order)
	}

	log.Printf("Cache warmup completed. Loaded %d orders", len(orders))
	return nil
}

func (s *OrderService) validateOrder(order *model.Order) error {
	if order.OrderUID == "" {
		return fmt.Errorf("order_uid is required")
	}

	if order.TrackNumber == "" {
		return fmt.Errorf("track_number is required")
	}

	if order.CustomerID == "" {
		return fmt.Errorf("customer_id is required")
	}

	if order.Delivery.Name == "" {
		return fmt.Errorf("delivery name is required")
	}

	if order.Delivery.Phone == "" {
		return fmt.Errorf("delivery phone is required")
	}

	if order.Payment.Transaction == "" {
		return fmt.Errorf("payment transaction is required")
	}

	if order.Payment.Amount <= 0 {
		return fmt.Errorf("payment amount must be positive")
	}

	if len(order.Items) == 0 {
		return fmt.Errorf("order must have at least one item")
	}

	for i, item := range order.Items {
		if item.Name == "" {
			return fmt.Errorf("item %d name is required", i)
		}
		if item.Price <= 0 {
			return fmt.Errorf("item %d price must be positive", i)
		}
	}

	return nil
}
