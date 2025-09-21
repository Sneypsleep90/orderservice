package handlers

import (
	"encoding/json"
	"log"
	"myapp/internal/model"
	"myapp/internal/service"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type Handler struct {
	service service.Service
}

func NewHandler(service service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	api := router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/orders", h.CreateOrder).Methods("POST")
	api.HandleFunc("/orders/{order_uid}", h.GetOrderByUID).Methods("GET")
	api.HandleFunc("/orders", h.GetAllOrders).Methods("GET")
	api.HandleFunc("/orders/{order_uid}", h.UpdateOrder).Methods("PUT")
	api.HandleFunc("/orders/{order_uid}", h.DeleteOrder).Methods("DELETE")
	api.HandleFunc("/cache/stats", h.GetCacheStats).Methods("GET")
	api.HandleFunc("/cache/warmup", h.WarmupCache).Methods("POST")

	router.HandleFunc("/order/{order_uid}", h.GetOrderByUID).Methods("GET")

	router.HandleFunc("/health", h.HealthCheck).Methods("GET")
}

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var order model.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.service.ProcessOrder(&order); err != nil {
		log.Printf("Error creating order: %v", err)
		http.Error(w, "Failed to create order", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(order); err != nil {
		log.Printf("Error encoding created order: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetOrderByUID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderUID := vars["order_uid"]

	if orderUID == "" {
		http.Error(w, "order_uid is required", http.StatusBadRequest)
		return
	}

	order, err := h.service.GetOrderByUID(orderUID)
	if err != nil {
		log.Printf("Error getting order %s: %v", orderUID, err)
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(order); err != nil {
		log.Printf("Error encoding order: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetAllOrders(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 100
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	orders, err := h.service.GetAllOrders()
	if err != nil {
		log.Printf("Error getting orders: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	total := len(orders)
	start := offset
	end := offset + limit

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginatedOrders := orders[start:end]

	response := map[string]interface{}{
		"orders": paginatedOrders,
		"pagination": map[string]interface{}{
			"total":  total,
			"limit":  limit,
			"offset": offset,
			"count":  len(paginatedOrders),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding orders: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderUID := vars["order_uid"]

	if orderUID == "" {
		http.Error(w, "order_uid is required", http.StatusBadRequest)
		return
	}

	var order model.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	order.OrderUID = orderUID

	if err := h.service.UpdateOrder(&order); err != nil {
		log.Printf("Error updating order %s: %v", orderUID, err)
		http.Error(w, "Failed to update order", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(order); err != nil {
		log.Printf("Error encoding updated order: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderUID := vars["order_uid"]

	if orderUID == "" {
		http.Error(w, "order_uid is required", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteOrder(orderUID); err != nil {
		log.Printf("Error deleting order %s: %v", orderUID, err)
		http.Error(w, "Failed to delete order", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetCacheStats(w http.ResponseWriter, r *http.Request) {
	stats := h.service.GetCacheStats()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		log.Printf("Error encoding cache stats: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) WarmupCache(w http.ResponseWriter, r *http.Request) {
	if err := h.service.WarmupCache(); err != nil {
		log.Printf("Error warming up cache: %v", err)
		http.Error(w, "Failed to warm up cache", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"message": "Cache warmup completed successfully",
		"time":    time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding warmup response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "order-service",
		"version":   "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding health check response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
