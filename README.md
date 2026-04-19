# Go webhook transformer

This service exposes a webhook endpoint that accepts an `order.created` event payload and transforms it into the compact order format:

- `payment_method`
- `aggregator_branch_id`
- `ninja_order_id`
- `total_discount_amount_cents`
- `total_amount_cents`
- `line_items` with nested `topping_options`

## Run

```bash
go run .
```

Server starts on `:8080`.

## Endpoint

`POST /webhook/order-created`

- Input: source event payload (with `event = order.created` and `body.products`)
- Output: transformed JSON payload in cents

## Quick test

```bash
curl -X POST http://localhost:8080/webhook/order-created \
  -H 'Content-Type: application/json' \
  -d '{
    "event":"order.created",
    "created_at":"2026-04-05 10:30:06",
    "brand_id":17947040,
    "body":{
      "branch_id":17948403,
      "order_id":368511898,
      "order_type":1,
      "pickup_time":"2026-04-05 10:30:06:00",
      "is_schedule":false,
      "discount":0,
      "payment":{"type":"Online","total_amount":418.25},
      "products":[
        {"external_id":"product-5334463","name":"BBQ Burger","quantity":1,"price":18.25},
        {"external_id":"product-5334428","name":"Classic Beef Burger","quantity":1,"price":25,
          "options":[
            {"external_id":"modifiervalue-4879273-5334439","name":"Extra Cheese","quantity":1,"price":25}
          ]
        }
      ],
      "instructions":null
    }
  }'
```
