package main

import (
	"encoding/json"
	"fmt"
	"log"
	"myapp/internal/kafka"
	"myapp/internal/model"
	"os"
	"time"

	"github.com/brianvoe/gofakeit/v7"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run cmd/producer/main.go <order_uid>")
	}

	orderUID := os.Args[1]

	gofakeit.Seed(time.Now().UnixNano())
	order := &model.Order{
		OrderUID:          orderUID,
		TrackNumber:       gofakeit.Numerify("WB#########"),
		Entry:             "WBIL",
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        gofakeit.Username(),
		DeliveryService:   gofakeit.Company(),
		ShardKey:          fmt.Sprintf("%d", gofakeit.Number(1, 10)),
		SMID:              gofakeit.Number(1, 1000),
		DateCreated:       time.Now(),
		OOFShard:          fmt.Sprintf("%d", gofakeit.Number(1, 10)),
		Delivery: model.Delivery{
			Name:    gofakeit.Name(),
			Phone:   gofakeit.Phone(),
			Zip:     gofakeit.Zip(),
			City:    gofakeit.City(),
			Address: gofakeit.Street(),
			Region:  gofakeit.StateAbr(),
			Email:   gofakeit.Email(),
		},
		Payment: model.Payment{
			Transaction:  orderUID,
			RequestID:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       gofakeit.Number(100, 10000),
			PaymentDT:    time.Now().Unix(),
			Bank:         gofakeit.Company(),
			DeliveryCost: gofakeit.Number(100, 2000),
			GoodsTotal:   gofakeit.Number(100, 5000),
			CustomFee:    gofakeit.Number(0, 100),
		},
		Items: []model.Item{
			{
				ChrtID:      gofakeit.Number(1000000, 9999999),
				TrackNumber: gofakeit.Numerify("WB#########"),
				Price:       gofakeit.Number(10, 1000),
				RID:         gofakeit.UUID(),
				Name:        gofakeit.ProductName(),
				Sale:        gofakeit.Number(0, 80),
				Size:        fmt.Sprintf("%d", gofakeit.Number(0, 54)),
				TotalPrice:  gofakeit.Number(10, 3000),
				NMID:        gofakeit.Number(100000, 999999),
				Brand:       gofakeit.Company(),
				Status:      gofakeit.Number(100, 300),
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
