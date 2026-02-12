# Company API Specs

API specifications for all 7 companies in the logistics ecosystem. Each company defines 2-3 endpoints with full request/response schemas.

> **Navigation:** [Back to Overview](README.md) | [Contracts](contracts.md) | [Perfect Integration](perfect.md) | [Mismatches](mismatches.md)

---

## 1.1 ShopStream (TypeScript) — E-commerce Storefront

#### POST /api/v1/orders

Create a new customer order.

```
Request:
  Content-Type: application/json
  Headers:
    Authorization: Bearer <api_key>
    Idempotency-Key: <uuid>
  Body:
    customer_id:    string, required, format: "cust_*"
    items:          array, required, min: 1
      sku:          string, required
      quantity:     integer, required, range: [1, 9999]
      unit_price:   integer, required, range: [1, 99999999]  # cents
    currency:       enum, required, values: ["usd","eur","gbp","cad","aud"]
    shipping_address:
      line1:        string, required
      line2:        string, optional
      city:         string, required
      state:        string, required
      postal_code:  string, required
      country:      string, required, format: ISO 3166-1 alpha-2
    callback_url:   string, optional, format: URL

Response 201:
    id:             string, format: "ord_*"
    status:         enum, values: ["pending","paid","processing","shipped","delivered","cancelled","refunded"]
    total_amount:   integer  # cents
    currency:       enum
    created_at:     string, format: ISO 8601 UTC
    items:          array

Response 400: { error: { code: string, message: string } }
Response 401: { error: { code: "unauthorized", message: string } }
Response 409: { error: { code: "duplicate", message: string, existing_order_id: string } }
Response 500: { error: { code: "internal", message: string } }
```

#### GET /api/v1/orders/:id

Retrieve order details.

```
Request:
  Headers:
    Authorization: Bearer <api_key>
  Params:
    id: string, format: "ord_*"

Response 200:
    id:             string, format: "ord_*"
    status:         enum, values: ["pending","paid","processing","shipped","delivered","cancelled","refunded"]
    total_amount:   integer  # cents
    currency:       enum
    tracking_number: string, nullable
    estimated_delivery: string, nullable, format: ISO 8601 UTC
    created_at:     string, format: ISO 8601 UTC
    updated_at:     string, format: ISO 8601 UTC

Response 404: { error: { code: "not_found", message: string } }
```

#### POST /api/v1/webhooks/delivery

Inbound webhook for delivery status updates.

```
Request:
  Method: POST
  Content-Type: application/json
  Headers:
    X-Webhook-Signature: string, required
  Body:
    order_id:       string, required, format: "ord_*"
    status:         enum, required, values: ["in_transit","out_for_delivery","delivered","failed"]
    timestamp:      string, required, format: ISO 8601 UTC
    details:        string, optional

Response 200: { received: true }
Response 400: { error: { code: "invalid_payload", message: string } }
Response 401: { error: { code: "invalid_signature", message: string } }
```

---

## 1.2 PayCore (Go) — Payment Processing

#### POST /v1/charges

Create a payment charge.

```
Request:
  Content-Type: application/json
  Headers:
    Authorization: Bearer <api_key>
    Idempotency-Key: <uuid>
  Body:
    amount:         integer, required, range: [50, 99999999]  # cents
    currency:       enum, required, values: ["usd","eur","gbp","cad","aud"]
    source_token:   string, required, format: "tok_*"
    description:    string, optional
    metadata:       object, optional

Response 200:
    id:             string, format: "pay_*"
    amount:         integer  # cents
    currency:       enum
    status:         enum, values: ["succeeded","pending","failed"]
    created_at:     string, format: ISO 8601 UTC
    metadata:       object

Response 400: { error: { type: "invalid_request", message: string } }
Response 402: { error: { type: "card_error", code: string, message: string } }
Response 500: { error: { type: "api_error", message: string } }
```

#### POST /v1/refunds

Create a refund.

```
Request:
  Content-Type: application/json
  Headers:
    Authorization: Bearer <api_key>
  Body:
    charge_id:      string, required, format: "pay_*"
    amount:         integer, optional, range: [50, 99999999]  # partial refund in cents
    reason:         enum, optional, values: ["duplicate","fraudulent","requested_by_customer"]

Response 200:
    id:             string, format: "ref_*"
    charge_id:      string, format: "pay_*"
    amount:         integer  # cents
    status:         enum, values: ["succeeded","pending","failed"]
    created_at:     string, format: ISO 8601 UTC

Response 400: { error: { type: "invalid_request", message: string } }
Response 404: { error: { type: "not_found", message: string } }
```

#### POST /v1/webhooks (outbound to ShopStream)

PayCore sends payment status webhooks.

```
Outbound webhook:
  Content-Type: application/json
  Headers:
    X-PayCore-Signature: string, HMAC-SHA256
    X-PayCore-Timestamp: string, Unix epoch seconds
  Body:
    event_id:       string, format: "evt_*"
    event_type:     enum, values: ["charge.succeeded","charge.failed","refund.succeeded"]
    data:
      id:           string  # pay_* or ref_*
      amount:       integer  # cents
      currency:     enum
      status:       enum
    created_at:     string, format: ISO 8601 UTC
```

---

## 1.3 VaultStore (Go) — Warehouse Management

#### POST /api/inventory/reserve

Reserve inventory for an order.

```
Request:
  Content-Type: application/json
  Headers:
    Authorization: Bearer <api_key>
  Body:
    order_id:       string, required
    items:          array, required
      sku:          string, required
      quantity:     integer, required, range: [1, 9999]
    warehouse_id:   string, optional, format: "wh_*"

Response 200:
    reservation_id: string, format: "res_*"
    status:         enum, values: ["reserved","partial","unavailable"]
    items:          array
      sku:          string
      quantity_reserved: integer
      quantity_available: integer
    expires_at:     string, format: ISO 8601 UTC

Response 400: { error: { code: "invalid_request", message: string } }
Response 409: { error: { code: "insufficient_stock", message: string, items: array } }
```

#### POST /api/shipments/prepare

Prepare a shipment (triggers label generation).

```
Request:
  Content-Type: application/json
  Headers:
    Authorization: Bearer <api_key>
  Body:
    reservation_id: string, required, format: "res_*"
    ship_to:
      name:         string, required
      line1:        string, required
      line2:        string, optional
      city:         string, required
      state:        string, required
      postal_code:  string, required
      country:      string, required, format: ISO 3166-1 alpha-2
    packages:       array, required
      weight:       integer, required  # grams
      dimensions:
        length:     integer, required  # millimeters
        width:      integer, required  # millimeters
        height:     integer, required  # millimeters

Response 200:
    shipment_id:    string, format: "shp_*"
    status:         enum, values: ["label_pending","label_created","picked_up","in_transit"]
    packages:       array
      weight:       integer  # grams
      dimensions:   { length: integer, width: integer, height: integer }  # mm

Response 400: { error: { code: "invalid_request", message: string } }
Response 404: { error: { code: "reservation_not_found", message: string } }
```

---

## 1.4 LabelForge (TypeScript) — Shipping Label Generation

#### POST /v2/labels

Generate a shipping label.

```
Request:
  Content-Type: application/json
  Headers:
    Authorization: Bearer <api_key>
  Body:
    from_address:
      name:         string, required
      street1:      string, required
      city:         string, required
      state:        string, required
      zip:          string, required
      country:      string, required, format: ISO 3166-1 alpha-2
    to_address:     (same structure as from_address)
    parcel:
      weight:       number, required  # grams (integer)
      length:       number, required  # centimeters
      width:        number, required  # centimeters
      height:       number, required  # centimeters
    service_level:  enum, required, values: ["ground","express","overnight","international"]

Response 200:
    label_id:       string, format: "lbl_*"
    tracking_number: string  # carrier-specific format
    carrier:        enum, values: ["ups","fedex","usps","dhl"]
    service_level:  enum
    rate:           integer  # cents
    label_url:      string, format: URL
    estimated_delivery: string, format: ISO 8601 UTC
    created_at:     string, format: ISO 8601 UTC

Response 400: { error: { code: string, message: string } }
Response 422: { error: { code: "unprocessable", message: string, details: array } }
```

#### GET /v2/tracking/:tracking_number

Get tracking information.

```
Request:
  Headers:
    Authorization: Bearer <api_key>
  Params:
    tracking_number: string, required

Response 200:
    tracking_number: string
    carrier:        enum, values: ["ups","fedex","usps","dhl"]
    status:         enum, values: ["pre_transit","in_transit","out_for_delivery","delivered","returned","failure"]
    estimated_delivery: string, nullable, format: ISO 8601 UTC
    events:         array
      timestamp:    string, format: ISO 8601 UTC
      location:     string
      description:  string
      status:       enum

Response 404: { error: { code: "not_found", message: string } }
```

---

## 1.5 SwiftHaul (Go) — Last-Mile Delivery

#### POST /api/pickups

Schedule a package pickup.

```
Request:
  Content-Type: application/json
  Headers:
    Authorization: Bearer <api_key>
  Body:
    tracking_number: string, required
    carrier:        enum, required, values: ["ups","fedex","usps","dhl"]
    pickup_address:
      street:       string, required
      city:         string, required
      state:        string, required
      zip:          string, required
      country:      string, required, format: ISO 3166-1 alpha-2
    package_details:
      weight_grams: integer, required
      length_cm:    integer, required
      width_cm:     integer, required
      height_cm:    integer, required
    pickup_date:    string, required, format: "YYYY-MM-DD"
    is_international: boolean, required

Response 200:
    pickup_id:      string, format: "pku_*"
    status:         enum, values: ["SCHEDULED","PICKED_UP","IN_TRANSIT","OUT_FOR_DELIVERY","DELIVERED","FAILED"]
    tracking_number: string
    estimated_delivery: string, format: ISO 8601 UTC

Response 400: { error: { code: string, message: string } }
Response 422: { error: { code: "invalid_address", message: string } }
```

#### POST /api/handoff/international

Hand off a package to international shipping partner.

```
Request:
  Content-Type: application/json
  Headers:
    Authorization: Bearer <api_key>
  Body:
    pickup_id:      string, required, format: "pku_*"
    tracking_number: string, required
    destination_country: string, required, format: ISO 3166-1 alpha-2
    package_details:
      weight_grams: integer, required
      length_cm:    integer, required
      width_cm:     integer, required
      height_cm:    integer, required
    declared_value: integer, required  # cents
    currency:       enum, required, values: ["usd","eur","gbp"]
    contents_description: string, required
    hs_code:        string, optional, format: "NNNN.NN.NNNN"

Response 200:
    handoff_id:     string, format: "hof_*"
    status:         enum, values: ["PENDING","ACCEPTED","IN_TRANSIT","DELIVERED"]
    ocean_tracking_id: string, nullable

Response 400: { error: { code: string, message: string } }
Response 404: { error: { code: "pickup_not_found", message: string } }
```

---

## 1.6 OceanBridge (Python) — International Shipping

#### POST /api/v1/bookings

Book a container / international shipment.

```
Request:
  Content-Type: application/json
  Headers:
    Authorization: Bearer <api_key>
  Body:
    origin_country:      string, required, format: ISO 3166-1 alpha-2
    destination_country:  string, required, format: ISO 3166-1 alpha-2
    packages:            array, required
      weight_kg:         number, required  # kilograms (float)
      length_cm:         number, required
      width_cm:          number, required
      height_cm:         number, required
    declared_value:      string, required  # decimal string, e.g. "29.99"
    currency:            enum, required, values: ["USD","EUR","GBP"]
    contents_description: string, required
    hs_code:             string, optional
    customs_info:
      sender_name:       string, required
      sender_tax_id:     string, optional
      recipient_name:    string, required

Response 201:
    booking_id:          string, format: "obk_*"
    status:              enum, values: ["pending","confirmed","in_transit","customs_hold","delivered","cancelled"]
    estimated_arrival:   string, format: ISO 8601 UTC
    tracking_url:        string, format: URL
    created_at:          string, format: ISO 8601 UTC

Response 400: { error: { code: string, message: string } }
Response 422: { error: { code: "customs_rejection", message: string, details: object } }
```

#### POST /api/v1/customs/notify

Send customs clearance notification (outbound to PingWave).

```
Outbound webhook:
  Content-Type: application/json
  Headers:
    X-OceanBridge-Signature: string, HMAC-SHA256
  Body:
    booking_id:          string, format: "obk_*"
    event_type:          enum, values: ["customs_cleared","customs_hold","customs_rejected"]
    notification_targets: array
      channel:           enum, values: ["email","sms","webhook"]
      destination:       string
    shipment_details:
      origin_country:    string, format: ISO 3166-1 alpha-2
      destination_country: string, format: ISO 3166-1 alpha-2
      estimated_arrival: string, format: ISO 8601 UTC
    created_at:          string, format: ISO 8601 UTC
```

---

## 1.7 PingWave (TypeScript) — Notifications

#### POST /v1/notifications/send

Send a notification via SMS, email, or webhook.

```
Request:
  Content-Type: application/json
  Headers:
    Authorization: Bearer <api_key>
  Body:
    channel:        enum, required, values: ["sms","email","webhook"]
    recipient:      string, required  # phone, email, or URL
    template_id:    string, optional, format: "tmpl_*"
    variables:      object, optional  # template variables
    message:        string, optional  # raw message if no template
    callback_url:   string, optional, format: URL  # delivery receipt callback
    metadata:       object, optional

Response 200:
    notification_id: string, format: "ntf_*"
    channel:        enum
    status:         enum, values: ["queued","sent","delivered","failed","bounced"]
    created_at:     string, format: ISO 8601 UTC

Response 400: { error: { code: string, message: string } }
Response 422: { error: { code: "invalid_recipient", message: string } }
```

#### POST /v1/callbacks/delivery (outbound to ShopStream)

PingWave sends delivery status callbacks to ShopStream.

```
Outbound webhook:
  Content-Type: application/json
  Headers:
    X-PingWave-Signature: string, HMAC-SHA256
    X-PingWave-Timestamp: string, Unix epoch seconds
  Body:
    event_type:     enum, values: ["delivery_confirmed","delivery_failed","notification_sent"]
    order_id:       string, format: "ord_*"
    status:         enum, values: ["in_transit","out_for_delivery","delivered","failed"]
    timestamp:      string, format: ISO 8601 UTC
    details:        string, optional
```
