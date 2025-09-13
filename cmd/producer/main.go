package main

import (
	"encoding/json"
	"fmt"
	"log"
	"myapp/internal/kafka"
	"myapp/internal/model"
	"os"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run cmd/producer/main.go <order_uid>")
	}

	orderUID := os.Args[1]

	order := &model.Order{
		OrderUID:          orderUID,
		TrackNumber:       "WBILMTESTTRACK",
		Entry:             "WBIL",
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "test",
		DeliveryService:   "meest",
		ShardKey:          "9",
		SMID:              99,
		DateCreated:       time.Now(),
		OOFShard:          "1",
		Delivery: model.Delivery{
			Name:    "Test Testov",
			Phone:   "+9720000000",
			Zip:     "2639809",
			City:    "Kiryat Mozkin",
			Address: "Ploshad Mira 15",
			Region:  "Kraiot",
			Email:   "test@gmail.com",
		},
		Payment: model.Payment{
			Transaction:  orderUID,
			RequestID:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1817,
			PaymentDT:    1637907727,
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   317,
			CustomFee:    0,
		},
		Items: []model.Item{
			{
				ChrtID:      9934930,
				TrackNumber: "WBILMTESTTRACK",
				Price:       453,
				RID:         "ab4219087a764ae0btest",
				Name:        "Mascaras",
				Sale:        30,
				Size:        "0",
				TotalPrice:  317,
				NMID:        2389212,
				Brand:       "Vivienne Sabo",
				Status:      202,
			},
		},
	}

	producer, err := kafka.NewProducer("localhost:9092", "orders")
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}
	defer producer.Close()

	if err := producer.SendOrder(order); err != nil {
		log.Fatalf("Failed to send order: %v", err)
	}

	jsonData, _ := json.MarshalIndent(order, "", "  ")
	fmt.Printf("Order sent successfully!\nOrder UID: %s\nJSON:\n%s\n", orderUID, string(jsonData))
}
