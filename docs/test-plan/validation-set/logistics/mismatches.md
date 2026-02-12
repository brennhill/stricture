# Cross-Company Mismatch Scenarios (M01-M10)

Ten realistic cross-boundary bugs between the 7 logistics companies. Each mismatch includes the producer code (with inline BUG comment), the consumer code (showing where it breaks), the Stricture rules that catch it, and the production impact.

---

## M01: Float Dollars vs Integer Cents

**Boundary:** ShopStream <-> PayCore (Contract C1)

**What's wrong:** PayCore returns `amount: 2999` (integer cents, per the contract). ShopStream's order confirmation page renders the charge amount directly as dollars without dividing by 100, displaying "$2999.00" instead of "$29.99".

**Why it's realistic:** This is the single most common money bug in payment integrations. A developer reads the PayCore response and assumes the amount is already in dollars because "$29.99" is what the customer sees. The contract says "integer cents" but the rendering code was written by a frontend developer who never read the API spec.

### Producer: PayCore (Go)

```go
// paycore/charges.go — Charge handler (correct, no bug here).

func HandleCreateCharge(w http.ResponseWriter, r *http.Request) {
	var req ChargeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	if err := validateChargeRequest(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	chargeID := fmt.Sprintf("pay_%s", generateID())
	now := time.Now().UTC().Format(time.RFC3339)

	resp := ChargeResponse{
		ID:        chargeID,
		Amount:    req.Amount,    // integer cents -- contract-correct
		Currency:  req.Currency,
		Status:    "succeeded",
		CreatedAt: now,
		Metadata:  req.Metadata,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
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
  metadata: Record<string, string>;
}

async function renderOrderConfirmation(orderId: string, charge: PayCoreCharge): Promise<string> {
  // BUG: Treats integer cents as dollars -- displays $2999.00 instead of $29.99
  const displayAmount = `$${charge.amount.toFixed(2)}`;
  // CORRECT would be: const displayAmount = `$${(charge.amount / 100).toFixed(2)}`;

  return `
    <div class="confirmation">
      <h2>Order ${orderId} Confirmed</h2>
      <p>Amount charged: ${displayAmount}</p>
      <p>Payment ID: ${charge.id}</p>
      <p>Status: ${charge.status}</p>
    </div>
  `;
}

// Also affects the order summary API response
async function getOrderSummary(charge: PayCoreCharge): Promise<{
  order_total: string;
  payment_status: string;
}> {
  return {
    // BUG: Same cents-as-dollars mistake in the API response
    order_total: charge.amount.toFixed(2),
    // CORRECT would be: order_total: (charge.amount / 100).toFixed(2),
    payment_status: charge.status,
  };
}
```

### Expected Stricture Violations

| Rule | Reason |
|------|--------|
| CTR-strictness-parity | Consumer treats `amount` (integer cents) as dollars without conversion -- unit semantics mismatch across the C1 contract boundary |
| CTR-response-shape | Consumer's rendering assumes float-dollar shape for a field documented as integer-cents in the contract |

### Production Impact

Customers see an amount 100x higher than what they paid. A $29.99 order displays as "$2999.00" on the confirmation page. This triggers support tickets, chargebacks, and trust erosion. If the bug also affects the refund calculation path, refund requests could be submitted for 100x the correct amount.

---

## M02: Weight Grams vs Kilograms

**Boundary:** VaultStore -> OceanBridge (Contract C5, via SwiftHaul)

**What's wrong:** VaultStore stores weight in grams (e.g., `500` grams for a 0.5kg item). When SwiftHaul hands off to OceanBridge, the code assigns `WeightKg: float64(pkg.WeightGrams)` without dividing by 1000. OceanBridge receives `weight_kg: 500.0` and interprets it as 500 kilograms -- a half-ton package.

**Why it's realistic:** Unit conversion bugs at service boundaries are a top cause of logistics failures. The field is named `weight_kg` on the OceanBridge side, but the developer copies the gram value directly into it without conversion, possibly because they confused the variable name with its actual unit.

### Producer: SwiftHaul (Go) -- sends to OceanBridge

```go
// swifthaul/handoff.go — International handoff to OceanBridge.

func HandoffToOceanBridge(
	oceanBridgeURL, oceanBridgeKey string,
	pickupID string,
	pkg PackageDetails,
	destCountry string,
	declaredValueCents int,
	currency string,
	description string,
) error {
	// BUG: Missing / 1000.0 -- sends grams as kilograms
	// 500 grams becomes 500.0 kg instead of 0.5 kg
	weightKg := float64(pkg.WeightGrams)
	// CORRECT would be: weightKg := float64(pkg.WeightGrams) / 1000.0

	declaredValueStr := fmt.Sprintf("%.2f", float64(declaredValueCents)/100.0)
	currencyUpper := strings.ToUpper(currency)

	booking := OceanBridgeBookingRequest{
		OriginCountry:      "US",
		DestinationCountry: destCountry,
		Packages: []OceanBridgePackage{{
			WeightKg: weightKg, // BUG: 500.0 instead of 0.5
			LengthCm: pkg.LengthCm,
			WidthCm:  pkg.WidthCm,
			HeightCm: pkg.HeightCm,
		}},
		DeclaredValue:       declaredValueStr,
		Currency:            currencyUpper,
		ContentsDescription: description,
		CustomsInfo: CustomsInfo{
			SenderName:    "VaultStore Warehouse",
			RecipientName: "Customer",
		},
	}

	body, err := json.Marshal(booking)
	if err != nil {
		return fmt.Errorf("marshal booking request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, oceanBridgeURL+"/api/v1/bookings", nil)
	if err != nil {
		return fmt.Errorf("create booking request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+oceanBridgeKey)
	_ = body

	return nil
}
```

### Consumer: OceanBridge (Python)

```python
# oceanbridge/bookings.py — Booking handler receives inflated weight.

def create_booking(data: dict[str, Any]) -> dict[str, Any]:
    req = BookingRequest(data)
    req.validate()  # Passes validation because weight_kg > 0

    # OceanBridge uses weight_kg to calculate shipping rate and customs fees.
    # A 500g package (0.5 kg) arrives as 500.0 kg.
    # This causes:
    #   1. Freight rate calculated for a half-ton shipment
    #   2. Container space allocation for 500 kg
    #   3. Customs declaration with wildly wrong weight
    for pkg in req.packages:
        rate = calculate_freight_rate(
            weight_kg=pkg["weight_kg"],  # 500.0 instead of 0.5
            origin=req.origin_country,
            destination=req.destination_country,
        )

    booking_id = f"obk_{int(datetime.now(timezone.utc).timestamp() * 1000)}"
    eta = datetime.now(timezone.utc).replace(
        hour=14, minute=0, second=0, microsecond=0,
    )
    estimated_arrival = eta.isoformat()

    resp = BookingResponse(
        booking_id=booking_id,
        status="pending",
        estimated_arrival=estimated_arrival,
        tracking_url=f"https://track.oceanbridge.io/{booking_id}",
    )

    return resp.to_dict()


def calculate_freight_rate(
    weight_kg: float,
    origin: str,
    destination: str,
) -> float:
    """Rate per kg -- a 500 kg shipment costs 1000x more than 0.5 kg."""
    base_rate_per_kg = 2.50
    return weight_kg * base_rate_per_kg  # 500.0 * 2.50 = $1250 instead of $1.25
```

### Expected Stricture Violations

| Rule | Reason |
|------|--------|
| CTR-request-shape | Field `weight_kg` receives a value in grams (integer magnitude) instead of kilograms (float, divided by 1000) -- type/unit mismatch against the C5 contract |
| CTR-strictness-parity | Producer (SwiftHaul) stores weight in grams but sends to consumer (OceanBridge) without the kg conversion documented in the contract invariants |

### Production Impact

A 500g book ships as a 500kg pallet. Freight charges are 1000x too high ($1,250 instead of $1.25). Customs declarations are flagged for impossible weight. Container space allocation is wildly incorrect, causing real logistics congestion. The shipment may be rejected outright by the carrier.

---

## M03: Dimensions mm Sent as cm Without Conversion

**Boundary:** VaultStore -> LabelForge (Contract C3)

**What's wrong:** VaultStore stores dimensions in millimeters (e.g., `length: 300` mm for a 30cm box). When calling LabelForge's label API, the code sends the raw mm value without dividing by 10. LabelForge expects centimeters, so it creates a label for a 300cm (3 meter) package instead of a 30cm package.

**Why it's realistic:** The internal storage unit (mm) differs from the external API unit (cm). A developer building the integration copies the dimension fields directly, missing the conversion comment in the contract spec.

### Producer: VaultStore (Go)

```go
// vaultstore/shipments.go — Prepares shipment and calls LabelForge.

func PrepareShipment(labelForgeURL, labelForgeKey string, req *ShipmentRequest) (*LabelForgeResponse, error) {
	warehouseAddr := LabelForgeAddress{
		Name:    "VaultStore Warehouse",
		Street1: "123 Warehouse Blvd",
		City:    "Dallas",
		State:   "TX",
		Zip:     "75201",
		Country: "US",
	}

	pkg := req.Packages[0]

	// BUG: Dimensions sent in mm, LabelForge expects cm
	// 300mm becomes 300cm (3 meters!) instead of 30cm
	labelReq := LabelForgeRequest{
		FromAddress: warehouseAddr,
		ToAddress: LabelForgeAddress{
			Name:    req.ShipTo.Name,
			Street1: req.ShipTo.Line1,
			City:    req.ShipTo.City,
			State:   req.ShipTo.State,
			Zip:     req.ShipTo.PostalCode,
			Country: req.ShipTo.Country,
		},
		Parcel: LabelForgeParcel{
			Weight: pkg.Weight,              // grams -> grams (correct)
			Length: pkg.Dimensions.Length,    // BUG: mm sent as cm -- 300 instead of 30
			Width:  pkg.Dimensions.Width,    // BUG: mm sent as cm -- 200 instead of 20
			Height: pkg.Dimensions.Height,   // BUG: mm sent as cm -- 150 instead of 15
			// CORRECT would be:
			// Length: pkg.Dimensions.Length / 10,  // mm -> cm
			// Width:  pkg.Dimensions.Width / 10,   // mm -> cm
			// Height: pkg.Dimensions.Height / 10,  // mm -> cm
		},
		ServiceLevel: "ground",
	}

	body, err := json.Marshal(labelReq)
	if err != nil {
		return nil, fmt.Errorf("marshal label request: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, labelForgeURL+"/v2/labels", nil)
	if err != nil {
		return nil, fmt.Errorf("create label request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+labelForgeKey)
	_ = body

	return nil, nil
}
```

### Consumer: LabelForge (TypeScript)

```typescript
// labelforge-server.ts — Label generation handler receives inflated dimensions.

interface CreateLabelBody {
  from_address: LabelAddress;
  to_address: LabelAddress;
  parcel: {
    weight: number;   // grams (correct)
    length: number;   // centimeters -- but receives 300 (mm value)
    width: number;    // centimeters -- but receives 200 (mm value)
    height: number;   // centimeters -- but receives 150 (mm value)
  };
  service_level: "ground" | "express" | "overnight" | "international";
}

function validateParcel(parcel: CreateLabelBody["parcel"]): void {
  if (!Number.isInteger(parcel.weight) || parcel.weight < 1) {
    throw new Error(`Weight must be a positive integer in grams, got ${parcel.weight}`);
  }
  // Dimensions pass validation because 300 > 1 -- no upper bound check
  if (parcel.length < 1 || parcel.width < 1 || parcel.height < 1) {
    throw new Error("Dimensions must be positive numbers in centimeters");
  }
}

function calculateShippingRate(parcel: CreateLabelBody["parcel"]): number {
  // Dimensional weight: (L x W x H) / 5000 (in cm)
  const dimWeight = (parcel.length * parcel.width * parcel.height) / 5000;
  // With mm values: (300 * 200 * 150) / 5000 = 1,800,000 "kg"
  // With correct cm: (30 * 20 * 15) / 5000 = 1.8 kg
  const actualWeight = parcel.weight / 1000; // grams to kg
  const billableWeight = Math.max(dimWeight, actualWeight);

  // Rate per kg
  return Math.round(billableWeight * 250); // cents per kg
}
```

### Expected Stricture Violations

| Rule | Reason |
|------|--------|
| CTR-strictness-parity | VaultStore dimensions are in mm internally but the C3 contract requires cm for LabelForge -- missing `/ 10` conversion at the boundary |

### Production Impact

A 30cm x 20cm x 15cm shoebox-sized package gets a label for a 3m x 2m x 1.5m crate. The dimensional weight calculation inflates the shipping rate by 1000x. The carrier either rejects the label as physically impossible, or charges freight rates for an oversized shipment. Every VaultStore shipment is mispriced.

---

## M04: Country Name vs ISO Code

**Boundary:** SwiftHaul -> OceanBridge (Contract C5)

**What's wrong:** OceanBridge expects `destination_country: "US"` (ISO 3166-1 alpha-2 code). SwiftHaul sends `destination_country: "United States"` (the full English country name). OceanBridge's regex validation rejects it because `"United States"` does not match the `^[A-Z]{2}$` pattern.

**Why it's realistic:** Country representation is a classic interoperability gap. Developers who work with user-facing forms often have country names in their data model. When integrating with a logistics API that requires ISO codes, they forget to map the name to the code.

### Producer: SwiftHaul (Go)

```go
// swifthaul/handoff.go — Maps internal country data to OceanBridge request.

// BUG: Uses full country name instead of ISO 3166-1 alpha-2 code
func HandoffToOceanBridge(
	oceanBridgeURL, oceanBridgeKey string,
	pickupID string,
	pkg PackageDetails,
	destCountry string, // "United States" from upstream -- should be "US"
	declaredValueCents int,
	currency string,
	description string,
) error {
	weightKg := float64(pkg.WeightGrams) / 1000.0
	declaredValueStr := fmt.Sprintf("%.2f", float64(declaredValueCents)/100.0)
	currencyUpper := strings.ToUpper(currency)

	booking := OceanBridgeBookingRequest{
		OriginCountry:      "United States", // BUG: Should be "US"
		DestinationCountry: destCountry,     // BUG: "United States" instead of "US"
		// CORRECT would be:
		// OriginCountry:      "US",
		// DestinationCountry: countryNameToISO(destCountry),  // "US", "DE", "JP", etc.
		Packages: []OceanBridgePackage{{
			WeightKg: weightKg,
			LengthCm: pkg.LengthCm,
			WidthCm:  pkg.WidthCm,
			HeightCm: pkg.HeightCm,
		}},
		DeclaredValue:       declaredValueStr,
		Currency:            currencyUpper,
		ContentsDescription: description,
		CustomsInfo: CustomsInfo{
			SenderName:    "VaultStore Warehouse",
			RecipientName: "Customer",
		},
	}

	body, err := json.Marshal(booking)
	if err != nil {
		return fmt.Errorf("marshal booking request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, oceanBridgeURL+"/api/v1/bookings", nil)
	if err != nil {
		return fmt.Errorf("create booking request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+oceanBridgeKey)
	_ = body

	return nil
}
```

### Consumer: OceanBridge (Python)

```python
# oceanbridge/bookings.py — Validates country codes and rejects full names.

ISO_COUNTRY_PATTERN = re.compile(r"^[A-Z]{2}$")

class BookingRequest:
    def __init__(self, data: dict[str, Any]) -> None:
        self.origin_country: str = data["origin_country"]
        self.destination_country: str = data["destination_country"]
        self.packages: list[dict[str, Any]] = data["packages"]
        self.declared_value: str = data["declared_value"]
        self.currency: str = data["currency"]
        self.contents_description: str = data["contents_description"]
        self.hs_code: str | None = data.get("hs_code")
        self.customs_info: dict[str, str] = data["customs_info"]

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
        # ... remaining validation
```

### Expected Stricture Violations

| Rule | Reason |
|------|--------|
| CTR-request-shape | Field `destination_country` sends a free-text country name where the C5 contract specifies `format: iso3166_alpha2` (2-letter code) |
| CTR-strictness-parity | Producer (SwiftHaul) uses country names internally but the contract boundary requires ISO alpha-2 codes -- no mapping function exists |

### Production Impact

Every international handoff from SwiftHaul to OceanBridge fails with a 400 validation error. No international shipments are booked. Packages pile up at the domestic hub waiting for handoff. Customers with international orders see indefinite "processing" status.

---

## M05: Enum Case Mismatch (SCREAMING_SNAKE vs snake_case)

**Boundary:** SwiftHaul -> ShopStream via PingWave (Contracts C4, C7)

**What's wrong:** SwiftHaul returns status values in SCREAMING_SNAKE_CASE (e.g., `"IN_TRANSIT"`), which is its documented convention. When PingWave relays the delivery status to ShopStream, ShopStream's webhook handler compares against lowercase snake_case values (`"in_transit"`). The comparison never matches, so every delivery status update is silently ignored.

**Why it's realistic:** Enum case conventions vary between services. SwiftHaul (a Go service modeled after FedEx/UPS) uses SCREAMING_SNAKE_CASE. ShopStream (TypeScript, modeled after Shopify) uses snake_case. When a third service (PingWave) passes status values through, neither side converts.

### Producer: SwiftHaul (Go) -- returns SCREAMING_SNAKE_CASE

```go
// swifthaul/pickups.go — Returns status in SCREAMING_SNAKE_CASE.

func HandleCreatePickup(w http.ResponseWriter, r *http.Request) {
	var req PickupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writePickupError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	if err := validatePickupRequest(&req); err != nil {
		writePickupError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	resp := PickupResponse{
		PickupID:          fmt.Sprintf("pku_%d", time.Now().UnixNano()),
		Status:            "IN_TRANSIT",  // SCREAMING_SNAKE_CASE -- SwiftHaul convention
		TrackingNumber:    req.TrackingNumber,
		EstimatedDelivery: time.Now().AddDate(0, 0, 7).UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
```

### Relay: PingWave (TypeScript) -- passes status through unchanged

```typescript
// pingwave-relay.ts — Relays SwiftHaul status to ShopStream callback.

async function relayDeliveryStatus(
  callbackUrl: string,
  orderId: string,
  swiftHaulStatus: string,  // "IN_TRANSIT" from SwiftHaul
  webhookSecret: string,
): Promise<void> {
  // BUG: Passes SwiftHaul's SCREAMING_SNAKE_CASE status directly to ShopStream
  // without converting to ShopStream's expected snake_case
  const payload: DeliveryCallbackPayload = {
    event_type: "delivery_confirmed",
    order_id: orderId,
    status: swiftHaulStatus as "in_transit" | "out_for_delivery" | "delivered" | "failed",
    // BUG: status is "IN_TRANSIT" but type assertion silently passes at runtime
    // CORRECT would be: status: swiftHaulStatus.toLowerCase() as ...
    timestamp: new Date().toISOString(),
    details: `Package status updated to ${swiftHaulStatus}`,
  };

  const body = JSON.stringify(payload);
  const timestamp = Math.floor(Date.now() / 1000).toString();
  const signedPayload = `${timestamp}.${body}`;
  const signature = crypto
    .createHmac("sha256", webhookSecret)
    .update(signedPayload)
    .digest("hex");

  const response = await fetch(callbackUrl, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "X-PingWave-Signature": signature,
      "X-PingWave-Timestamp": timestamp,
    },
    body,
  });

  if (!response.ok) {
    throw new Error(`Callback failed with status ${response.status}`);
  }
}
```

### Consumer: ShopStream (TypeScript) -- expects snake_case

```typescript
// shopstream-webhook-handler.ts — Handles delivery callbacks from PingWave.

interface DeliveryWebhookEvent {
  event_type: "delivery_confirmed" | "delivery_failed" | "notification_sent";
  order_id: string;
  status: "in_transit" | "out_for_delivery" | "delivered" | "failed";
  timestamp: string;
  details?: string;
}

function handleDeliveryWebhook(event: DeliveryWebhookEvent): void {
  // BUG: These comparisons use snake_case but SwiftHaul sends SCREAMING_SNAKE_CASE
  // "IN_TRANSIT" !== "in_transit" -- none of these cases ever match
  switch (event.status) {
    case "in_transit":
      // Never reached when status is "IN_TRANSIT"
      updateOrderStatus(event.order_id, "shipped");
      break;
    case "out_for_delivery":
      // Never reached when status is "OUT_FOR_DELIVERY"
      updateOrderStatus(event.order_id, "shipped");
      notifyCustomer(event.order_id, "Your package is out for delivery!");
      break;
    case "delivered":
      // Never reached when status is "DELIVERED"
      updateOrderStatus(event.order_id, "delivered");
      break;
    case "failed":
      // Never reached when status is "FAILED"
      updateOrderStatus(event.order_id, "cancelled");
      initiateRefund(event.order_id);
      break;
    default:
      // All SwiftHaul statuses fall through to here
      console.warn(`Unknown delivery status: ${event.status}`);
      // Status update silently dropped
  }
}

function updateOrderStatus(orderId: string, status: string): void {
  // Database update -- never called for SwiftHaul-originating events
}

function notifyCustomer(orderId: string, message: string): void {
  // Push notification -- never called
}

function initiateRefund(orderId: string): void {
  // Refund flow -- never called
}
```

### Expected Stricture Violations

| Rule | Reason |
|------|--------|
| CTR-strictness-parity | SwiftHaul uses SCREAMING_SNAKE_CASE enums (`IN_TRANSIT`) but the C7 contract and ShopStream consumer expect snake_case (`in_transit`) -- enum case convention mismatch across boundaries |

### Production Impact

Order status never updates past "shipped." Customers see stale tracking. Deliveries are completed but the order stays in "processing." Failed deliveries never trigger refunds. The system appears to work (no errors thrown) but all delivery status transitions are silently dropped.

---

## M06: String Decimal vs Integer Cents for Customs

**Boundary:** SwiftHaul -> OceanBridge (Contract C5)

**What's wrong:** OceanBridge expects `declared_value: "29.99"` (a string containing a decimal number, per customs requirements). SwiftHaul converts its internal integer cents value to a string, but uses `strconv.Itoa(2999)` which produces `"2999"` instead of `fmt.Sprintf("%.2f", float64(2999)/100.0)` which would produce `"29.99"`. OceanBridge parses `"2999"` as $2,999.00.

**Why it's realistic:** The declared value field has a unique contract: it's a string (not integer, not float) containing a decimal representation. This triple conversion (int cents -> float dollars -> string) is easy to get wrong, especially when `strconv.Itoa` seems like a natural way to "convert to string."

### Producer: SwiftHaul (Go)

```go
// swifthaul/handoff.go — Converts declared value for OceanBridge customs.

func HandoffToOceanBridge(
	oceanBridgeURL, oceanBridgeKey string,
	pickupID string,
	pkg PackageDetails,
	destCountry string,
	declaredValueCents int, // 2999 (i.e., $29.99)
	currency string,
	description string,
) error {
	weightKg := float64(pkg.WeightGrams) / 1000.0

	// BUG: Converts cents integer to string directly -- "2999" instead of "29.99"
	declaredValueStr := strconv.Itoa(declaredValueCents)
	// CORRECT would be: declaredValueStr := fmt.Sprintf("%.2f", float64(declaredValueCents)/100.0)

	currencyUpper := strings.ToUpper(currency)

	booking := OceanBridgeBookingRequest{
		OriginCountry:      "US",
		DestinationCountry: destCountry,
		Packages: []OceanBridgePackage{{
			WeightKg: weightKg,
			LengthCm: pkg.LengthCm,
			WidthCm:  pkg.WidthCm,
			HeightCm: pkg.HeightCm,
		}},
		DeclaredValue:       declaredValueStr, // BUG: "2999" not "29.99"
		Currency:            currencyUpper,
		ContentsDescription: description,
		CustomsInfo: CustomsInfo{
			SenderName:    "VaultStore Warehouse",
			RecipientName: "Customer",
		},
	}

	body, err := json.Marshal(booking)
	if err != nil {
		return fmt.Errorf("marshal booking request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, oceanBridgeURL+"/api/v1/bookings", nil)
	if err != nil {
		return fmt.Errorf("create booking request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+oceanBridgeKey)
	_ = body

	return nil
}
```

### Consumer: OceanBridge (Python)

```python
# oceanbridge/bookings.py — Parses declared_value as decimal string.

class BookingRequest:
    def __init__(self, data: dict[str, Any]) -> None:
        self.origin_country: str = data["origin_country"]
        self.destination_country: str = data["destination_country"]
        self.packages: list[dict[str, Any]] = data["packages"]
        self.declared_value: str = data["declared_value"]  # expects "29.99"
        self.currency: str = data["currency"]
        self.contents_description: str = data["contents_description"]
        self.hs_code: str | None = data.get("hs_code")
        self.customs_info: dict[str, str] = data["customs_info"]

    def validate(self) -> None:
        if not ISO_COUNTRY_PATTERN.match(self.origin_country):
            raise ValueError(f"Invalid origin_country: {self.origin_country}")
        if not ISO_COUNTRY_PATTERN.match(self.destination_country):
            raise ValueError(f"Invalid destination_country: {self.destination_country}")
        if self.currency not in VALID_CURRENCIES:
            raise ValueError(f"Invalid currency: {self.currency}")

        # Validates declared_value is a valid decimal string
        # "2999" is valid as a decimal -- Decimal("2999") succeeds
        # but it represents $2999.00, not $29.99
        try:
            val = Decimal(self.declared_value)
            if val <= 0:
                raise ValueError(f"declared_value must be positive, got {self.declared_value}")
        except Exception as e:
            raise ValueError(f"Invalid declared_value: {self.declared_value}") from e

        for pkg in self.packages:
            if pkg.get("weight_kg", 0) <= 0:
                raise ValueError(f"weight_kg must be positive: {pkg.get('weight_kg')}")


def generate_customs_declaration(req: BookingRequest) -> dict[str, Any]:
    """Generates customs paperwork using the declared value."""
    # Declared value "2999" is interpreted as $2,999.00
    # Customs duty is calculated on this inflated amount
    declared = Decimal(req.declared_value)
    duty_rate = Decimal("0.05")  # 5% duty
    duty_amount = declared * duty_rate  # $149.95 instead of $1.50

    return {
        "declared_value": str(declared),   # "2999"
        "currency": req.currency,
        "duty_amount": str(duty_amount),   # "149.9500" instead of "1.4950"
        "total_customs_charge": str(declared + duty_amount),
    }
```

### Expected Stricture Violations

| Rule | Reason |
|------|--------|
| CTR-request-shape | Field `declared_value` should be a decimal string like `"29.99"` per the C5 contract, but receives `"2999"` (integer-as-string without decimal conversion) |
| CTR-strictness-parity | SwiftHaul's internal representation is integer cents but the contract boundary requires a decimal dollar string -- the conversion uses `strconv.Itoa` (integer-to-string) instead of cents-to-dollars-to-string |

### Production Impact

Customs declarations show 100x the actual value. A $29.99 item is declared at $2,999.00. Import duties are calculated on the inflated amount (5% of $2,999 = $149.95 instead of $1.50). Shipments may be flagged for additional customs scrutiny. Customers are charged excessive duty fees on delivery.

---

## M07: Missing Webhook Signature Verification

**Boundary:** PingWave -> ShopStream (Contract C7)

**What's wrong:** PingWave sends delivery callbacks to ShopStream with `X-PingWave-Signature` and `X-PingWave-Timestamp` headers for HMAC-SHA256 verification. ShopStream's webhook handler skips signature verification entirely -- it parses the request body without checking the signature. Any attacker who knows the endpoint URL can forge delivery status updates.

**Why it's realistic:** Webhook signature verification is commonly skipped during development ("I'll add it later") and never backfilled. The handler works correctly for legitimate requests, so the missing verification is invisible in normal testing. It only matters when an attacker discovers the endpoint.

### Producer: PingWave (TypeScript) -- sends signed webhook

```typescript
// pingwave-client.ts — Sends signed delivery callbacks.

async function sendDeliveryCallback(
  callbackUrl: string,
  payload: DeliveryCallbackPayload,
  webhookSecret: string,
): Promise<void> {
  const body = JSON.stringify(payload);
  const timestamp = Math.floor(Date.now() / 1000).toString();
  const signedPayload = `${timestamp}.${body}`;

  // PingWave correctly signs the payload with HMAC-SHA256
  const signature = crypto
    .createHmac("sha256", webhookSecret)
    .update(signedPayload)
    .digest("hex");

  let response: Response;
  try {
    response = await fetch(callbackUrl, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-PingWave-Signature": signature,      // HMAC signature
        "X-PingWave-Timestamp": timestamp,       // Unix epoch seconds
      },
      body,
    });
  } catch (err) {
    throw new Error(`Network error sending callback: ${(err as Error).message}`);
  }

  if (!response.ok) {
    throw new Error(`Callback failed with status ${response.status}`);
  }
}
```

### Consumer: ShopStream (TypeScript)

```typescript
// shopstream-webhook-handler.ts — Handles delivery callbacks WITHOUT verification.

import type { Request, Response } from "express";

interface DeliveryWebhookEvent {
  event_type: "delivery_confirmed" | "delivery_failed" | "notification_sent";
  order_id: string;
  status: "in_transit" | "out_for_delivery" | "delivered" | "failed";
  timestamp: string;
  details?: string;
}

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

  let event: DeliveryWebhookEvent;
  try {
    event = req.body as DeliveryWebhookEvent;
  } catch {
    res.status(400).json({ error: { code: "invalid_payload", message: "Invalid JSON" } });
    return;
  }

  if (!event.order_id || !event.status) {
    res.status(400).json({
      error: { code: "invalid_payload", message: "Missing required fields" },
    });
    return;
  }

  // Processes the webhook without any authentication
  processDeliveryUpdate(event);

  res.status(200).json({ received: true });
}

function processDeliveryUpdate(event: DeliveryWebhookEvent): void {
  // Updates order status based on unverified data
  switch (event.status) {
    case "delivered":
      markOrderDelivered(event.order_id, event.timestamp);
      break;
    case "failed":
      initiateRefund(event.order_id);
      break;
    default:
      updateTrackingStatus(event.order_id, event.status);
  }
}

function markOrderDelivered(orderId: string, timestamp: string): void { /* ... */ }
function initiateRefund(orderId: string): void { /* ... */ }
function updateTrackingStatus(orderId: string, status: string): void { /* ... */ }
```

### Expected Stricture Violations

| Rule | Reason |
|------|--------|
| CTR-manifest-conformance | Contract C7 specifies `verification.method: hmac-sha256` and `verification.header: X-PingWave-Signature`, but the consumer handler does not read or verify these headers |
| CTR-request-shape | The contract requires `X-Webhook-Signature` header verification on inbound webhooks, but the handler accepts unsigned requests |

### Production Impact

The delivery webhook endpoint is unauthenticated. An attacker who discovers the URL can forge delivery status updates: mark orders as "delivered" to suppress refund eligibility, mark orders as "failed" to trigger unauthorized refunds, or inject fake tracking data. This is a privilege escalation vulnerability that allows external actors to manipulate order state.

---

## M08: Timezone-Naive Datetime

**Boundary:** OceanBridge -> PingWave (Contract C6)

**What's wrong:** OceanBridge generates `estimated_arrival` using `datetime.now().isoformat()` (no timezone -- produces `"2024-01-20T14:00:00"`). The contract requires ISO 8601 with UTC timezone. PingWave treats the naive timestamp as UTC, but the OceanBridge server runs in CET (Central European Time, UTC+01:00). The customer receives an ETA that is 1 hour off.

**Why it's realistic:** Python's `datetime.now()` returns a timezone-naive datetime by default. Developers must explicitly use `datetime.now(timezone.utc)` or `datetime.now(tz=timezone.utc)` to get a timezone-aware datetime. This is one of Python's most common datetime pitfalls, and it silently produces incorrect timestamps that look valid.

### Producer: OceanBridge (Python)

```python
# oceanbridge/bookings.py — Generates timezone-naive estimated_arrival.

from datetime import datetime  # BUG: timezone not imported

def create_booking(data: dict[str, Any]) -> dict[str, Any]:
    req = BookingRequest(data)
    req.validate()

    booking_id = f"obk_{int(datetime.now().timestamp() * 1000)}"

    # BUG: datetime.now() without timezone -- produces naive datetime
    # On a CET server, 14:00 local = 13:00 UTC, but output has no timezone info
    eta = datetime.now().replace(
        hour=14, minute=0, second=0, microsecond=0,
    )
    estimated_arrival = eta.isoformat()  # "2024-01-20T14:00:00" -- no Z, no +00:00
    # CORRECT would be:
    # from datetime import timezone
    # eta = datetime.now(timezone.utc).replace(hour=14, minute=0, second=0, microsecond=0)
    # estimated_arrival = eta.isoformat()  # "2024-01-20T14:00:00+00:00"

    resp = BookingResponse(
        booking_id=booking_id,
        status="pending",
        estimated_arrival=estimated_arrival,
        tracking_url=f"https://track.oceanbridge.io/{booking_id}",
    )

    return resp.to_dict()


def send_customs_notification(
    pingwave_url: str,
    pingwave_key: str,
    booking_id: str,
    event_type: str,
    recipient_email: str,
    estimated_arrival: str,  # "2024-01-20T14:00:00" -- no timezone
    origin_country: str,
    destination_country: str,
) -> dict[str, Any]:
    """Send customs clearance notification via PingWave."""

    payload = {
        "channel": "email",
        "recipient": recipient_email,
        "template_id": "tmpl_customs_update",
        "variables": {
            "booking_id": booking_id,
            "event_type": event_type,
            "origin": origin_country,
            "destination": destination_country,
            # BUG: timezone-naive timestamp passed through to customer notification
            "estimated_arrival": estimated_arrival,  # "2024-01-20T14:00:00"
        },
        "callback_url": f"{pingwave_url}/callbacks/customs/{booking_id}",
    }

    return payload
```

### Consumer: PingWave (TypeScript)

```typescript
// pingwave-notification.ts — Processes estimated arrival as UTC.

interface CustomsNotificationVariables {
  booking_id: string;
  event_type: string;
  origin: string;
  destination: string;
  estimated_arrival: string;  // expects ISO 8601 with timezone
}

function formatEstimatedArrival(
  variables: CustomsNotificationVariables,
  customerTimezone: string,
): string {
  // PingWave assumes the timestamp is UTC (per contract)
  // "2024-01-20T14:00:00" has no timezone -- JavaScript Date assumes UTC for
  // ISO-like strings without timezone, but behavior varies across engines
  const eta = new Date(variables.estimated_arrival);

  if (isNaN(eta.getTime())) {
    throw new Error(`Invalid estimated_arrival: ${variables.estimated_arrival}`);
  }

  // Converts "UTC" time to customer's local timezone for the email
  // But the source time was CET (14:00 CET = 13:00 UTC)
  // So the customer sees 14:00 UTC = 15:00 CET -- 1 hour late
  const formatted = eta.toLocaleString("en-US", {
    timeZone: customerTimezone,
    dateStyle: "medium",
    timeStyle: "short",
  });

  return formatted;
}

function buildCustomsEmail(
  recipientEmail: string,
  variables: CustomsNotificationVariables,
): { subject: string; body: string } {
  const arrivalDisplay = formatEstimatedArrival(variables, "Europe/Berlin");

  return {
    subject: `Customs Update: ${variables.event_type} for ${variables.booking_id}`,
    body: `Your shipment from ${variables.origin} to ${variables.destination} ` +
      `has an estimated arrival of ${arrivalDisplay}. ` +
      `Booking reference: ${variables.booking_id}`,
  };
}
```

### Expected Stricture Violations

| Rule | Reason |
|------|--------|
| CTR-strictness-parity | OceanBridge produces timezone-naive datetimes (`datetime.now()`) but the C6 contract requires ISO 8601 with UTC timezone -- the `estimated_arrival` field crosses the boundary without timezone information |

### Production Impact

Estimated arrival times are silently off by the OceanBridge server's UTC offset. For a CET server (UTC+1), customers see ETAs 1 hour later than intended. During daylight saving time transitions, the offset changes to 2 hours. The bug is subtle because the timestamp format looks valid -- only the timezone is missing. Customers plan around incorrect ETAs, leading to missed deliveries and support escalations.

---

## M09: Date Format Degradation Across 3 Hops

**Boundary:** PayCore -> ShopStream -> VaultStore (Contracts C8, C2)

**What's wrong:** PayCore sends `created_at: "2024-01-15T10:30:00Z"` (ISO 8601 with UTC timezone). ShopStream parses it with `new Date()` but then re-serializes it using `.toLocaleString()`, producing `"1/15/2024, 10:30:00 AM"`. When ShopStream forwards this to VaultStore, VaultStore's Go code tries to parse it with `time.Parse(time.RFC3339, ...)` and fails because the format is locale-dependent, not RFC 3339.

**Why it's realistic:** Date format degradation is a systemic problem in multi-hop architectures. Each hop parses and re-serializes timestamps in its own preferred format. The first hop (PayCore -> ShopStream) works fine because `new Date()` accepts ISO 8601. The second hop (ShopStream -> VaultStore) breaks because ShopStream's serialization uses a locale-dependent format that downstream services cannot parse.

### Hop 1 -- Producer: PayCore (Go) -- sends correct ISO 8601

```go
// paycore/webhooks.go — Sends payment webhook with ISO 8601 timestamp.

func SendWebhook(targetURL, secret string, charge *ChargeResponse, eventType string) error {
	payload := WebhookPayload{
		EventID:   fmt.Sprintf("evt_%s", generateID()),
		EventType: eventType,
		Data: WebhookData{
			ID:       charge.ID,
			Amount:   charge.Amount,
			Currency: charge.Currency,
			Status:   charge.Status,
		},
		CreatedAt: time.Now().UTC().Format(time.RFC3339), // "2024-01-15T10:30:00Z" -- correct
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal webhook payload: %w", err)
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signedPayload := timestamp + "." + string(body)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signedPayload))
	signature := hex.EncodeToString(mac.Sum(nil))

	req, err := http.NewRequest(http.MethodPost, targetURL, nil)
	if err != nil {
		return fmt.Errorf("create webhook request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-PayCore-Signature", signature)
	req.Header.Set("X-PayCore-Timestamp", timestamp)

	_ = body
	return nil
}
```

### Hop 2 -- Relay: ShopStream (TypeScript) -- degrades the format

```typescript
// shopstream-payment-handler.ts — Receives PayCore webhook, forwards to VaultStore.

interface PayCoreWebhookEvent {
  event_id: string;
  event_type: "charge.succeeded" | "charge.failed" | "refund.succeeded";
  data: {
    id: string;
    amount: number;
    currency: string;
    status: string;
  };
  created_at: string;  // "2024-01-15T10:30:00Z" from PayCore
}

async function handlePaymentWebhook(event: PayCoreWebhookEvent): Promise<void> {
  if (event.event_type === "charge.succeeded") {
    // Parse the ISO 8601 timestamp -- this works fine
    const createdAt = new Date(event.created_at);

    // BUG: Re-serialize using toLocaleString() -- produces locale-dependent format
    // "2024-01-15T10:30:00Z" becomes "1/15/2024, 10:30:00 AM"
    const formattedDate = createdAt.toLocaleString();
    // CORRECT would be: const formattedDate = createdAt.toISOString();

    // Forward payment confirmation to VaultStore for inventory reconciliation
    await notifyVaultStore({
      order_id: event.data.id,
      payment_status: event.data.status,
      amount: event.data.amount,
      currency: event.data.currency,
      // BUG: Locale string instead of ISO 8601
      created_at: formattedDate,  // "1/15/2024, 10:30:00 AM"
    });
  }
}

async function notifyVaultStore(data: {
  order_id: string;
  payment_status: string;
  amount: number;
  currency: string;
  created_at: string;
}): Promise<void> {
  let response: Response;
  try {
    response = await fetch(`${vaultstoreUrl}/api/inventory/payment-confirmed`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${vaultstoreKey}`,
      },
      body: JSON.stringify(data),
    });
  } catch (err) {
    throw new Error(`Network error calling VaultStore: ${(err as Error).message}`);
  }

  if (!response.ok) {
    const errorBody = await response.json() as { error: { code: string; message: string } };
    throw new Error(`VaultStore error: ${errorBody.error.message}`);
  }
}

const vaultstoreUrl = process.env["VAULTSTORE_URL"] ?? "";
const vaultstoreKey = process.env["VAULTSTORE_KEY"] ?? "";
```

### Hop 3 -- Consumer: VaultStore (Go) -- fails to parse

```go
// vaultstore/payment_confirmation.go — Receives payment confirmation from ShopStream.

type PaymentConfirmation struct {
	OrderID       string `json:"order_id"`
	PaymentStatus string `json:"payment_status"`
	Amount        int    `json:"amount"`
	Currency      string `json:"currency"`
	CreatedAt     string `json:"created_at"` // expects ISO 8601 / RFC 3339
}

func HandlePaymentConfirmation(w http.ResponseWriter, r *http.Request) {
	var conf PaymentConfirmation
	if err := json.NewDecoder(r.Body).Decode(&conf); err != nil {
		writeShipmentError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	// BUG: Tries to parse "1/15/2024, 10:30:00 AM" as RFC 3339
	// time.Parse fails because the format is locale-dependent
	parsedTime, err := time.Parse(time.RFC3339, conf.CreatedAt)
	if err != nil {
		// This error fires for every payment confirmation from ShopStream
		// Error: parsing time "1/15/2024, 10:30:00 AM" as "2006-01-02T15:04:05Z07:00": cannot parse
		writeShipmentError(w, http.StatusBadRequest, "invalid_timestamp",
			fmt.Sprintf("created_at must be RFC 3339: %s", err.Error()))
		return
	}

	// Never reached -- the timestamp parse always fails
	_ = parsedTime
	recordPaymentConfirmation(conf.OrderID, conf.PaymentStatus, conf.Amount)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"received": true})
}

func writeShipmentError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{"code": code, "message": message},
	})
}

func recordPaymentConfirmation(orderID, status string, amount int) {
	// Inventory release logic -- never called due to timestamp parse failure
}
```

### Expected Stricture Violations

| Rule | Reason |
|------|--------|
| CTR-strictness-parity | ShopStream receives ISO 8601 from PayCore but re-serializes with `.toLocaleString()`, degrading the format before forwarding to VaultStore -- timestamp format not preserved across the relay hop |
| CTR-response-shape | ShopStream's outbound payload uses a locale-dependent date string where both the C8 (inbound) and downstream contracts require ISO 8601 |

### Production Impact

Every payment confirmation from ShopStream to VaultStore fails with a 400 timestamp parsing error. VaultStore never receives payment confirmations, so inventory reserved for paid orders is never released for fulfillment. Orders stay in "paid" status indefinitely, shipments are never prepared, and customers never receive their purchases despite successful payment.

---

## M10: Implicit Callback Contract (Undocumented Endpoint)

**Boundary:** PingWave -> ShopStream (Contract C7)

**What's wrong:** PingWave sends delivery status callbacks to ShopStream, but the two services disagree on the field names. PingWave sends `event_type` and `order_id` (per the PingWave API spec). ShopStream's callback handler expects `type` and `orderId` (camelCase, per an older internal convention). Neither side has the other's contract in their manifest -- the callback URL was shared informally.

**Why it's realistic:** Callback/webhook endpoints are the most common source of implicit contracts. The producer (PingWave) documents what it sends. The consumer (ShopStream) documents what it expects. But nobody verifies that these match. The field name divergence happens because each team names fields according to their own naming convention.

### Producer: PingWave (TypeScript)

```typescript
// pingwave-callback.ts — Sends delivery callback to ShopStream.

interface DeliveryCallbackPayload {
  event_type: "delivery_confirmed" | "delivery_failed" | "notification_sent";
  order_id: string;       // snake_case -- PingWave convention
  status: "in_transit" | "out_for_delivery" | "delivered" | "failed";
  timestamp: string;
  details?: string;
}

async function sendDeliveryCallback(
  callbackUrl: string,
  orderId: string,
  status: string,
  webhookSecret: string,
): Promise<void> {
  // PingWave sends snake_case field names per its API spec
  const payload: DeliveryCallbackPayload = {
    event_type: "delivery_confirmed",   // PingWave calls it "event_type"
    order_id: orderId,                  // PingWave calls it "order_id"
    status: status as DeliveryCallbackPayload["status"],
    timestamp: new Date().toISOString(),
    details: "Package delivered successfully",
  };

  const body = JSON.stringify(payload);
  const ts = Math.floor(Date.now() / 1000).toString();
  const signedPayload = `${ts}.${body}`;
  const signature = crypto
    .createHmac("sha256", webhookSecret)
    .update(signedPayload)
    .digest("hex");

  let response: Response;
  try {
    response = await fetch(callbackUrl, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-PingWave-Signature": signature,
        "X-PingWave-Timestamp": ts,
      },
      body,
    });
  } catch (err) {
    throw new Error(`Network error sending callback: ${(err as Error).message}`);
  }

  if (!response.ok) {
    throw new Error(`Callback failed with status ${response.status}`);
  }
}
```

### Consumer: ShopStream (TypeScript)

```typescript
// shopstream-callback-handler.ts — Expects different field names than PingWave sends.

import type { Request, Response } from "express";
import * as crypto from "crypto";

// BUG: ShopStream expects camelCase field names and "type" instead of "event_type"
interface ExpectedCallbackPayload {
  type: string;           // BUG: PingWave sends "event_type", not "type"
  orderId: string;        // BUG: PingWave sends "order_id", not "orderId"
  status: string;
  timestamp: string;
  details?: string;
}

// CORRECT interface would match PingWave's contract:
// interface ExpectedCallbackPayload {
//   event_type: string;
//   order_id: string;
//   status: string;
//   timestamp: string;
//   details?: string;
// }

function handleDeliveryCallback(req: Request, res: Response): void {
  const signature = req.headers["x-pingwave-signature"] as string | undefined;
  const timestamp = req.headers["x-pingwave-timestamp"] as string | undefined;

  if (!signature || !timestamp) {
    res.status(401).json({ error: { code: "invalid_signature", message: "Missing signature" } });
    return;
  }

  // Signature verification works correctly
  const webhookSecret = process.env["PINGWAVE_WEBHOOK_SECRET"] ?? "";
  const rawBody = JSON.stringify(req.body);
  const signedPayload = `${timestamp}.${rawBody}`;
  const expected = crypto
    .createHmac("sha256", webhookSecret)
    .update(signedPayload)
    .digest("hex");

  if (signature !== expected) {
    res.status(401).json({ error: { code: "invalid_signature", message: "Signature mismatch" } });
    return;
  }

  // Parse body with wrong field names
  const event = req.body as ExpectedCallbackPayload;

  // BUG: event.type is undefined (PingWave sent event_type)
  if (!event.type) {
    // This fires for every callback -- "type" is always undefined
    console.error("Callback missing event type:", JSON.stringify(req.body));
    res.status(400).json({
      error: { code: "invalid_payload", message: "Missing required field: type" },
    });
    return;
  }

  // BUG: event.orderId is undefined (PingWave sent order_id)
  if (!event.orderId) {
    // If the type check above were removed, this would fire instead
    res.status(400).json({
      error: { code: "invalid_payload", message: "Missing required field: orderId" },
    });
    return;
  }

  // Never reached due to field name mismatches above
  processDeliveryUpdate(event.type, event.orderId, event.status);
  res.status(200).json({ received: true });
}

function processDeliveryUpdate(type: string, orderId: string, status: string): void {
  // Never called -- callback always rejected at validation
}
```

### Expected Stricture Violations

| Rule | Reason |
|------|--------|
| CTR-manifest-conformance | The callback contract is not explicitly defined in either service's manifest -- PingWave's outbound webhook schema and ShopStream's inbound handler use different field names (`event_type` vs `type`, `order_id` vs `orderId`) with no shared contract to reconcile them |

### Production Impact

Every delivery callback from PingWave to ShopStream returns 400. ShopStream never receives delivery status updates. Customers see orders stuck in "shipped" status even after delivery. PingWave retries the callback (up to the retry limit), generating unnecessary load. The debugging is especially difficult because both sides have valid code -- the bug only exists in the implicit contract between them.

---

## Detection Summary Table

| Mismatch | Rules Triggered | Severity | Auto-fixable? |
|----------|----------------|----------|---------------|
| M01: Float dollars vs integer cents | CTR-strictness-parity, CTR-response-shape | Critical | Yes -- insert `/ 100` before rendering |
| M02: Weight grams vs kilograms | CTR-request-shape, CTR-strictness-parity | Critical | Yes -- insert `/ 1000.0` conversion |
| M03: Dimensions mm sent as cm | CTR-strictness-parity | High | Yes -- insert `/ 10` conversion |
| M04: Country name vs ISO code | CTR-request-shape, CTR-strictness-parity | High | No -- requires country name-to-code mapping table |
| M05: Enum case mismatch | CTR-strictness-parity | High | Yes -- insert `.toLowerCase()` or case-insensitive comparison |
| M06: String decimal vs integer cents | CTR-request-shape, CTR-strictness-parity | Critical | Yes -- replace `strconv.Itoa` with `fmt.Sprintf("%.2f", cents/100.0)` |
| M07: Missing webhook signature | CTR-manifest-conformance, CTR-request-shape | Critical | No -- requires implementing HMAC verification logic |
| M08: Timezone-naive datetime | CTR-strictness-parity | Medium | Yes -- replace `datetime.now()` with `datetime.now(timezone.utc)` |
| M09: Date format degradation | CTR-strictness-parity, CTR-response-shape | High | Yes -- replace `.toLocaleString()` with `.toISOString()` |
| M10: Implicit callback contract | CTR-manifest-conformance | High | No -- requires contract negotiation between teams to agree on field names |
