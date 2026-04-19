package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOrderCreatedWebhook_AllowsExtraFields(t *testing.T) {
	payload := []byte(`{
  "event": "order.created",
  "created_at": "2026-04-05 10:30:06",
  "brand_id": 17947040,
  "body": {
    "branch_id": 17948403,
    "order_id": 368511898,
    "order_type": 1,
    "pickup_time": "2026-04-05 10:30:06:00",
    "is_schedule": false,
    "discount": 0,
    "discount_details": {"type": 1, "value": 0, "total_amount": 0},
    "payment": {"type": "Online", "total_amount": 418.25},
    "order_status": 1,
    "products": [
      {
        "id": 8469611,
        "external_id": "product-5334463",
        "name": "BBQ Burger",
        "quantity": 1,
        "price": 18.25,
        "original_price": 18.25,
        "notes": ""
      },
      {
        "id": 8469614,
        "external_id": "product-5334428",
        "name": "Classic Beef Burger",
        "quantity": 1,
        "price": 25,
        "original_price": 25,
        "notes": "",
        "options": [
          {
            "id": 32991589,
            "external_id": "modifiervalue-4879273-5334439",
            "name": "Extra Cheese",
            "quantity": 1,
            "price": 25,
            "original_price": 25
          }
        ]
      }
    ],
    "instructions": null
  }
}`)

	req := httptest.NewRequest(http.MethodPost, "/webhook/order-created", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	orderCreatedWebhook(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	var got NinjaWebhook
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if got.PaymentMethod != "ONLINE" {
		t.Fatalf("unexpected payment method: %s", got.PaymentMethod)
	}
	if got.AggregatorBranchID != "17948403" || got.NinjaOrderID != "368511898" {
		t.Fatalf("unexpected ids: branch=%s order=%s", got.AggregatorBranchID, got.NinjaOrderID)
	}
	if got.TotalAmountCents != 41825 {
		t.Fatalf("unexpected total cents: %d", got.TotalAmountCents)
	}
	if len(got.LineItems) != 2 {
		t.Fatalf("expected 2 line items, got %d", len(got.LineItems))
	}
	if len(got.LineItems[1].ToppingOptions) != 1 {
		t.Fatalf("expected 1 topping option, got %d", len(got.LineItems[1].ToppingOptions))
	}
}
