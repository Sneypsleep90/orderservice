package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"myapp/internal/cache"
	"myapp/internal/model"

	"github.com/gorilla/mux"
)

type fakeService struct {
	order *model.Order
	err   error
}

func (f *fakeService) ProcessOrder(order *model.Order) error               { return nil }
func (f *fakeService) GetOrderByUID(orderUID string) (*model.Order, error) { return f.order, f.err }
func (f *fakeService) GetAllOrders() ([]*model.Order, error)               { return []*model.Order{f.order}, nil }
func (f *fakeService) UpdateOrder(order *model.Order) error                { return nil }
func (f *fakeService) DeleteOrder(orderUID string) error                   { return nil }
func (f *fakeService) GetCacheStats() cache.CacheStats                     { return cache.CacheStats{Size: 1} }
func (f *fakeService) WarmupCache() error                                  { return nil }

func TestGetOrderByUID(t *testing.T) {
	order := &model.Order{OrderUID: "uid1", TrackNumber: "trk", Entry: "en", Locale: "en", CustomerID: "c", DeliveryService: "d",
		Delivery: model.Delivery{Name: "n", Phone: "1", City: "c", Address: "a"},
		Payment:  model.Payment{Transaction: "txn", Currency: "USD", Provider: "p", Amount: 1, PaymentDT: 1, Bank: "b"},
		Items:    []model.Item{{ChrtID: 1, TrackNumber: "t", Price: 1, RID: "r", Name: "n", TotalPrice: 1, NMID: 1, Brand: "b", Status: 1}},
	}
	fs := &fakeService{order: order}
	h := NewHandler(fs)
	r := mux.NewRouter()
	h.RegisterRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/order/uid1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var got model.Order
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if got.OrderUID != "uid1" {
		t.Fatalf("unexpected OrderUID: %s", got.OrderUID)
	}
}
