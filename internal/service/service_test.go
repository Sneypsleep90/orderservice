package service

import (
	"testing"
	"time"

	"myapp/internal/cache"
	"myapp/internal/model"
)

type fakeRepo struct {
	createCalled bool
}

func (f *fakeRepo) CreateOrder(order *model.Order) error { f.createCalled = true; return nil }
func (f *fakeRepo) GetOrderByUID(orderUID string) (*model.Order, error) {
	return &model.Order{OrderUID: orderUID, TrackNumber: "t", Entry: "e", Locale: "en", CustomerID: "c", DeliveryService: "d", DateCreated: time.Now(), Delivery: model.Delivery{Name: "n", Phone: "1", City: "c", Address: "a"}, Payment: model.Payment{Transaction: "t", Currency: "USD", Provider: "p", Amount: 1, PaymentDT: time.Now().Unix(), Bank: "b"}, Items: []model.Item{{ChrtID: 1, TrackNumber: "t", Price: 1, RID: "r", Name: "n", TotalPrice: 1, NMID: 1, Brand: "b", Status: 1}}}, nil
}
func (f *fakeRepo) GetAllOrders() ([]*model.Order, error) { return nil, nil }
func (f *fakeRepo) UpdateOrder(order *model.Order) error  { return nil }
func (f *fakeRepo) DeleteOrder(orderUID string) error     { return nil }

func TestProcessOrder_Valid(t *testing.T) {
	repo := &fakeRepo{}
	c := cache.NewInMemoryCache()
	s := NewOrderService(repo, c)

	order := &model.Order{
		OrderUID:        "uid1",
		TrackNumber:     "trk",
		Entry:           "en",
		Locale:          "en",
		CustomerID:      "cust",
		DeliveryService: "svc",
		DateCreated:     time.Now(),
		Delivery:        model.Delivery{Name: "name", Phone: "12345", City: "city", Address: "addr"},
		Payment:         model.Payment{Transaction: "txn", Currency: "USD", Provider: "prov", Amount: 10, PaymentDT: time.Now().Unix(), Bank: "bank"},
		Items:           []model.Item{{ChrtID: 1, TrackNumber: "trk", Price: 10, RID: "rid", Name: "nm", TotalPrice: 10, NMID: 1, Brand: "br", Status: 1}},
	}
	if err := s.ProcessOrder(order); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
