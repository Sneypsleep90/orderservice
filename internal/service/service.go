package service

import (
	"fmt"
	"log"
	"myapp/internal/cache"
	"myapp/internal/model"
	"myapp/internal/repository"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/prometheus/client_golang/prometheus"
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
		ordersProcessErrorsTotal.Inc()
		return fmt.Errorf("order validation failed: %w", err)
	}

	if order.DateCreated.IsZero() {
		order.DateCreated = time.Now()
	}

	timer := prometheus.NewTimer(orderProcessDurationSeconds)
	defer timer.ObserveDuration()

	log.Printf("Saving order %s to database", order.OrderUID)
	if err := s.repo.CreateOrder(order); err != nil {
		ordersProcessErrorsTotal.Inc()
		return fmt.Errorf("failed to save order to database: %w", err)
	}

	log.Printf("Adding order %s to cache", order.OrderUID)
	s.cache.Set(order.OrderUID, order)

	log.Printf("Order %s processed successfully", order.OrderUID)
	ordersProcessedTotal.Inc()
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
	validatorInstance := validator.New(validator.WithRequiredStructEnabled())
	if err := validatorInstance.Struct(order); err != nil {
		return err
	}
	return nil
}
