package repository

import (
	"regexp"
	"testing"

	"myapp/internal/model"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestCreateOrder_InsertsAllParts(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	repo := &PostgresRepository{db: db}
	order := &model.Order{OrderUID: "u", TrackNumber: "t", Entry: "e", Locale: "en", CustomerID: "c", DeliveryService: "d"}
	order.Payment = model.Payment{Transaction: "txn"}
	order.Delivery = model.Delivery{Name: "n"}
	order.Items = []model.Item{{ChrtID: 1, TrackNumber: "t", Price: 1, RID: "r", Name: "n", TotalPrice: 1, NMID: 1, Brand: "b", Status: 1}}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO orders")).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO delivery")).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO payment")).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM items WHERE order_uid = $1")).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO items")).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := repo.CreateOrder(order); err != nil {
		t.Fatalf("CreateOrder error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
