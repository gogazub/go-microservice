package tests

import (
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gogazub/myapp/internal/model"
)

var validate = validator.New()

// validOrderFixture создает корректный заказ.
// Большинство негативных тестов берут этот объект за основу
// и точечно ломают конкретное поле.
func validOrderFixture() model.Order {
	return model.Order{
		OrderUID:          "b563feb7b2b84b6test",
		TrackNumber:       "WBILMTESTTRACK",
		Entry:             "WBIL",
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "test",
		DeliveryService:   "meest",
		Shardkey:          "9",
		SmID:              99,
		DateCreated:       time.Date(2021, 11, 26, 6, 22, 19, 0, time.UTC),
		OofShard:          "1",
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
			Transaction:  "b563feb7b2b84b6test",
			RequestID:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1817,
			PaymentDt:    1637907727,
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
				Rid:         "ab4219087a764ae0btest",
				Name:        "Mascaras",
				Sale:        30,
				Size:        "0",
				TotalPrice:  317,
				NmID:        2389212,
				Brand:       "Vivienne Sabo",
				Status:      202,
			},
		},
	}
}

// mustBeInvalid убеждается, что валидация вернула ошибку именно типа validator.ValidationErrors.
// Если ошибки нет или тип другой - тест падает с понятным сообщением.
// Возвращает сам список нарушений для дальнейшей проверки.
func mustBeInvalid(t *testing.T, err error) validator.ValidationErrors {
	t.Helper()
	if err == nil {
		t.Fatal("expected validation error, got <nil>")
	}
	ve, ok := err.(validator.ValidationErrors)
	if !ok {
		t.Fatalf("expected validator.ValidationErrors, got %T: %v", err, err)
	}
	return ve
}

// mustHaveTagOnAnyField проверяет, что среди нарушений есть хотя бы одно с указанным тегом
// (например, "gte" или "required"), вне зависимости от поля.
func mustHaveTagOnAnyField(t *testing.T, ve validator.ValidationErrors, wantTag string) {
	t.Helper()
	for _, fe := range ve {
		if fe.Tag() == wantTag {
			return
		}
	}
	t.Fatalf("expected at least one error with tag=%q, got: %+v", wantTag, ve)
}

// mustHaveFieldTag проверяет, что нарушение относится к конкретному полю и имеет нужный тег.
// Здесь fe.Field() - это имя поля структуры (а не json ключ).
func mustHaveFieldTag(t *testing.T, ve validator.ValidationErrors, fieldName, wantTag string) {
	t.Helper()
	for _, fe := range ve {
		if fe.Field() == fieldName && fe.Tag() == wantTag {
			return
		}
	}
	t.Fatalf("expected error on field=%q with tag=%q, got: %+v", fieldName, wantTag, ve)
}

// TestOrder_OK — базовый позитивный сценарий.
// Проверяем, что валидный объект проходит валидацию без ошибок.
func TestOrder_OK(t *testing.T) {
	o := validOrderFixture()
	if err := validate.Struct(o); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

// TestOrder_MissingOrderUID - negative-case: отсутствует обязательное поле order_uid.
// Ожидаем ошибку по тегу "required" на поле OrderUID.
func TestOrder_MissingOrderUID(t *testing.T) {
	o := validOrderFixture()
	o.OrderUID = ""
	err := validate.Struct(o)
	ve := mustBeInvalid(t, err)
	mustHaveFieldTag(t, ve, "OrderUID", "required")
}

// TestOrder_MissingCustomerID - отсутствует обязательный customer_id.
func TestOrder_MissingCustomerID(t *testing.T) {
	o := validOrderFixture()
	o.CustomerID = ""
	err := validate.Struct(o)
	ve := mustBeInvalid(t, err)
	mustHaveFieldTag(t, ve, "CustomerID", "required")
}

// TestOrder_ZeroDateCreated - дата равна нулевому time.Time{}, что нарушает required.
func TestOrder_ZeroDateCreated(t *testing.T) {
	o := validOrderFixture()
	o.DateCreated = time.Time{}
	err := validate.Struct(o)
	ve := mustBeInvalid(t, err)
	mustHaveFieldTag(t, ve, "DateCreated", "required")
}

// TestOrder_MissingOofShard - пустой oof_shard (обязательное поле).
func TestOrder_MissingOofShard(t *testing.T) {
	o := validOrderFixture()
	o.OofShard = ""
	err := validate.Struct(o)
	ve := mustBeInvalid(t, err)
	mustHaveFieldTag(t, ve, "OofShard", "required")
}

// TestOrder_ItemsEmpty - items присутствует, но пустой срез.
// Нарушение тега "min=1" у поля Items.
func TestOrder_ItemsEmpty(t *testing.T) {
	o := validOrderFixture()
	o.Items = []model.Item{}
	err := validate.Struct(o)
	ve := mustBeInvalid(t, err)
	mustHaveFieldTag(t, ve, "Items", "min")
}

// TestOrder_ItemsNil - items = nil, что нарушает "required" для среза.
func TestOrder_ItemsNil(t *testing.T) {
	o := validOrderFixture()
	o.Items = nil
	err := validate.Struct(o)
	ve := mustBeInvalid(t, err)
	mustHaveFieldTag(t, ve, "Items", "required")
}

// TestOrder_SmIDNegative - sm_id < 0 нарушает "gte=0".
func TestOrder_SmIDNegative(t *testing.T) {
	o := validOrderFixture()
	o.SmID = -1
	err := validate.Struct(o)
	ve := mustBeInvalid(t, err)
	mustHaveFieldTag(t, ve, "SmID", "gte")
}

// TestDelivery_InvalidEmail - формат email некорректный.
// Ожидаем ошибку "email" на поле Email (вложенная структура Delivery).
func TestDelivery_InvalidEmail(t *testing.T) {
	o := validOrderFixture()
	o.Delivery.Email = "not-an-email"
	err := validate.Struct(o)
	ve := mustBeInvalid(t, err)
	mustHaveFieldTag(t, ve, "Email", "email")
}

// TestDelivery_MissingName - обязательное поле name пустое.
func TestDelivery_MissingName(t *testing.T) {
	o := validOrderFixture()
	o.Delivery.Name = ""
	err := validate.Struct(o)
	ve := mustBeInvalid(t, err)
	mustHaveFieldTag(t, ve, "Name", "required")
}

// TestPayment_AllRequiredPresent_OK - проверяем, что корректная оплата валидна сама по себе.
func TestPayment_AllRequiredPresent_OK(t *testing.T) {
	o := validOrderFixture()
	if err := validate.Struct(o.Payment); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

// TestPayment_MissingBank - обязательное поле bank пустое.
func TestPayment_MissingBank(t *testing.T) {
	o := validOrderFixture()
	o.Payment.Bank = ""
	err := validate.Struct(o)
	ve := mustBeInvalid(t, err)
	mustHaveFieldTag(t, ve, "Bank", "required")
}

// TestPayment_NegativeNumbers - проверяем все числовые поля с ограничением "gte=0".
// Для каждого из них выставляем отрицательные значения - ожидаем как минимум один "gte".
func TestPayment_NegativeNumbers(t *testing.T) {
	o := validOrderFixture()
	o.Payment.Amount = -1
	o.Payment.DeliveryCost = -1
	o.Payment.GoodsTotal = -1
	o.Payment.CustomFee = -1
	o.Payment.PaymentDt = -1
	err := validate.Struct(o)
	ve := mustBeInvalid(t, err)
	mustHaveTagOnAnyField(t, ve, "gte")
}

// TestItem_MinimalValid_OK - минимально валидный item.
// Проверяем, что нулевые значения, разрешенные правилами, проходят валидацию.
func TestItem_MinimalValid_OK(t *testing.T) {
	it := model.Item{
		ChrtID:      1,    // >=1
		TrackNumber: "TN", // required
		Price:       0,    // gte=0
		Sale:        0,    // 0..100
		TotalPrice:  0,    // gte=0
		NmID:        0,    // gte=0
		Status:      0,    // gte=0
	}
	if err := validate.Struct(it); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

// TestItem_ChrtIDZero - ChrtID=0. Для чисел "required" трактуется как "не ноль",
// поэтому ожидаем ошибку "required" (а не "gte").
func TestItem_ChrtIDZero(t *testing.T) {
	it := model.Item{
		ChrtID:      0, // нарушает required
		TrackNumber: "TN",
		Price:       10,
		Sale:        10,
		TotalPrice:  10,
		NmID:        1,
		Status:      1,
	}
	err := validate.Struct(it)
	ve := mustBeInvalid(t, err)
	mustHaveFieldTag(t, ve, "ChrtID", "required")
}

// TestItem_TrackNumberEmpty - пустой track_number нарушает "required".
func TestItem_TrackNumberEmpty(t *testing.T) {
	it := model.Item{
		ChrtID:      1,
		TrackNumber: "", // нарушает required
		Price:       10,
		Sale:        10,
		TotalPrice:  10,
		NmID:        1,
		Status:      1,
	}
	err := validate.Struct(it)
	ve := mustBeInvalid(t, err)
	mustHaveFieldTag(t, ve, "TrackNumber", "required")
}

// TestItem_SaleOutOfRange - sale=101 выходит за пределы [0..100], ожидаем ошибку "lte".
func TestItem_SaleOutOfRange(t *testing.T) {
	it := model.Item{
		ChrtID:      1,
		TrackNumber: "TN",
		Price:       10,
		Sale:        101, // > 100
		TotalPrice:  10,
		NmID:        1,
		Status:      1,
	}
	err := validate.Struct(it)
	ve := mustBeInvalid(t, err)
	mustHaveFieldTag(t, ve, "Sale", "lte")
}

// TestItems_Dive_FailsOnBadElement - проверяем работу "dive" для среза Items:
// добавляем во второй элемент заведомую ошибку (ChrtID=0) и ожидаем "required" на ChrtID.
func TestItems_Dive_FailsOnBadElement(t *testing.T) {
	o := validOrderFixture()
	o.Items = append(o.Items, model.Item{
		ChrtID:      0, // нарушает required
		TrackNumber: "TN2",
		Price:       1,
		Sale:        10,
		TotalPrice:  1,
		NmID:        1,
		Status:      1,
	})
	err := validate.Struct(o)
	ve := mustBeInvalid(t, err)
	mustHaveFieldTag(t, ve, "ChrtID", "required")
}
