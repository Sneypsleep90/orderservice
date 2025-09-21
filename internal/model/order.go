package model

import (
	"time"
)

type Order struct {
	OrderUID          string    `json:"order_uid" db:"order_uid" validate:"required,alphanumunicode,min=4,max=255"`
	TrackNumber       string    `json:"track_number" db:"track_number" validate:"required,min=3,max=255"`
	Entry             string    `json:"entry" db:"entry" validate:"required,min=2,max=255"`
	Delivery          Delivery  `json:"delivery" db:"delivery" validate:"required"`
	Payment           Payment   `json:"payment" db:"payment" validate:"required"`
	Items             []Item    `json:"items" db:"items" validate:"required,min=1,dive"`
	Locale            string    `json:"locale" db:"locale" validate:"required,len=2|len=5"`
	InternalSignature string    `json:"internal_signature" db:"internal_signature"`
	CustomerID        string    `json:"customer_id" db:"customer_id" validate:"required,min=1,max=255"`
	DeliveryService   string    `json:"delivery_service" db:"delivery_service" validate:"required,min=2,max=255"`
	ShardKey          string    `json:"shardkey" db:"shardkey" validate:"omitempty,max=10"`
	SMID              int       `json:"sm_id" db:"sm_id" validate:"gte=0"`
	DateCreated       time.Time `json:"date_created" db:"date_created"`
	OOFShard          string    `json:"oof_shard" db:"oof_shard" validate:"omitempty,max=10"`
}

type Delivery struct {
	Name    string `json:"name" db:"name" validate:"required,min=2,max=255"`
	Phone   string `json:"phone" db:"phone" validate:"required,min=5,max=50"`
	Zip     string `json:"zip" db:"zip" validate:"omitempty,min=3,max=20"`
	City    string `json:"city" db:"city" validate:"required,min=2,max=255"`
	Address string `json:"address" db:"address" validate:"required,min=3"`
	Region  string `json:"region" db:"region" validate:"omitempty,max=255"`
	Email   string `json:"email" db:"email" validate:"omitempty,email"`
}

type Payment struct {
	Transaction  string `json:"transaction" db:"transaction" validate:"required,min=3,max=255"`
	RequestID    string `json:"request_id" db:"request_id" validate:"omitempty,max=255"`
	Currency     string `json:"currency" db:"currency" validate:"required,uppercase,len=3"`
	Provider     string `json:"provider" db:"provider" validate:"required,min=2,max=100"`
	Amount       int    `json:"amount" db:"amount" validate:"required,gt=0"`
	PaymentDT    int64  `json:"payment_dt" db:"payment_dt" validate:"required,gt=0"`
	Bank         string `json:"bank" db:"bank" validate:"required,min=2,max=100"`
	DeliveryCost int    `json:"delivery_cost" db:"delivery_cost" validate:"gte=0"`
	GoodsTotal   int    `json:"goods_total" db:"goods_total" validate:"gte=0"`
	CustomFee    int    `json:"custom_fee" db:"custom_fee" validate:"gte=0"`
}

type Item struct {
	ChrtID      int    `json:"chrt_id" db:"chrt_id" validate:"required,gt=0"`
	TrackNumber string `json:"track_number" db:"track_number" validate:"required,min=3,max=255"`
	Price       int    `json:"price" db:"price" validate:"required,gt=0"`
	RID         string `json:"rid" db:"rid" validate:"required,min=3,max=255"`
	Name        string `json:"name" db:"name" validate:"required,min=1"`
	Sale        int    `json:"sale" db:"sale" validate:"gte=0,lte=100"`
	Size        string `json:"size" db:"size" validate:"omitempty,max=10"`
	TotalPrice  int    `json:"total_price" db:"total_price" validate:"required,gt=0"`
	NMID        int    `json:"nm_id" db:"nm_id" validate:"required,gt=0"`
	Brand       string `json:"brand" db:"brand" validate:"required,min=1"`
	Status      int    `json:"status" db:"status" validate:"required"`
}
