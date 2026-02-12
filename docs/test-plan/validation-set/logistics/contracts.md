# Inter-Service Contracts & Stricture Manifest

8 contracts connecting the 7 companies, plus the complete Stricture manifest.

> **Navigation:** [Back to Overview](README.md) | [Companies](companies.md) | [Perfect Integration](perfect.md) | [Mismatches](mismatches.md)

---

## Inter-Service Contracts

### C1: ShopStream -> PayCore (Create Charge)

When a customer places an order, ShopStream calls PayCore to charge their payment method.

```
Producer: PayCore (POST /v1/charges)
Consumer: ShopStream

Data flow:
  ShopStream sends:
    amount:       order.total_amount (integer, cents)
    currency:     order.currency (lowercase enum)
    source_token: customer.payment_token (tok_*)
    description:  "Order {order.id}"
    metadata:     { order_id: order.id, customer_id: order.customer_id }

  PayCore returns:
    id:           "pay_*" -> stored as order.payment_id
    amount:       integer cents -> verified against order.total_amount
    status:       "succeeded" | "pending" | "failed"
    created_at:   ISO 8601 UTC

Invariants:
  - Amount is always integer cents (never float dollars)
  - Currency is always lowercase 3-letter ISO 4217
  - ShopStream must handle all 3 status values
  - Idempotency-Key prevents duplicate charges on retry
```

### C2: ShopStream -> VaultStore (Reserve Inventory)

After payment succeeds, ShopStream reserves inventory.

```
Producer: VaultStore (POST /api/inventory/reserve)
Consumer: ShopStream

Data flow:
  ShopStream sends:
    order_id:     order.id (ord_*)
    items:        order.items mapped to [{ sku, quantity }]

  VaultStore returns:
    reservation_id: "res_*" -> stored as order.reservation_id
    status:       "reserved" | "partial" | "unavailable"
    items:        per-item reservation status
    expires_at:   ISO 8601 UTC -> ShopStream sets timer

Invariants:
  - If status is "partial" or "unavailable", ShopStream must refund payment
  - Reservation has TTL (expires_at); must be fulfilled before expiry
  - SKU identifiers must match exactly between systems
```

### C3: VaultStore -> LabelForge (Request Shipping Label)

VaultStore prepares the shipment and requests a label from LabelForge.

```
Producer: LabelForge (POST /v2/labels)
Consumer: VaultStore

Data flow:
  VaultStore sends:
    from_address:  warehouse address
    to_address:    order shipping address
    parcel:
      weight:      package.weight (grams, integer)
      length:      package.dimensions.length (converted from mm to cm)
      width:       package.dimensions.width (converted from mm to cm)
      height:      package.dimensions.height (converted from mm to cm)
    service_level: determined by shipping speed selection

  LabelForge returns:
    label_id:      "lbl_*"
    tracking_number: carrier-specific format
    carrier:       "ups" | "fedex" | "usps" | "dhl"
    rate:          integer cents
    estimated_delivery: ISO 8601 UTC

Invariants:
  - Weight: VaultStore stores in grams, LabelForge accepts grams
  - Dimensions: VaultStore stores in mm, must convert to cm for LabelForge
  - Country codes: both use ISO 3166-1 alpha-2
```

### C4: LabelForge -> SwiftHaul (Schedule Pickup)

After label creation, a pickup is scheduled with SwiftHaul.

```
Producer: SwiftHaul (POST /api/pickups)
Consumer: LabelForge (or VaultStore orchestrating)

Data flow:
  Sender sends:
    tracking_number: from LabelForge response
    carrier:        from LabelForge response
    pickup_address: warehouse address
    package_details:
      weight_grams: package weight in grams
      length_cm:    dimensions in cm
      width_cm:     dimensions in cm
      height_cm:    dimensions in cm
    pickup_date:    "YYYY-MM-DD"
    is_international: boolean

  SwiftHaul returns:
    pickup_id:      "pku_*"
    status:         "SCHEDULED" (SCREAMING_SNAKE_CASE)
    estimated_delivery: ISO 8601 UTC

Invariants:
  - tracking_number format varies by carrier (UPS: 1Z*, FedEx: numeric)
  - SwiftHaul uses SCREAMING_SNAKE_CASE for status enums
  - Weight in grams, dimensions in cm
```

### C5: SwiftHaul -> OceanBridge (International Handoff)

For international shipments, SwiftHaul hands off to OceanBridge.

```
Producer: OceanBridge (POST /api/v1/bookings)
Consumer: SwiftHaul

Data flow:
  SwiftHaul sends (via /api/handoff/international -> transforms to OceanBridge format):
    origin_country:      "US" (ISO alpha-2)
    destination_country: from order (ISO alpha-2)
    packages:
      weight_kg:         package.weight_grams / 1000.0 (convert grams to kg)
      length_cm:         package.length_cm
      width_cm:          package.width_cm
      height_cm:         package.height_cm
    declared_value:      fmt.Sprintf("%.2f", float64(cents) / 100.0)  # string decimal
    currency:            strings.ToUpper(currency)  # OceanBridge uses uppercase
    contents_description: from handoff request

  OceanBridge returns:
    booking_id:     "obk_*"
    status:         "pending" | "confirmed" | ... (lowercase)
    estimated_arrival: ISO 8601 UTC

Invariants:
  - Weight conversion: grams (SwiftHaul) -> kg (OceanBridge)
  - Currency case: lowercase (SwiftHaul) -> uppercase (OceanBridge)
  - declared_value: integer cents (SwiftHaul) -> string decimal (OceanBridge)
  - Country: both use ISO alpha-2 codes
```

### C6: OceanBridge -> PingWave (Customs Clearance Notification)

OceanBridge notifies PingWave to send customs status alerts.

```
Producer: PingWave (POST /v1/notifications/send)
Consumer: OceanBridge

Data flow:
  OceanBridge sends:
    channel:        "email" or "sms"
    recipient:      customer contact from booking metadata
    template_id:    "tmpl_customs_update"
    variables:
      booking_id:   "obk_*"
      event_type:   "customs_cleared" | "customs_hold" | "customs_rejected"
      origin:       country name (resolved from ISO code)
      destination:  country name (resolved from ISO code)
      estimated_arrival: ISO 8601 UTC string
    callback_url:   OceanBridge webhook URL for delivery receipts

  PingWave returns:
    notification_id: "ntf_*"
    status:         "queued" | "sent" | ...

Invariants:
  - estimated_arrival must include timezone (Z suffix or +00:00)
  - All timestamps in UTC
```

### C7: PingWave -> ShopStream (Delivery Callback)

PingWave sends delivery status updates back to ShopStream.

```
Producer: PingWave (outbound webhook)
Consumer: ShopStream (POST /api/v1/webhooks/delivery)

Data flow:
  PingWave sends:
    event_type:     "delivery_confirmed" | "delivery_failed" | "notification_sent"
    order_id:       "ord_*"
    status:         "in_transit" | "out_for_delivery" | "delivered" | "failed"
    timestamp:      ISO 8601 UTC
    details:        human-readable description

  ShopStream expects:
    Method: POST
    Headers: X-Webhook-Signature (HMAC verification)
    Body: validated against webhook schema

Invariants:
  - ShopStream must verify X-Webhook-Signature
  - Status enum values must be snake_case (ShopStream convention)
  - Timestamp must be ISO 8601 with UTC timezone
  - ShopStream responds 200 on success, 401 on bad signature
```

### C8: PayCore -> ShopStream (Payment Webhook)

PayCore sends payment status webhooks to ShopStream.

```
Producer: PayCore (outbound webhook)
Consumer: ShopStream

Data flow:
  PayCore sends:
    event_id:       "evt_*"
    event_type:     "charge.succeeded" | "charge.failed" | "refund.succeeded"
    data:
      id:           "pay_*" or "ref_*"
      amount:       integer cents
      currency:     lowercase enum
      status:       enum
    created_at:     ISO 8601 UTC

  ShopStream expects:
    Method: POST
    Headers: X-PayCore-Signature, X-PayCore-Timestamp
    HMAC-SHA256 verification with tolerance of 300 seconds

Invariants:
  - Amount is always integer cents
  - Timestamp tolerance: 300 seconds
  - ShopStream must handle all event_type values
  - Idempotent processing (event_id used for deduplication)
```

---

## Stricture Manifest

```yaml
# .stricture-manifest.yml — Multi-company logistics ecosystem.
# 7 services, 8 cross-service contracts, 3 languages.

services:
  - id: shopstream
    language: typescript
    owner: "ShopStream Inc."
    base_url: "https://api.shopstream.io"

  - id: paycore
    language: go
    owner: "PayCore Ltd."
    base_url: "https://api.paycore.com"

  - id: vaultstore
    language: go
    owner: "VaultStore Systems"
    base_url: "https://api.vaultstore.io"

  - id: labelforge
    language: typescript
    owner: "LabelForge Corp."
    base_url: "https://api.labelforge.com"

  - id: swifthaul
    language: go
    owner: "SwiftHaul Logistics"
    base_url: "https://api.swifthaul.com"

  - id: oceanbridge
    language: python
    owner: "OceanBridge Maritime"
    base_url: "https://api.oceanbridge.io"

  - id: pingwave
    language: typescript
    owner: "PingWave Comms"
    base_url: "https://api.pingwave.io"

contracts:
  # ── C1: ShopStream -> PayCore ────────────────────────────
  - id: "c1-shopstream-paycore-charge"
    producer: paycore
    consumers: [shopstream]
    protocol: http
    endpoints:
      - path: "/v1/charges"
        method: POST
        request:
          content_type: application/json
          headers:
            Authorization: { type: string, format: "Bearer *", required: true }
            Idempotency-Key: { type: string, format: uuid, required: true }
          fields:
            amount:       { type: integer, range: [50, 99999999], required: true }
            currency:     { type: enum, values: ["usd","eur","gbp","cad","aud"], required: true }
            source_token: { type: string, format: "tok_*", required: true }
            description:  { type: string, required: false }
            metadata:     { type: object, required: false }
        response:
          fields:
            id:           { type: string, format: "pay_*", required: true }
            amount:       { type: integer, range: [50, 99999999], required: true }
            currency:     { type: enum, values: ["usd","eur","gbp","cad","aud"], required: true }
            status:       { type: enum, values: ["succeeded","pending","failed"], required: true }
            created_at:   { type: string, format: iso8601, required: true }
            metadata:     { type: object, required: true }
        status_codes: [200, 400, 402, 500]

  # ── C2: ShopStream -> VaultStore ─────────────────────────
  - id: "c2-shopstream-vaultstore-reserve"
    producer: vaultstore
    consumers: [shopstream]
    protocol: http
    endpoints:
      - path: "/api/inventory/reserve"
        method: POST
        request:
          content_type: application/json
          headers:
            Authorization: { type: string, format: "Bearer *", required: true }
          fields:
            order_id:     { type: string, required: true }
            items:        { type: array, required: true, items: { sku: string, quantity: integer } }
            warehouse_id: { type: string, format: "wh_*", required: false }
        response:
          fields:
            reservation_id: { type: string, format: "res_*", required: true }
            status:       { type: enum, values: ["reserved","partial","unavailable"], required: true }
            items:        { type: array, required: true }
            expires_at:   { type: string, format: iso8601, required: true }
        status_codes: [200, 400, 409]

  # ── C3: VaultStore -> LabelForge ─────────────────────────
  - id: "c3-vaultstore-labelforge-label"
    producer: labelforge
    consumers: [vaultstore]
    protocol: http
    endpoints:
      - path: "/v2/labels"
        method: POST
        request:
          content_type: application/json
          headers:
            Authorization: { type: string, format: "Bearer *", required: true }
          fields:
            from_address: { type: object, required: true }
            to_address:   { type: object, required: true }
            parcel:
              weight:     { type: integer, unit: grams, required: true }
              length:     { type: number, unit: centimeters, required: true }
              width:      { type: number, unit: centimeters, required: true }
              height:     { type: number, unit: centimeters, required: true }
            service_level: { type: enum, values: ["ground","express","overnight","international"], required: true }
        response:
          fields:
            label_id:     { type: string, format: "lbl_*", required: true }
            tracking_number: { type: string, required: true }
            carrier:      { type: enum, values: ["ups","fedex","usps","dhl"], required: true }
            service_level: { type: enum, values: ["ground","express","overnight","international"], required: true }
            rate:         { type: integer, unit: cents, required: true }
            label_url:    { type: string, format: url, required: true }
            estimated_delivery: { type: string, format: iso8601, required: true }
            created_at:   { type: string, format: iso8601, required: true }
        status_codes: [200, 400, 422]

  # ── C4: LabelForge -> SwiftHaul ──────────────────────────
  - id: "c4-labelforge-swifthaul-pickup"
    producer: swifthaul
    consumers: [labelforge, vaultstore]
    protocol: http
    endpoints:
      - path: "/api/pickups"
        method: POST
        request:
          content_type: application/json
          headers:
            Authorization: { type: string, format: "Bearer *", required: true }
          fields:
            tracking_number: { type: string, required: true }
            carrier:      { type: enum, values: ["ups","fedex","usps","dhl"], required: true }
            pickup_address: { type: object, required: true }
            package_details:
              weight_grams: { type: integer, unit: grams, required: true }
              length_cm:  { type: integer, unit: centimeters, required: true }
              width_cm:   { type: integer, unit: centimeters, required: true }
              height_cm:  { type: integer, unit: centimeters, required: true }
            pickup_date:  { type: string, format: "YYYY-MM-DD", required: true }
            is_international: { type: boolean, required: true }
        response:
          fields:
            pickup_id:    { type: string, format: "pku_*", required: true }
            status:       { type: enum, values: ["SCHEDULED","PICKED_UP","IN_TRANSIT","OUT_FOR_DELIVERY","DELIVERED","FAILED"], required: true }
            tracking_number: { type: string, required: true }
            estimated_delivery: { type: string, format: iso8601, required: true }
        status_codes: [200, 400, 422]

  # ── C5: SwiftHaul -> OceanBridge ─────────────────────────
  - id: "c5-swifthaul-oceanbridge-booking"
    producer: oceanbridge
    consumers: [swifthaul]
    protocol: http
    endpoints:
      - path: "/api/v1/bookings"
        method: POST
        request:
          content_type: application/json
          headers:
            Authorization: { type: string, format: "Bearer *", required: true }
          fields:
            origin_country:      { type: string, format: iso3166_alpha2, required: true }
            destination_country: { type: string, format: iso3166_alpha2, required: true }
            packages:            { type: array, required: true }
            declared_value:      { type: string, format: decimal_string, required: true }
            currency:            { type: enum, values: ["USD","EUR","GBP"], required: true }
            contents_description: { type: string, required: true }
            hs_code:             { type: string, format: "NNNN.NN.NNNN", required: false }
            customs_info:        { type: object, required: true }
        response:
          fields:
            booking_id:          { type: string, format: "obk_*", required: true }
            status:              { type: enum, values: ["pending","confirmed","in_transit","customs_hold","delivered","cancelled"], required: true }
            estimated_arrival:   { type: string, format: iso8601, required: true }
            tracking_url:        { type: string, format: url, required: true }
            created_at:          { type: string, format: iso8601, required: true }
        status_codes: [201, 400, 422]

  # ── C6: OceanBridge -> PingWave ──────────────────────────
  - id: "c6-oceanbridge-pingwave-notify"
    producer: pingwave
    consumers: [oceanbridge]
    protocol: http
    endpoints:
      - path: "/v1/notifications/send"
        method: POST
        request:
          content_type: application/json
          headers:
            Authorization: { type: string, format: "Bearer *", required: true }
          fields:
            channel:      { type: enum, values: ["sms","email","webhook"], required: true }
            recipient:    { type: string, required: true }
            template_id:  { type: string, format: "tmpl_*", required: false }
            variables:    { type: object, required: false }
            message:      { type: string, required: false }
            callback_url: { type: string, format: url, required: false }
            metadata:     { type: object, required: false }
        response:
          fields:
            notification_id: { type: string, format: "ntf_*", required: true }
            channel:      { type: enum, values: ["sms","email","webhook"], required: true }
            status:       { type: enum, values: ["queued","sent","delivered","failed","bounced"], required: true }
            created_at:   { type: string, format: iso8601, required: true }
        status_codes: [200, 400, 422]

  # ── C7: PingWave -> ShopStream (Delivery Callback) ──────
  - id: "c7-pingwave-shopstream-callback"
    producer: pingwave
    consumers: [shopstream]
    protocol: http
    direction: inbound
    verification:
      method: hmac-sha256
      header: X-PingWave-Signature
      timestamp_header: X-PingWave-Timestamp
      tolerance_seconds: 300
    endpoints:
      - path: "/api/v1/webhooks/delivery"
        method: POST
        request:
          content_type: application/json
          fields:
            event_type:   { type: enum, values: ["delivery_confirmed","delivery_failed","notification_sent"], required: true }
            order_id:     { type: string, format: "ord_*", required: true }
            status:       { type: enum, values: ["in_transit","out_for_delivery","delivered","failed"], required: true }
            timestamp:    { type: string, format: iso8601, required: true }
            details:      { type: string, required: false }
        response:
          fields:
            received:     { type: boolean, required: true }
        status_codes: [200, 400, 401]

  # ── C8: PayCore -> ShopStream (Payment Webhook) ─────────
  - id: "c8-paycore-shopstream-webhook"
    producer: paycore
    consumers: [shopstream]
    protocol: http
    direction: inbound
    verification:
      method: hmac-sha256
      header: X-PayCore-Signature
      timestamp_header: X-PayCore-Timestamp
      tolerance_seconds: 300
    endpoints:
      - path: "/api/v1/webhooks/payment"
        method: POST
        request:
          content_type: application/json
          fields:
            event_id:     { type: string, format: "evt_*", required: true }
            event_type:   { type: enum, values: ["charge.succeeded","charge.failed","refund.succeeded"], required: true }
            data:
              id:         { type: string, required: true }
              amount:     { type: integer, required: true }
              currency:   { type: enum, values: ["usd","eur","gbp","cad","aud"], required: true }
              status:     { type: enum, required: true }
            created_at:   { type: string, format: iso8601, required: true }
        response:
          fields:
            received:     { type: boolean, required: true }
        status_codes: [200, 400, 401]

conventions:
  monetary_values: integer_cents
  weight: grams
  dimensions: millimeters_internal_centimeters_external
  country_codes: iso_3166_1_alpha_2
  timestamps: iso_8601_utc
  enum_case:
    shopstream: snake_case
    paycore: snake_case
    vaultstore: snake_case
    labelforge: snake_case
    swifthaul: SCREAMING_SNAKE_CASE
    oceanbridge: snake_case
    pingwave: snake_case
```
