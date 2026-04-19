package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"
	"time"
)

type SourceEvent struct {
	Event     string    `json:"event"`
	CreatedAt string    `json:"created_at"`
	BrandID   int64     `json:"brand_id"`
	Body      EventBody `json:"body"`
}

type EventBody struct {
	BranchID     int64           `json:"branch_id"`
	OrderID      int64           `json:"order_id"`
	OrderType    int             `json:"order_type"`
	PickupTime   string          `json:"pickup_time"`
	IsSchedule   bool            `json:"is_schedule"`
	Discount     float64         `json:"discount"`
	Payment      SourcePayment   `json:"payment"`
	Products     []SourceProduct `json:"products"`
	Instructions *string         `json:"instructions"`
}

type SourcePayment struct {
	Type        string  `json:"type"`
	TotalAmount float64 `json:"total_amount"`
}

type SourceProduct struct {
	ExternalID string         `json:"external_id"`
	Name       string         `json:"name"`
	Quantity   int            `json:"quantity"`
	Price      float64        `json:"price"`
	Options    []SourceOption `json:"options"`
}

type SourceOption struct {
	ExternalID string  `json:"external_id"`
	Name       string  `json:"name"`
	Quantity   int     `json:"quantity"`
	Price      float64 `json:"price"`
}

type NinjaWebhook struct {
	PaymentMethod            string     `json:"payment_method"`
	AggregatorBranchID       string     `json:"aggregator_branch_id"`
	NinjaOrderID             string     `json:"ninja_order_id"`
	TotalDiscountAmountCents int64      `json:"total_discount_amount_cents"`
	TotalAmountCents         int64      `json:"total_amount_cents"`
	LineItems                []LineItem `json:"line_items"`
}

type LineItem struct {
	ItemID         string          `json:"item_id"`
	Quantity       int             `json:"quantity"`
	Name           string          `json:"name"`
	PriceCents     int64           `json:"price_cents"`
	ToppingOptions []ToppingOption `json:"topping_options,omitempty"`
}

type ToppingOption struct {
	ItemID     string `json:"item_id"`
	Name       string `json:"name"`
	PriceCents int64  `json:"price_cents"`
	Quantity   int    `json:"quantity"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook/order-created", orderCreatedWebhook)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("listening on :8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}

func orderCreatedWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	var input SourceEvent
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		http.Error(w, fmt.Sprintf("invalid payload: %v", err), http.StatusBadRequest)
		return
	}

	if input.Event != "order.created" {
		http.Error(w, "unsupported event", http.StatusBadRequest)
		return
	}

	payload := transformToNinja(input)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, fmt.Sprintf("failed encoding response: %v", err), http.StatusInternalServerError)
		return
	}
}

func transformToNinja(input SourceEvent) NinjaWebhook {
	result := NinjaWebhook{
		PaymentMethod:            strings.ToUpper(input.Body.Payment.Type),
		AggregatorBranchID:       fmt.Sprint(input.Body.BranchID),
		NinjaOrderID:             fmt.Sprint(input.Body.OrderID),
		TotalDiscountAmountCents: moneyToCents(input.Body.Discount),
		TotalAmountCents:         moneyToCents(input.Body.Payment.TotalAmount),
		LineItems:                make([]LineItem, 0, len(input.Body.Products)),
	}

	for _, p := range input.Body.Products {
		line := LineItem{
			ItemID:     p.ExternalID,
			Quantity:   p.Quantity,
			Name:       p.Name,
			PriceCents: moneyToCents(p.Price),
		}

		if len(p.Options) > 0 {
			line.ToppingOptions = make([]ToppingOption, 0, len(p.Options))
			for _, o := range p.Options {
				line.ToppingOptions = append(line.ToppingOptions, ToppingOption{
					ItemID:     o.ExternalID,
					Name:       o.Name,
					PriceCents: moneyToCents(o.Price),
					Quantity:   o.Quantity,
				})
			}
		}

		result.LineItems = append(result.LineItems, line)
	}

	return result
}

func moneyToCents(amount float64) int64 {
	return int64(math.Round(amount * 100))
}
