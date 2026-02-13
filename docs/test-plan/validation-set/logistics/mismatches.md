# Cross-Company Mismatch Scenarios (M01-M10)

Ten realistic cross-boundary bugs between the 7 logistics companies. Each mismatch includes producer code (with inline BUG comment), consumer code (showing where it breaks), Stricture rules that catch it, and production impact.

---

## M01: Float Dollars vs Integer Cents

**Boundary:** ShopStream <-> PayCore (Contract C1)

PayCore returns `amount: 2999` (integer cents). ShopStream renders the charge directly as dollars without dividing by 100, displaying "$2999.00" instead of "$29.99". The most common money bug in payment integrations.

### Producer: PayCore (Go)

```go
// paycore/charges.go — Charge handler (correct, no bug here).

func HandleCreateCharge(w http.ResponseWriter, r *http.Request) {
	var req ChargeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}
	chargeID := fmt.Sprintf("pay_%s", generateID())
	resp := ChargeResponse{
		ID:        chargeID,
		Amount:    req.Amount,    // integer cents -- contract-correct
		Currency:  req.Currency,
		Status:    "succeeded",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
```

### Consumer: ShopStream (TypeScript)

```typescript
// shopstream-order-confirmation.ts — Renders payment confirmation.

interface PayCoreCharge {
  id: string;
  amount: number;       // BUG: Developer assumes this is dollars, but it's cents
  currency: string;
  status: "succeeded" | "pending" | "failed";
  created_at: string;
}

async function renderOrderConfirmation(orderId: string, charge: PayCoreCharge): Promise<string> {
  // BUG: Treats integer cents as dollars -- displays $2999.00 instead of $29.99
  const displayAmount = `$${charge.amount.toFixed(2)}`;
  // CORRECT would be: const displayAmount = `$${(charge.amount / 100).toFixed(2)}`;

  return `<div class="confirmation">
    <h2>Order ${orderId} Confirmed</h2>
    <p>Amount charged: ${displayAmount}</p>
  </div>`;
}
```

### Expected Stricture Violations

| Rule | Reason |
|------|--------|
| CTR-strictness-parity | Consumer treats `amount` (integer cents) as dollars without conversion -- unit semantics mismatch across C1 boundary |
| CTR-response-shape | Consumer's rendering assumes float-dollar shape for a field documented as integer-cents |

### Production Impact

Customers see an amount 100x higher than what they paid. A $29.99 order displays as "$2999.00." Triggers support tickets, chargebacks, and trust erosion.

---

## M02: Weight Grams vs Kilograms

**Boundary:** VaultStore -> OceanBridge (Contract C5, via SwiftHaul)

VaultStore stores weight in grams (e.g., `500` for a 0.5kg item). SwiftHaul assigns `WeightKg: float64(pkg.WeightGrams)` without dividing by 1000. OceanBridge receives 500.0 as kilograms -- a half-ton package.

### Producer: SwiftHaul (Go)

```go
// swifthaul/handoff.go — International handoff to OceanBridge.

func HandoffToOceanBridge(
	oceanBridgeURL, oceanBridgeKey string,
	pkg PackageDetails,
	destCountry string,
	declaredValueCents int,
	currency, description string,
) error {
	// BUG: Missing / 1000.0 -- sends grams as kilograms
	// 500 grams becomes 500.0 kg instead of 0.5 kg
	weightKg := float64(pkg.WeightGrams)
	// CORRECT would be: weightKg := float64(pkg.WeightGrams) / 1000.0

	booking := OceanBridgeBookingRequest{
		OriginCountry:      "US",
		DestinationCountry: destCountry,
		Packages: []OceanBridgePackage{{
			WeightKg: weightKg, // BUG: 500.0 instead of 0.5
			LengthCm: pkg.LengthCm,
			WidthCm:  pkg.WidthCm,
			HeightCm: pkg.HeightCm,
		}},
		DeclaredValue:       fmt.Sprintf("%.2f", float64(declaredValueCents)/100.0),
		Currency:            strings.ToUpper(currency),
		ContentsDescription: description,
	}
	body, _ := json.Marshal(booking)
	_ = body
	return nil
}
```

### Consumer: OceanBridge (Python)

```python
# oceanbridge/bookings.py — Booking handler receives inflated weight.

def create_booking(data: dict[str, Any]) -> dict[str, Any]:
    req = BookingRequest(data)
    req.validate()  # Passes -- weight_kg > 0

    # A 500g package (0.5 kg) arrives as 500.0 kg
    for pkg in req.packages:
        rate = calculate_freight_rate(
            weight_kg=pkg["weight_kg"],  # 500.0 instead of 0.5
            origin=req.origin_country,
            destination=req.destination_country,
        )
    # ... create booking response

def calculate_freight_rate(weight_kg: float, origin: str, destination: str) -> float:
    base_rate_per_kg = 2.50
    return weight_kg * base_rate_per_kg  # 500.0 * 2.50 = $1250 instead of $1.25
```

### Expected Stricture Violations

| Rule | Reason |
|------|--------|
| CTR-request-shape | `weight_kg` receives a gram-magnitude value instead of kilograms -- unit mismatch against C5 contract |
| CTR-strictness-parity | Producer stores weight in grams but sends without the kg conversion documented in contract invariants |

### Production Impact

A 500g book ships as a 500kg pallet. Freight charges are 1000x too high. Customs declarations flagged for impossible weight. Shipments rejected outright by carrier.

---

## M03: Dimensions mm Sent as cm Without Conversion

**Boundary:** VaultStore -> LabelForge (Contract C3)

VaultStore stores dimensions in mm (e.g., `300` mm). Sends raw mm value to LabelForge which expects cm. LabelForge creates a label for a 300cm (3 meter) package instead of 30cm.

### Producer: VaultStore (Go)

```go
// vaultstore/shipments.go — Prepares shipment and calls LabelForge.

func PrepareShipment(labelForgeURL, labelForgeKey string, req *ShipmentRequest) (*LabelForgeResponse, error) {
	pkg := req.Packages[0]

	// BUG: Dimensions sent in mm, LabelForge expects cm
	// 300mm becomes 300cm (3 meters!) instead of 30cm
	labelReq := LabelForgeRequest{
		FromAddress: warehouseAddr,
		ToAddress:   mapAddress(req.ShipTo),
		Parcel: LabelForgeParcel{
			Weight: pkg.Weight,              // grams -> grams (correct)
			Length: pkg.Dimensions.Length,    // BUG: mm sent as cm -- 300 instead of 30
			Width:  pkg.Dimensions.Width,    // BUG: mm sent as cm -- 200 instead of 20
			Height: pkg.Dimensions.Height,   // BUG: mm sent as cm -- 150 instead of 15
			// CORRECT would be:
			// Length: pkg.Dimensions.Length / 10,  // mm -> cm
			// Width:  pkg.Dimensions.Width / 10,
			// Height: pkg.Dimensions.Height / 10,
		},
		ServiceLevel: "ground",
	}
	body, _ := json.Marshal(labelReq)
	_ = body
	return nil, nil
}
```

### Consumer: LabelForge (TypeScript)

```typescript
// labelforge-server.ts — Label generation receives inflated dimensions.

function calculateShippingRate(parcel: { weight: number; length: number; width: number; height: number }): number {
  // Dimensional weight: (L x W x H) / 5000 (in cm)
  const dimWeight = (parcel.length * parcel.width * parcel.height) / 5000;
  // With mm values: (300 * 200 * 150) / 5000 = 1,800,000 "kg"
  // With correct cm: (30 * 20 * 15) / 5000 = 1.8 kg
  const actualWeight = parcel.weight / 1000;
  const billableWeight = Math.max(dimWeight, actualWeight);
  return Math.round(billableWeight * 250); // cents per kg
}
```

### Expected Stricture Violations

| Rule | Reason |
|------|--------|
| CTR-strictness-parity | VaultStore dimensions are mm internally but C3 contract requires cm for LabelForge -- missing `/ 10` conversion at the boundary |

### Production Impact

A 30cm shoebox gets a label for a 3-meter crate. Dimensional weight inflated 1000x. Carrier rejects the label or charges freight rates for an oversized shipment.

---

## M04: Country Name vs ISO Code

**Boundary:** SwiftHaul -> OceanBridge (Contract C5)

OceanBridge expects `destination_country: "US"` (ISO 3166-1 alpha-2). SwiftHaul sends `destination_country: "United States"` (full name). OceanBridge's regex validation rejects it.

### Producer: SwiftHaul (Go)

```go
// swifthaul/handoff.go — Uses country name instead of ISO code.

func HandoffToOceanBridge(
	oceanBridgeURL, oceanBridgeKey string,
	pkg PackageDetails,
	destCountry string, // "United States" from upstream -- should be "US"
	declaredValueCents int,
	currency, description string,
) error {
	booking := OceanBridgeBookingRequest{
		OriginCountry:      "United States", // BUG: Should be "US"
		DestinationCountry: destCountry,     // BUG: "United States" instead of "US"
		// CORRECT would be:
		// OriginCountry:      "US",
		// DestinationCountry: countryNameToISO(destCountry),
		Packages: []OceanBridgePackage{{
			WeightKg: float64(pkg.WeightGrams) / 1000.0,
			LengthCm: pkg.LengthCm,
			WidthCm:  pkg.WidthCm,
			HeightCm: pkg.HeightCm,
		}},
		DeclaredValue:       fmt.Sprintf("%.2f", float64(declaredValueCents)/100.0),
		Currency:            strings.ToUpper(currency),
		ContentsDescription: description,
	}
	body, _ := json.Marshal(booking)
	_ = body
	return nil
}
```

### Consumer: OceanBridge (Python)

```python
# oceanbridge/bookings.py — Validates country codes, rejects full names.

ISO_COUNTRY_PATTERN = re.compile(r"^[A-Z]{2}$")

class BookingRequest:
    def validate(self) -> None:
        # "United States" fails this regex -- only 2-letter codes match
        if not ISO_COUNTRY_PATTERN.match(self.origin_country):
            raise ValueError(
                f"Invalid origin_country: {self.origin_country} "
                f"(expected ISO 3166-1 alpha-2, e.g. 'US')"
            )
        if not ISO_COUNTRY_PATTERN.match(self.destination_country):
            raise ValueError(
                f"Invalid destination_country: {self.destination_country} "
                f"(expected ISO 3166-1 alpha-2, e.g. 'DE')"
            )
```

### Expected Stricture Violations

| Rule | Reason |
|------|--------|
| CTR-request-shape | `destination_country` sends a free-text name where C5 specifies `format: iso3166_alpha2` |
| CTR-strictness-parity | Producer uses country names internally but contract boundary requires ISO alpha-2 -- no mapping function exists |

### Production Impact

Every international handoff fails with 400 validation error. No international shipments are booked. Packages pile up at domestic hub.

---

## M05: Enum Case Mismatch (SCREAMING_SNAKE vs snake_case)

**Boundary:** SwiftHaul -> ShopStream via PingWave (Contracts C4, C7)

SwiftHaul returns `"IN_TRANSIT"` (SCREAMING_SNAKE_CASE). PingWave relays it unchanged. ShopStream compares against `"in_transit"` (snake_case). The comparison never matches, so all delivery status updates are silently ignored.

### Producer: SwiftHaul (Go)

```go
// swifthaul/pickups.go — Returns SCREAMING_SNAKE_CASE status.

resp := PickupResponse{
	PickupID:          fmt.Sprintf("pku_%d", time.Now().UnixNano()),
	Status:            "IN_TRANSIT",  // SCREAMING_SNAKE_CASE -- SwiftHaul convention
	TrackingNumber:    req.TrackingNumber,
	EstimatedDelivery: time.Now().AddDate(0, 0, 7).UTC().Format(time.RFC3339),
}
```

### Relay: PingWave (TypeScript)

```typescript
// pingwave-relay.ts — Relays SwiftHaul status to ShopStream without case conversion.

async function relayDeliveryStatus(
  callbackUrl: string,
  orderId: string,
  swiftHaulStatus: string,  // "IN_TRANSIT" from SwiftHaul
  webhookSecret: string,
): Promise<void> {
  // BUG: Passes SCREAMING_SNAKE_CASE status directly to ShopStream
  const payload: DeliveryCallbackPayload = {
    event_type: "delivery_confirmed",
    order_id: orderId,
    status: swiftHaulStatus as "in_transit" | "out_for_delivery" | "delivered" | "failed",
    // BUG: "IN_TRANSIT" cast silently passes at runtime
    // CORRECT would be: status: swiftHaulStatus.toLowerCase() as ...
    timestamp: new Date().toISOString(),
  };
  // ... sign and send
}
```

### Consumer: ShopStream (TypeScript)

```typescript
// shopstream-webhook-handler.ts — Expects snake_case status values.

function handleDeliveryWebhook(event: DeliveryWebhookEvent): void {
  // BUG: snake_case comparisons vs SCREAMING_SNAKE_CASE values
  // "IN_TRANSIT" !== "in_transit" -- none of these cases ever match
  switch (event.status) {
    case "in_transit":       // Never reached for "IN_TRANSIT"
      updateOrderStatus(event.order_id, "shipped");
      break;
    case "out_for_delivery": // Never reached for "OUT_FOR_DELIVERY"
      updateOrderStatus(event.order_id, "shipped");
      break;
    case "delivered":        // Never reached for "DELIVERED"
      updateOrderStatus(event.order_id, "delivered");
      break;
    case "failed":           // Never reached for "FAILED"
      initiateRefund(event.order_id);
      break;
    default:
      // All SwiftHaul statuses fall through here -- silently dropped
      console.warn(`Unknown delivery status: ${event.status}`);
  }
}
```

### Expected Stricture Violations

| Rule | Reason |
|------|--------|
| CTR-strictness-parity | SwiftHaul uses `IN_TRANSIT` but C7 contract and ShopStream expect `in_transit` -- enum case convention mismatch across boundaries |

### Production Impact

Order status never updates past "shipped." Deliveries complete but orders stay in "processing." Failed deliveries never trigger refunds. No errors thrown -- all transitions silently dropped.

---

## M06: String Decimal vs Integer Cents for Customs

**Boundary:** SwiftHaul -> OceanBridge (Contract C5)

OceanBridge expects `declared_value: "29.99"` (string decimal). SwiftHaul uses `strconv.Itoa(2999)` producing `"2999"` instead of `fmt.Sprintf("%.2f", float64(2999)/100.0)` producing `"29.99"`.

### Producer: SwiftHaul (Go)

```go
// swifthaul/handoff.go — Converts declared value for OceanBridge customs.

func HandoffToOceanBridge(
	oceanBridgeURL, oceanBridgeKey string,
	pkg PackageDetails,
	destCountry string,
	declaredValueCents int, // 2999 (i.e., $29.99)
	currency, description string,
) error {
	// BUG: Converts cents integer to string directly -- "2999" instead of "29.99"
	declaredValueStr := strconv.Itoa(declaredValueCents)
	// CORRECT would be: declaredValueStr := fmt.Sprintf("%.2f", float64(declaredValueCents)/100.0)

	booking := OceanBridgeBookingRequest{
		OriginCountry:      "US",
		DestinationCountry: destCountry,
		Packages: []OceanBridgePackage{{
			WeightKg: float64(pkg.WeightGrams) / 1000.0,
			LengthCm: pkg.LengthCm, WidthCm: pkg.WidthCm, HeightCm: pkg.HeightCm,
		}},
		DeclaredValue:       declaredValueStr, // BUG: "2999" not "29.99"
		Currency:            strings.ToUpper(currency),
		ContentsDescription: description,
	}
	body, _ := json.Marshal(booking)
	_ = body
	return nil
}
```

### Consumer: OceanBridge (Python)

```python
# oceanbridge/bookings.py — Parses declared_value as decimal string.

def validate(self) -> None:
    # "2999" is valid as a decimal -- Decimal("2999") succeeds
    # but it represents $2999.00, not $29.99
    try:
        val = Decimal(self.declared_value)
        if val <= 0:
            raise ValueError(f"declared_value must be positive, got {self.declared_value}")
    except Exception as e:
        raise ValueError(f"Invalid declared_value: {self.declared_value}") from e

def generate_customs_declaration(req: BookingRequest) -> dict[str, Any]:
    declared = Decimal(req.declared_value)  # "2999" -> $2,999.00
    duty_rate = Decimal("0.05")
    duty_amount = declared * duty_rate  # $149.95 instead of $1.50
    return {
        "declared_value": str(declared),
        "duty_amount": str(duty_amount),  # "149.9500" instead of "1.4950"
    }
```

### Expected Stricture Violations

| Rule | Reason |
|------|--------|
| CTR-request-shape | `declared_value` should be `"29.99"` per C5 contract but receives `"2999"` (integer-as-string) |
| CTR-strictness-parity | SwiftHaul uses integer cents internally but contract requires decimal dollar string -- `strconv.Itoa` instead of cents-to-dollars-to-string |

### Production Impact

Customs declarations show 100x the actual value. Import duties calculated on inflated amount ($149.95 instead of $1.50). Shipments flagged for scrutiny. Customers charged excessive duty fees.

---

## M07: Missing Webhook Signature Verification

**Boundary:** PingWave -> ShopStream (Contract C7)

PingWave sends `X-PingWave-Signature` and `X-PingWave-Timestamp` headers for HMAC-SHA256 verification. ShopStream's webhook handler skips verification entirely.

### Producer: PingWave (TypeScript)

```typescript
// pingwave-client.ts — Sends signed delivery callbacks (correct).

async function sendDeliveryCallback(
  callbackUrl: string,
  payload: DeliveryCallbackPayload,
  webhookSecret: string,
): Promise<void> {
  const body = JSON.stringify(payload);
  const timestamp = Math.floor(Date.now() / 1000).toString();
  const signedPayload = `${timestamp}.${body}`;
  const signature = crypto
    .createHmac("sha256", webhookSecret)
    .update(signedPayload)
    .digest("hex");

  await fetch(callbackUrl, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "X-PingWave-Signature": signature,
      "X-PingWave-Timestamp": timestamp,
    },
    body,
  });
}
```

### Consumer: ShopStream (TypeScript)

```typescript
// shopstream-webhook-handler.ts — Handles callbacks WITHOUT signature verification.

import type { Request, Response } from "express";

// BUG: No signature verification -- any POST to this endpoint is accepted
function handleDeliveryWebhook(req: Request, res: Response): void {
  // MISSING: Should verify X-PingWave-Signature header
  // MISSING: Should check X-PingWave-Timestamp for replay protection
  //
  // CORRECT would include:
  // const signature = req.headers["x-pingwave-signature"] as string;
  // const timestamp = req.headers["x-pingwave-timestamp"] as string;
  // if (!verifyWebhookSignature(rawBody, signature, timestamp, webhookSecret)) {
  //   res.status(401).json({ error: { code: "invalid_signature", message: "..." } });
  //   return;
  // }

  const event = req.body as DeliveryWebhookEvent;

  if (!event.order_id || !event.status) {
    res.status(400).json({ error: { code: "invalid_payload", message: "Missing required fields" } });
    return;
  }

  // Processes unverified webhook -- attacker can forge delivery status updates
  switch (event.status) {
    case "delivered": markOrderDelivered(event.order_id, event.timestamp); break;
    case "failed":    initiateRefund(event.order_id); break;
    default:          updateTrackingStatus(event.order_id, event.status);
  }
  res.status(200).json({ received: true });
}
```

### Expected Stricture Violations

| Rule | Reason |
|------|--------|
| CTR-manifest-conformance | C7 specifies `verification.method: hmac-sha256` and `verification.header: X-PingWave-Signature`, but the consumer does not verify these headers |
| CTR-request-shape | Contract requires `X-Webhook-Signature` verification on inbound webhooks but handler accepts unsigned requests |

### Production Impact

Unauthenticated endpoint. Attacker can forge delivery updates: mark orders "delivered" to suppress refunds, mark orders "failed" to trigger unauthorized refunds, or inject fake tracking data. Privilege escalation vulnerability.

---

## M08: Timezone-Naive Datetime

**Boundary:** OceanBridge -> PingWave (Contract C6)

OceanBridge uses `datetime.now().isoformat()` (no timezone) producing `"2024-01-20T14:00:00"`. PingWave treats it as UTC, but the OceanBridge server runs in CET (UTC+01:00). Customer gets wrong ETA.

### Producer: OceanBridge (Python)

```python
# oceanbridge/bookings.py — Generates timezone-naive estimated_arrival.

from datetime import datetime  # BUG: timezone not imported

def create_booking(data: dict[str, Any]) -> dict[str, Any]:
    req = BookingRequest(data)
    req.validate()

    # BUG: datetime.now() without timezone -- produces naive datetime
    # On a CET server, 14:00 local = 13:00 UTC, but no timezone in output
    eta = datetime.now().replace(hour=14, minute=0, second=0, microsecond=0)
    estimated_arrival = eta.isoformat()  # "2024-01-20T14:00:00" -- no Z, no +00:00
    # CORRECT would be:
    # from datetime import timezone
    # eta = datetime.now(timezone.utc).replace(hour=14, minute=0, second=0, microsecond=0)
    # estimated_arrival = eta.isoformat()  # "2024-01-20T14:00:00+00:00"

    resp = BookingResponse(
        booking_id=f"obk_{int(datetime.now().timestamp() * 1000)}",
        status="pending",
        estimated_arrival=estimated_arrival,
        tracking_url=f"https://track.oceanbridge.io/obk_123",
    )
    return resp.to_dict()
```

### Consumer: PingWave (TypeScript)

```typescript
// pingwave-notification.ts — Treats estimated_arrival as UTC.

function formatEstimatedArrival(estimatedArrival: string, customerTimezone: string): string {
  // Assumes UTC per contract -- but source was CET (14:00 CET = 13:00 UTC)
  // "2024-01-20T14:00:00" without timezone -- JS Date may treat as UTC or local
  const eta = new Date(estimatedArrival);
  if (isNaN(eta.getTime())) {
    throw new Error(`Invalid estimated_arrival: ${estimatedArrival}`);
  }
  // Customer sees 14:00 UTC = 15:00 CET -- 1 hour late
  return eta.toLocaleString("en-US", { timeZone: customerTimezone, dateStyle: "medium", timeStyle: "short" });
}
```

### Expected Stricture Violations

| Rule | Reason |
|------|--------|
| CTR-strictness-parity | OceanBridge produces timezone-naive datetimes but C6 contract requires ISO 8601 with UTC -- `estimated_arrival` crosses boundary without timezone |

### Production Impact

ETAs off by OceanBridge server's UTC offset (1-2 hours depending on DST). Timestamp looks valid -- only timezone is missing. Customers plan around wrong ETAs, leading to missed deliveries.

---

## M09: Date Format Degradation Across 3 Hops

**Boundary:** PayCore -> ShopStream -> VaultStore (Contracts C8, C2)

PayCore sends `"2024-01-15T10:30:00Z"` (ISO 8601). ShopStream parses with `new Date()`, re-serializes with `.toLocaleString()` producing `"1/15/2024, 10:30:00 AM"`. VaultStore fails to parse with `time.Parse(time.RFC3339, ...)`.

### Hop 1 -- PayCore (Go) -- correct

```go
// paycore/webhooks.go — Sends ISO 8601 timestamp (correct).

payload := WebhookPayload{
	EventID:   fmt.Sprintf("evt_%s", generateID()),
	EventType: eventType,
	Data:      WebhookData{ID: charge.ID, Amount: charge.Amount, Currency: charge.Currency, Status: charge.Status},
	CreatedAt: time.Now().UTC().Format(time.RFC3339), // "2024-01-15T10:30:00Z"
}
```

### Hop 2 -- ShopStream (TypeScript) -- degrades the format

```typescript
// shopstream-payment-handler.ts — Receives PayCore webhook, forwards to VaultStore.

async function handlePaymentWebhook(event: PayCoreWebhookEvent): Promise<void> {
  if (event.event_type === "charge.succeeded") {
    const createdAt = new Date(event.created_at);  // Parse works fine

    // BUG: Re-serialize with toLocaleString() -- locale-dependent format
    // "2024-01-15T10:30:00Z" becomes "1/15/2024, 10:30:00 AM"
    const formattedDate = createdAt.toLocaleString();
    // CORRECT would be: const formattedDate = createdAt.toISOString();

    await notifyVaultStore({
      order_id: event.data.id,
      payment_status: event.data.status,
      amount: event.data.amount,
      created_at: formattedDate,  // BUG: "1/15/2024, 10:30:00 AM"
    });
  }
}
```

### Hop 3 -- VaultStore (Go) -- fails to parse

```go
// vaultstore/payment_confirmation.go — Receives degraded timestamp.

type PaymentConfirmation struct {
	OrderID       string `json:"order_id"`
	PaymentStatus string `json:"payment_status"`
	Amount        int    `json:"amount"`
	CreatedAt     string `json:"created_at"` // expects RFC 3339
}

func HandlePaymentConfirmation(w http.ResponseWriter, r *http.Request) {
	var conf PaymentConfirmation
	json.NewDecoder(r.Body).Decode(&conf)

	// BUG: Tries to parse "1/15/2024, 10:30:00 AM" as RFC 3339 -- always fails
	// Error: cannot parse "1/15/2024, 10:30:00 AM" as "2006-01-02T15:04:05Z07:00"
	parsedTime, err := time.Parse(time.RFC3339, conf.CreatedAt)
	if err != nil {
		writeShipmentError(w, http.StatusBadRequest, "invalid_timestamp",
			fmt.Sprintf("created_at must be RFC 3339: %s", err.Error()))
		return
	}
	_ = parsedTime
	// Never reached -- timestamp parse always fails
}
```

### Expected Stricture Violations

| Rule | Reason |
|------|--------|
| CTR-strictness-parity | ShopStream receives ISO 8601 from PayCore but re-serializes with `.toLocaleString()`, degrading the format before forwarding |
| CTR-response-shape | ShopStream's outbound payload uses locale-dependent date string where contracts require ISO 8601 |

### Production Impact

Every payment confirmation from ShopStream to VaultStore fails with 400. VaultStore never releases inventory for fulfillment. Orders stay in "paid" status indefinitely despite successful payment.

---

## M10: Implicit Callback Contract (Undocumented Endpoint)

**Boundary:** PingWave -> ShopStream (Contract C7)

PingWave sends `event_type` and `order_id` (snake_case). ShopStream expects `type` and `orderId` (camelCase). Neither side has the other's contract in their manifest.

### Producer: PingWave (TypeScript)

```typescript
// pingwave-callback.ts — Sends snake_case fields per PingWave convention.

async function sendDeliveryCallback(
  callbackUrl: string,
  orderId: string,
  status: string,
  webhookSecret: string,
): Promise<void> {
  const payload: DeliveryCallbackPayload = {
    event_type: "delivery_confirmed",   // PingWave calls it "event_type"
    order_id: orderId,                  // PingWave calls it "order_id"
    status: status as DeliveryCallbackPayload["status"],
    timestamp: new Date().toISOString(),
    details: "Package delivered successfully",
  };
  const body = JSON.stringify(payload);
  const ts = Math.floor(Date.now() / 1000).toString();
  const signature = crypto.createHmac("sha256", webhookSecret).update(`${ts}.${body}`).digest("hex");

  await fetch(callbackUrl, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "X-PingWave-Signature": signature,
      "X-PingWave-Timestamp": ts,
    },
    body,
  });
}
```

### Consumer: ShopStream (TypeScript)

```typescript
// shopstream-callback-handler.ts — Expects different field names than PingWave sends.

import type { Request, Response } from "express";

// BUG: Expects camelCase and "type" instead of "event_type"
interface ExpectedCallbackPayload {
  type: string;           // BUG: PingWave sends "event_type", not "type"
  orderId: string;        // BUG: PingWave sends "order_id", not "orderId"
  status: string;
  timestamp: string;
  details?: string;
}
// CORRECT would match PingWave's contract:
// interface ExpectedCallbackPayload {
//   event_type: string;
//   order_id: string;
//   ...
// }

function handleDeliveryCallback(req: Request, res: Response): void {
  // Signature verification works correctly (omitted for brevity)
  const event = req.body as ExpectedCallbackPayload;

  // BUG: event.type is undefined (PingWave sent event_type)
  if (!event.type) {
    // This fires for every callback -- "type" is always undefined
    res.status(400).json({ error: { code: "invalid_payload", message: "Missing required field: type" } });
    return;
  }

  // BUG: event.orderId is undefined (PingWave sent order_id)
  if (!event.orderId) {
    res.status(400).json({ error: { code: "invalid_payload", message: "Missing required field: orderId" } });
    return;
  }

  // Never reached due to field name mismatches
  processDeliveryUpdate(event.type, event.orderId, event.status);
  res.status(200).json({ received: true });
}
```

### Expected Stricture Violations

| Rule | Reason |
|------|--------|
| CTR-manifest-conformance | Callback contract not explicitly defined in either manifest -- PingWave sends `event_type`/`order_id`, ShopStream expects `type`/`orderId` with no shared contract to reconcile |

### Production Impact

Every delivery callback returns 400. ShopStream never receives delivery updates. Orders stuck in "shipped" status after delivery. PingWave retries generate unnecessary load. Debugging is hard because both sides have valid code -- the bug exists only in the implicit contract between them.

---

## Detection Summary Table

| Mismatch | Rules Triggered | Severity | Auto-fixable? |
|----------|----------------|----------|---------------|
| M01: Float dollars vs integer cents | CTR-strictness-parity, CTR-response-shape | Critical | Yes -- insert `/ 100` before rendering |
| M02: Weight grams vs kilograms | CTR-request-shape, CTR-strictness-parity | Critical | Yes -- insert `/ 1000.0` conversion |
| M03: Dimensions mm sent as cm | CTR-strictness-parity | High | Yes -- insert `/ 10` conversion |
| M04: Country name vs ISO code | CTR-request-shape, CTR-strictness-parity | High | No -- requires country name-to-code mapping table |
| M05: Enum case mismatch | CTR-strictness-parity | High | Yes -- insert `.toLowerCase()` or case-insensitive compare |
| M06: String decimal vs integer cents | CTR-request-shape, CTR-strictness-parity | Critical | Yes -- replace `strconv.Itoa` with `fmt.Sprintf("%.2f", cents/100.0)` |
| M07: Missing webhook signature | CTR-manifest-conformance, CTR-request-shape | Critical | No -- requires implementing HMAC verification logic |
| M08: Timezone-naive datetime | CTR-strictness-parity | Medium | Yes -- replace `datetime.now()` with `datetime.now(timezone.utc)` |
| M09: Date format degradation | CTR-strictness-parity, CTR-response-shape | High | Yes -- replace `.toLocaleString()` with `.toISOString()` |
| M10: Implicit callback contract | CTR-manifest-conformance | High | No -- requires contract negotiation between teams |
