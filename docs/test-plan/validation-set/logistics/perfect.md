# PERFECT Integration

All 7 services with correct cross-boundary type matching, consistent units, and proper error handling.

> **Navigation:** [Back to Overview](README.md) | [Companies](companies.md) | [Contracts](contracts.md) | [Mismatches](mismatches.md)

### 4.1 ShopStream — Order Orchestrator (TypeScript)

```typescript
// shopstream-order-service.ts — E-commerce order orchestration.
// Calls PayCore, VaultStore, and receives webhooks from PayCore + PingWave.

import * as crypto from "crypto";

// ── Types ──────────────────────────────────────────────────

interface OrderItem {
  sku: string;
  quantity: number;    // integer
  unit_price: number;  // integer, cents
}

interface ShippingAddress {
  line1: string;
  line2?: string;
  city: string;
  state: string;
  postal_code: string;
  country: string;  // ISO 3166-1 alpha-2
}

interface CreateOrderParams {
  customer_id: string;
  items: OrderItem[];
  currency: "usd" | "eur" | "gbp" | "cad" | "aud";
  shipping_address: ShippingAddress;
  payment_token: string;
  callback_url?: string;
  idempotency_key: string;
}

interface Order {
  id: string;
  customer_id: string;
  status: "pending" | "paid" | "processing" | "shipped" | "delivered" | "cancelled" | "refunded";
  total_amount: number;  // integer, cents
  currency: "usd" | "eur" | "gbp" | "cad" | "aud";
  payment_id: string | null;
  reservation_id: string | null;
  tracking_number: string | null;
  estimated_delivery: string | null;
  created_at: string;
  updated_at: string;
}

// PayCore response types
interface PayCoreCharge {
  id: string;           // pay_*
  amount: number;       // integer, cents
  currency: string;
  status: "succeeded" | "pending" | "failed";
  created_at: string;   // ISO 8601 UTC
  metadata: Record<string, string>;
}

// VaultStore response types
interface VaultStoreReservation {
  reservation_id: string;  // res_*
  status: "reserved" | "partial" | "unavailable";
  items: Array<{
    sku: string;
    quantity_reserved: number;
    quantity_available: number;
  }>;
  expires_at: string;  // ISO 8601 UTC
}

// Webhook types
interface PayCoreWebhookEvent {
  event_id: string;
  event_type: "charge.succeeded" | "charge.failed" | "refund.succeeded";
  data: {
    id: string;
    amount: number;     // integer, cents
    currency: string;
    status: string;
  };
  created_at: string;
}

interface DeliveryWebhookEvent {
  event_type: "delivery_confirmed" | "delivery_failed" | "notification_sent";
  order_id: string;
  status: "in_transit" | "out_for_delivery" | "delivered" | "failed";
  timestamp: string;
  details?: string;
}

type ServiceResult<T> =
  | { ok: true; data: T }
  | { ok: false; status: number; error: { code: string; message: string } };

// ── Validation ─────────────────────────────────────────────

const ISO_COUNTRY_PATTERN = /^[A-Z]{2}$/;

function validateCountryCode(code: string): void {
  if (!ISO_COUNTRY_PATTERN.test(code)) {
    throw new Error(`Invalid country code: ${code} (expected ISO 3166-1 alpha-2)`);
  }
}

function validateAmountCents(amount: number): void {
  if (!Number.isInteger(amount) || amount < 1) {
    throw new Error(`Amount must be a positive integer in cents, got ${amount}`);
  }
}

// ── ShopStream Client ──────────────────────────────────────

class ShopStreamOrderService {
  private readonly paycoreUrl: string;
  private readonly paycoreKey: string;
  private readonly vaultstoreUrl: string;
  private readonly vaultstoreKey: string;
  private readonly webhookSecret: string;

  constructor(config: {
    paycoreUrl: string;
    paycoreKey: string;
    vaultstoreUrl: string;
    vaultstoreKey: string;
    webhookSecret: string;
  }) {
    this.paycoreUrl = config.paycoreUrl;
    this.paycoreKey = config.paycoreKey;
    this.vaultstoreUrl = config.vaultstoreUrl;
    this.vaultstoreKey = config.vaultstoreKey;
    this.webhookSecret = config.webhookSecret;
  }

  // ── Create Order (orchestrates PayCore + VaultStore) ────

  async createOrder(params: CreateOrderParams): Promise<ServiceResult<Order>> {
    validateCountryCode(params.shipping_address.country);
    for (const item of params.items) {
      validateAmountCents(item.unit_price);
      if (!Number.isInteger(item.quantity) || item.quantity < 1 || item.quantity > 9999) {
        throw new Error(`Invalid quantity: ${item.quantity}`);
      }
    }

    const totalAmount = params.items.reduce(
      (sum, item) => sum + item.unit_price * item.quantity,
      0,
    );
    validateAmountCents(totalAmount);

    // Step 1: Charge payment via PayCore
    const chargeResult = await this.createPayCoreCharge({
      amount: totalAmount,           // integer cents
      currency: params.currency,     // lowercase
      source_token: params.payment_token,
      description: `Order for customer ${params.customer_id}`,
      metadata: { customer_id: params.customer_id },
      idempotency_key: params.idempotency_key,
    });

    if (!chargeResult.ok) {
      return {
        ok: false,
        status: chargeResult.status,
        error: { code: "payment_failed", message: `Payment failed: ${chargeResult.error.message}` },
      };
    }

    // Verify amount matches (integer cents comparison)
    if (chargeResult.data.amount !== totalAmount) {
      throw new Error(
        `Amount mismatch: expected ${totalAmount} cents, got ${chargeResult.data.amount} cents`,
      );
    }

    // Step 2: Reserve inventory via VaultStore
    const reserveResult = await this.reserveInventory({
      order_id: `ord_${crypto.randomUUID().replace(/-/g, "").slice(0, 16)}`,
      items: params.items.map((i) => ({ sku: i.sku, quantity: i.quantity })),
    });

    if (!reserveResult.ok) {
      // Refund payment if inventory reservation fails
      // (omitted for brevity -- would call PayCore /v1/refunds)
      return {
        ok: false,
        status: reserveResult.status,
        error: { code: "inventory_failed", message: reserveResult.error.message },
      };
    }

    if (reserveResult.data.status !== "reserved") {
      return {
        ok: false,
        status: 409,
        error: {
          code: "insufficient_stock",
          message: `Inventory status: ${reserveResult.data.status}`,
        },
      };
    }

    const now = new Date().toISOString();  // ISO 8601 UTC
    const order: Order = {
      id: reserveResult.data.reservation_id.replace("res_", "ord_"),
      customer_id: params.customer_id,
      status: "paid",
      total_amount: totalAmount,
      currency: params.currency,
      payment_id: chargeResult.data.id,
      reservation_id: reserveResult.data.reservation_id,
      tracking_number: null,
      estimated_delivery: null,
      created_at: now,
      updated_at: now,
    };

    return { ok: true, data: order };
  }

  // ── PayCore Integration ────────────────────────────────

  private async createPayCoreCharge(params: {
    amount: number;
    currency: string;
    source_token: string;
    description: string;
    metadata: Record<string, string>;
    idempotency_key: string;
  }): Promise<ServiceResult<PayCoreCharge>> {
    let response: Response;
    try {
      response = await fetch(`${this.paycoreUrl}/v1/charges`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${this.paycoreKey}`,
          "Idempotency-Key": params.idempotency_key,
        },
        body: JSON.stringify({
          amount: params.amount,           // integer cents -- matches PayCore contract
          currency: params.currency,       // lowercase -- matches PayCore contract
          source_token: params.source_token,
          description: params.description,
          metadata: params.metadata,
        }),
      });
    } catch (err) {
      throw new Error(`Network error calling PayCore: ${(err as Error).message}`);
    }

    if (!response.ok) {
      const errorBody = await response.json() as { error: { type: string; message: string } };
      return {
        ok: false,
        status: response.status,
        error: { code: errorBody.error.type, message: errorBody.error.message },
      };
    }

    const charge = (await response.json()) as PayCoreCharge;
    return { ok: true, data: charge };
  }

  // ── VaultStore Integration ─────────────────────────────

  private async reserveInventory(params: {
    order_id: string;
    items: Array<{ sku: string; quantity: number }>;
  }): Promise<ServiceResult<VaultStoreReservation>> {
    let response: Response;
    try {
      response = await fetch(`${this.vaultstoreUrl}/api/inventory/reserve`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${this.vaultstoreKey}`,
        },
        body: JSON.stringify({
          order_id: params.order_id,
          items: params.items,
        }),
      });
    } catch (err) {
      throw new Error(`Network error calling VaultStore: ${(err as Error).message}`);
    }

    if (!response.ok) {
      const errorBody = await response.json() as { error: { code: string; message: string } };
      return {
        ok: false,
        status: response.status,
        error: { code: errorBody.error.code, message: errorBody.error.message },
      };
    }

    const reservation = (await response.json()) as VaultStoreReservation;
    return { ok: true, data: reservation };
  }

  // ── Webhook Handlers ───────────────────────────────────

  verifyWebhookSignature(
    payload: string,
    signature: string,
    timestamp: string,
    secret: string,
  ): boolean {
    const ts = parseInt(timestamp, 10);
    if (isNaN(ts)) return false;

    const now = Math.floor(Date.now() / 1000);
    if (Math.abs(now - ts) > 300) return false;

    const signedPayload = `${ts}.${payload}`;
    const expected = crypto
      .createHmac("sha256", secret)
      .update(signedPayload)
      .digest("hex");

    const sigBuf = Buffer.from(signature, "hex");
    const expBuf = Buffer.from(expected, "hex");
    if (sigBuf.length !== expBuf.length) return false;

    return crypto.timingSafeEqual(sigBuf, expBuf);
  }

  handlePaymentWebhook(event: PayCoreWebhookEvent): void {
    // Handles all 3 event types from PayCore
    switch (event.event_type) {
      case "charge.succeeded":
        // amount is integer cents -- no conversion needed
        break;
      case "charge.failed":
        // Cancel order, notify customer
        break;
      case "refund.succeeded":
        // Update order status to "refunded"
        break;
      default: {
        const _exhaustive: never = event.event_type;
        throw new Error(`Unknown event type: ${_exhaustive}`);
      }
    }
  }

  handleDeliveryWebhook(event: DeliveryWebhookEvent): void {
    // Handles all 4 status values from PingWave
    switch (event.status) {
      case "in_transit":
      case "out_for_delivery":
      case "delivered":
      case "failed":
        // Update order status accordingly
        // event.timestamp is ISO 8601 UTC -- parse directly
        const ts = new Date(event.timestamp);
        if (isNaN(ts.getTime())) {
          throw new Error(`Invalid timestamp: ${event.timestamp}`);
        }
        break;
      default: {
        const _exhaustive: never = event.status;
        throw new Error(`Unknown delivery status: ${_exhaustive}`);
      }
    }
  }
}

export { ShopStreamOrderService };
export type {
  Order, CreateOrderParams, OrderItem, ShippingAddress,
  PayCoreCharge, VaultStoreReservation,
  PayCoreWebhookEvent, DeliveryWebhookEvent,
};
```

### 4.2 PayCore — Payment Service (Go)

```go
// paycore/charges.go — Payment charge handling.

package paycore

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// ── Types ──────────────────────────────────────────────────

type ChargeRequest struct {
	Amount      int               `json:"amount"`       // integer, cents
	Currency    string            `json:"currency"`     // lowercase enum
	SourceToken string            `json:"source_token"` // tok_*
	Description string            `json:"description,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type ChargeResponse struct {
	ID        string            `json:"id"`         // pay_*
	Amount    int               `json:"amount"`     // integer, cents
	Currency  string            `json:"currency"`   // lowercase
	Status    string            `json:"status"`     // "succeeded" | "pending" | "failed"
	CreatedAt string            `json:"created_at"` // ISO 8601 UTC
	Metadata  map[string]string `json:"metadata"`
}

type WebhookPayload struct {
	EventID   string        `json:"event_id"`   // evt_*
	EventType string        `json:"event_type"` // "charge.succeeded" etc
	Data      WebhookData   `json:"data"`
	CreatedAt string        `json:"created_at"` // ISO 8601 UTC
}

type WebhookData struct {
	ID       string `json:"id"`
	Amount   int    `json:"amount"`   // integer, cents
	Currency string `json:"currency"` // lowercase
	Status   string `json:"status"`
}

// ── Validation ─────────────────────────────────────────────

var validCurrencies = map[string]bool{
	"usd": true, "eur": true, "gbp": true, "cad": true, "aud": true,
}

func validateChargeRequest(req *ChargeRequest) error {
	if req.Amount < 50 || req.Amount > 99999999 {
		return fmt.Errorf("amount must be between 50 and 99999999 cents, got %d", req.Amount)
	}
	if !validCurrencies[req.Currency] {
		return fmt.Errorf("invalid currency: %s", req.Currency)
	}
	if len(req.SourceToken) < 4 || req.SourceToken[:4] != "tok_" {
		return fmt.Errorf("invalid source_token format: %s (expected tok_*)", req.SourceToken)
	}
	return nil
}

// ── Charge Handler ─────────────────────────────────────────

func HandleCreateCharge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "invalid_request", "Method not allowed")
		return
	}

	var req ChargeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	if err := validateChargeRequest(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	// Process charge (simplified)
	chargeID := fmt.Sprintf("pay_%s", generateID())
	now := time.Now().UTC().Format(time.RFC3339) // ISO 8601 UTC

	resp := ChargeResponse{
		ID:        chargeID,
		Amount:    req.Amount,    // integer cents -- same as request
		Currency:  req.Currency,  // lowercase -- same as request
		Status:    "succeeded",
		CreatedAt: now,
		Metadata:  req.Metadata,
	}

	if resp.Metadata == nil {
		resp.Metadata = make(map[string]string)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// ── Webhook Sender ─────────────────────────────────────────

func SendWebhook(targetURL, secret string, charge *ChargeResponse, eventType string) error {
	payload := WebhookPayload{
		EventID:   fmt.Sprintf("evt_%s", generateID()),
		EventType: eventType,
		Data: WebhookData{
			ID:       charge.ID,
			Amount:   charge.Amount,   // integer cents
			Currency: charge.Currency, // lowercase
			Status:   charge.Status,
		},
		CreatedAt: time.Now().UTC().Format(time.RFC3339), // ISO 8601 UTC
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

	// Would send req with body here
	_ = body
	return nil
}

// ── Helpers ────────────────────────────────────────────────

func writeError(w http.ResponseWriter, status int, errType, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"type":    errType,
			"message": message,
		},
	})
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
```

### 4.3 VaultStore — Warehouse Management (Go)

```go
// vaultstore/shipments.go — Shipment preparation with LabelForge integration.

package vaultstore

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ── Types ──────────────────────────────────────────────────

type Dimensions struct {
	Length int `json:"length"` // millimeters
	Width  int `json:"width"`  // millimeters
	Height int `json:"height"` // millimeters
}

type Package struct {
	Weight     int        `json:"weight"`     // grams
	Dimensions Dimensions `json:"dimensions"` // millimeters
}

type ShipmentRequest struct {
	ReservationID string   `json:"reservation_id"` // res_*
	ShipTo        Address  `json:"ship_to"`
	Packages      []Package `json:"packages"`
}

type Address struct {
	Name       string `json:"name"`
	Line1      string `json:"line1"`
	Line2      string `json:"line2,omitempty"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"` // ISO 3166-1 alpha-2
}

// LabelForge request -- note the unit conversions
type LabelForgeRequest struct {
	FromAddress LabelForgeAddress `json:"from_address"`
	ToAddress   LabelForgeAddress `json:"to_address"`
	Parcel      LabelForgeParcel  `json:"parcel"`
	ServiceLevel string           `json:"service_level"`
}

type LabelForgeAddress struct {
	Name    string `json:"name"`
	Street1 string `json:"street1"`
	City    string `json:"city"`
	State   string `json:"state"`
	Zip     string `json:"zip"`
	Country string `json:"country"` // ISO 3166-1 alpha-2
}

type LabelForgeParcel struct {
	Weight int `json:"weight"` // grams (same unit)
	Length int `json:"length"` // centimeters (converted from mm)
	Width  int `json:"width"`  // centimeters (converted from mm)
	Height int `json:"height"` // centimeters (converted from mm)
}

type LabelForgeResponse struct {
	LabelID           string `json:"label_id"`           // lbl_*
	TrackingNumber    string `json:"tracking_number"`
	Carrier           string `json:"carrier"`
	ServiceLevel      string `json:"service_level"`
	Rate              int    `json:"rate"`               // cents
	LabelURL          string `json:"label_url"`
	EstimatedDelivery string `json:"estimated_delivery"` // ISO 8601 UTC
	CreatedAt         string `json:"created_at"`         // ISO 8601 UTC
}

// ── Shipment Handler ───────────────────────────────────────

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

	// CORRECT: Convert mm to cm by integer division
	labelReq := LabelForgeRequest{
		FromAddress: warehouseAddr,
		ToAddress: LabelForgeAddress{
			Name:    req.ShipTo.Name,
			Street1: req.ShipTo.Line1,
			City:    req.ShipTo.City,
			State:   req.ShipTo.State,
			Zip:     req.ShipTo.PostalCode,
			Country: req.ShipTo.Country, // ISO alpha-2, same convention
		},
		Parcel: LabelForgeParcel{
			Weight: pkg.Weight,                // grams -> grams (no conversion)
			Length: pkg.Dimensions.Length / 10, // mm -> cm
			Width:  pkg.Dimensions.Width / 10,  // mm -> cm
			Height: pkg.Dimensions.Height / 10, // mm -> cm
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
	_ = body // Would set body here

	// Simplified -- in production would send request and parse response
	resp := &LabelForgeResponse{
		LabelID:           fmt.Sprintf("lbl_%s", generateID()),
		TrackingNumber:    "1Z999AA10123456784",
		Carrier:           "ups",
		ServiceLevel:      "ground",
		Rate:              1295,
		LabelURL:          "https://labels.labelforge.com/lbl_123.pdf",
		EstimatedDelivery: time.Now().AddDate(0, 0, 5).UTC().Format(time.RFC3339),
		CreatedAt:         time.Now().UTC().Format(time.RFC3339),
	}

	return resp, nil
}
```

### 4.4 LabelForge — Label Generation (TypeScript)

```typescript
// labelforge-client.ts — Shipping label service.

// ── Types ──────────────────────────────────────────────────

interface LabelRequest {
  from_address: LabelAddress;
  to_address: LabelAddress;
  parcel: {
    weight: number;   // grams (integer)
    length: number;   // centimeters
    width: number;    // centimeters
    height: number;   // centimeters
  };
  service_level: "ground" | "express" | "overnight" | "international";
}

interface LabelAddress {
  name: string;
  street1: string;
  city: string;
  state: string;
  zip: string;
  country: string;  // ISO 3166-1 alpha-2
}

interface LabelResponse {
  label_id: string;        // lbl_*
  tracking_number: string; // carrier-specific
  carrier: "ups" | "fedex" | "usps" | "dhl";
  service_level: string;
  rate: number;            // integer, cents
  label_url: string;
  estimated_delivery: string;  // ISO 8601 UTC
  created_at: string;          // ISO 8601 UTC
}

type LabelResult =
  | { ok: true; data: LabelResponse }
  | { ok: false; status: number; error: { code: string; message: string } };

// ── Validation ─────────────────────────────────────────────

function validateParcel(parcel: LabelRequest["parcel"]): void {
  if (!Number.isInteger(parcel.weight) || parcel.weight < 1) {
    throw new Error(`Weight must be a positive integer in grams, got ${parcel.weight}`);
  }
  if (parcel.length < 1 || parcel.width < 1 || parcel.height < 1) {
    throw new Error("Dimensions must be positive numbers in centimeters");
  }
}

// ── Client ─────────────────────────────────────────────────

class LabelForgeClient {
  private readonly baseUrl: string;
  private readonly apiKey: string;

  constructor(baseUrl: string, apiKey: string) {
    this.baseUrl = baseUrl;
    this.apiKey = apiKey;
  }

  async createLabel(req: LabelRequest): Promise<LabelResult> {
    validateParcel(req.parcel);

    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/v2/labels`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${this.apiKey}`,
        },
        body: JSON.stringify(req),
      });
    } catch (err) {
      throw new Error(`Network error calling LabelForge: ${(err as Error).message}`);
    }

    if (!response.ok) {
      const errorBody = await response.json() as { error: { code: string; message: string } };
      return { ok: false, status: response.status, error: errorBody.error };
    }

    const label = (await response.json()) as LabelResponse;
    return { ok: true, data: label };
  }
}

export { LabelForgeClient };
export type { LabelRequest, LabelResponse, LabelAddress, LabelResult };
```

### 4.5 SwiftHaul — Last-Mile Delivery (Go)

```go
// swifthaul/pickups.go — Pickup scheduling and international handoff.

package swifthaul

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ── Types ──────────────────────────────────────────────────

type PickupRequest struct {
	TrackingNumber  string         `json:"tracking_number"`
	Carrier         string         `json:"carrier"`
	PickupAddress   PickupAddress  `json:"pickup_address"`
	PackageDetails  PackageDetails `json:"package_details"`
	PickupDate      string         `json:"pickup_date"`      // YYYY-MM-DD
	IsInternational bool           `json:"is_international"`
}

type PickupAddress struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	State   string `json:"state"`
	Zip     string `json:"zip"`
	Country string `json:"country"` // ISO 3166-1 alpha-2
}

type PackageDetails struct {
	WeightGrams int `json:"weight_grams"` // grams
	LengthCm    int `json:"length_cm"`    // centimeters
	WidthCm     int `json:"width_cm"`     // centimeters
	HeightCm    int `json:"height_cm"`    // centimeters
}

type PickupResponse struct {
	PickupID          string `json:"pickup_id"`          // pku_*
	Status            string `json:"status"`             // SCREAMING_SNAKE_CASE
	TrackingNumber    string `json:"tracking_number"`
	EstimatedDelivery string `json:"estimated_delivery"` // ISO 8601 UTC
}

// OceanBridge request types -- note conversions
type OceanBridgeBookingRequest struct {
	OriginCountry       string              `json:"origin_country"`       // ISO alpha-2
	DestinationCountry  string              `json:"destination_country"`  // ISO alpha-2
	Packages            []OceanBridgePackage `json:"packages"`
	DeclaredValue       string              `json:"declared_value"`       // string decimal
	Currency            string              `json:"currency"`             // UPPERCASE
	ContentsDescription string              `json:"contents_description"`
	HSCode              string              `json:"hs_code,omitempty"`
	CustomsInfo         CustomsInfo         `json:"customs_info"`
}

type OceanBridgePackage struct {
	WeightKg float64 `json:"weight_kg"` // kilograms (converted from grams)
	LengthCm int     `json:"length_cm"`
	WidthCm  int     `json:"width_cm"`
	HeightCm int     `json:"height_cm"`
}

type CustomsInfo struct {
	SenderName    string `json:"sender_name"`
	SenderTaxID   string `json:"sender_tax_id,omitempty"`
	RecipientName string `json:"recipient_name"`
}

// ── International Handoff ──────────────────────────────────

func HandoffToOceanBridge(
	oceanBridgeURL, oceanBridgeKey string,
	pickupID string,
	pkg PackageDetails,
	destCountry string,
	declaredValueCents int,
	currency string,
	description string,
) error {
	// CORRECT: Convert grams to kilograms
	weightKg := float64(pkg.WeightGrams) / 1000.0

	// CORRECT: Convert integer cents to string decimal
	declaredValueStr := fmt.Sprintf("%.2f", float64(declaredValueCents)/100.0)

	// CORRECT: Convert lowercase currency to uppercase for OceanBridge
	currencyUpper := strings.ToUpper(currency)

	booking := OceanBridgeBookingRequest{
		OriginCountry:      "US",
		DestinationCountry: destCountry, // ISO alpha-2
		Packages: []OceanBridgePackage{{
			WeightKg: weightKg,
			LengthCm: pkg.LengthCm,
			WidthCm:  pkg.WidthCm,
			HeightCm: pkg.HeightCm,
		}},
		DeclaredValue:       declaredValueStr,  // "29.99" not 2999
		Currency:            currencyUpper,     // "USD" not "usd"
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
	_ = body

	return nil
}

// ── Validation ─────────────────────────────────────────────

var validCarriers = map[string]bool{
	"ups": true, "fedex": true, "usps": true, "dhl": true,
}

func validatePickupRequest(req *PickupRequest) error {
	if req.TrackingNumber == "" {
		return fmt.Errorf("tracking_number is required")
	}
	if !validCarriers[req.Carrier] {
		return fmt.Errorf("invalid carrier: %s", req.Carrier)
	}
	if req.PackageDetails.WeightGrams < 1 {
		return fmt.Errorf("weight_grams must be positive, got %d", req.PackageDetails.WeightGrams)
	}
	// Validate date format
	if _, err := time.Parse("2006-01-02", req.PickupDate); err != nil {
		return fmt.Errorf("invalid pickup_date format (expected YYYY-MM-DD): %s", req.PickupDate)
	}
	return nil
}

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
		Status:            "SCHEDULED",  // SCREAMING_SNAKE_CASE
		TrackingNumber:    req.TrackingNumber,
		EstimatedDelivery: time.Now().AddDate(0, 0, 7).UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func writePickupError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{"code": code, "message": message},
	})
}

// unused import guard
var _ = strconv.Itoa
```

### 4.6 OceanBridge — International Shipping (Python)

```python
# oceanbridge/bookings.py — International shipping and customs.

from datetime import datetime, timezone
from decimal import Decimal
import hashlib
import hmac
import json
import re
from typing import Any

# ── Types ──────────────────────────────────────────────────

ISO_COUNTRY_PATTERN = re.compile(r"^[A-Z]{2}$")
VALID_CURRENCIES = {"USD", "EUR", "GBP"}

class BookingRequest:
    def __init__(self, data: dict[str, Any]) -> None:
        self.origin_country: str = data["origin_country"]
        self.destination_country: str = data["destination_country"]
        self.packages: list[dict[str, Any]] = data["packages"]
        self.declared_value: str = data["declared_value"]  # string decimal
        self.currency: str = data["currency"]               # UPPERCASE
        self.contents_description: str = data["contents_description"]
        self.hs_code: str | None = data.get("hs_code")
        self.customs_info: dict[str, str] = data["customs_info"]

    def validate(self) -> None:
        if not ISO_COUNTRY_PATTERN.match(self.origin_country):
            raise ValueError(f"Invalid origin_country: {self.origin_country}")
        if not ISO_COUNTRY_PATTERN.match(self.destination_country):
            raise ValueError(f"Invalid destination_country: {self.destination_country}")
        if self.currency not in VALID_CURRENCIES:
            raise ValueError(f"Invalid currency: {self.currency} (expected {VALID_CURRENCIES})")

        # Validate declared_value is a valid decimal string
        try:
            val = Decimal(self.declared_value)
            if val <= 0:
                raise ValueError(f"declared_value must be positive, got {self.declared_value}")
        except Exception as e:
            raise ValueError(f"Invalid declared_value: {self.declared_value}") from e

        for pkg in self.packages:
            if pkg.get("weight_kg", 0) <= 0:
                raise ValueError(f"weight_kg must be positive: {pkg.get('weight_kg')}")

class BookingResponse:
    def __init__(
        self,
        booking_id: str,
        status: str,
        estimated_arrival: str,
        tracking_url: str,
    ) -> None:
        self.booking_id = booking_id          # obk_*
        self.status = status                  # lowercase
        self.estimated_arrival = estimated_arrival  # ISO 8601 UTC
        self.tracking_url = tracking_url

    def to_dict(self) -> dict[str, Any]:
        now = datetime.now(timezone.utc).isoformat()  # ISO 8601 with timezone
        return {
            "booking_id": self.booking_id,
            "status": self.status,
            "estimated_arrival": self.estimated_arrival,
            "tracking_url": self.tracking_url,
            "created_at": now,
        }

# ── Booking Handler ────────────────────────────────────────

def create_booking(data: dict[str, Any]) -> dict[str, Any]:
    req = BookingRequest(data)
    req.validate()

    booking_id = f"obk_{int(datetime.now(timezone.utc).timestamp() * 1000)}"

    # CORRECT: estimated_arrival includes timezone (UTC)
    eta = datetime.now(timezone.utc).replace(
        hour=14, minute=0, second=0, microsecond=0,
    )
    estimated_arrival = eta.isoformat()  # "2024-01-20T14:00:00+00:00"

    resp = BookingResponse(
        booking_id=booking_id,
        status="pending",
        estimated_arrival=estimated_arrival,
        tracking_url=f"https://track.oceanbridge.io/{booking_id}",
    )

    return resp.to_dict()

# ── Customs Notification (to PingWave) ─────────────────────

def send_customs_notification(
    pingwave_url: str,
    pingwave_key: str,
    booking_id: str,
    event_type: str,
    recipient_email: str,
    estimated_arrival: str,
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
            "origin": origin_country,           # ISO alpha-2
            "destination": destination_country,  # ISO alpha-2
            "estimated_arrival": estimated_arrival,  # ISO 8601 with timezone
        },
        "callback_url": f"{pingwave_url}/callbacks/customs/{booking_id}",
    }

    # In production, would POST to pingwave_url + "/v1/notifications/send"
    # with Authorization: Bearer pingwave_key
    return payload
```

### 4.7 PingWave — Notifications (TypeScript)

```typescript
// pingwave-client.ts — Notification service with delivery callbacks.

import * as crypto from "crypto";

// ── Types ──────────────────────────────────────────────────

interface SendNotificationRequest {
  channel: "sms" | "email" | "webhook";
  recipient: string;
  template_id?: string;
  variables?: Record<string, string>;
  message?: string;
  callback_url?: string;
  metadata?: Record<string, string>;
}

interface NotificationResponse {
  notification_id: string;  // ntf_*
  channel: "sms" | "email" | "webhook";
  status: "queued" | "sent" | "delivered" | "failed" | "bounced";
  created_at: string;  // ISO 8601 UTC
}

interface DeliveryCallbackPayload {
  event_type: "delivery_confirmed" | "delivery_failed" | "notification_sent";
  order_id: string;
  status: "in_transit" | "out_for_delivery" | "delivered" | "failed";
  timestamp: string;  // ISO 8601 UTC
  details?: string;
}

type NotificationResult =
  | { ok: true; data: NotificationResponse }
  | { ok: false; status: number; error: { code: string; message: string } };

// ── Client ─────────────────────────────────────────────────

class PingWaveClient {
  private readonly baseUrl: string;
  private readonly apiKey: string;
  private readonly webhookSecret: string;

  constructor(baseUrl: string, apiKey: string, webhookSecret: string) {
    this.baseUrl = baseUrl;
    this.apiKey = apiKey;
    this.webhookSecret = webhookSecret;
  }

  async sendNotification(req: SendNotificationRequest): Promise<NotificationResult> {
    if (!req.channel) {
      throw new Error("channel is required");
    }
    if (!req.recipient) {
      throw new Error("recipient is required");
    }

    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/v1/notifications/send`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${this.apiKey}`,
        },
        body: JSON.stringify(req),
      });
    } catch (err) {
      throw new Error(`Network error calling PingWave: ${(err as Error).message}`);
    }

    if (!response.ok) {
      const errorBody = await response.json() as { error: { code: string; message: string } };
      return { ok: false, status: response.status, error: errorBody.error };
    }

    const notification = (await response.json()) as NotificationResponse;
    return { ok: true, data: notification };
  }

  // ── Send Delivery Callback to ShopStream ─────────────

  async sendDeliveryCallback(
    callbackUrl: string,
    payload: DeliveryCallbackPayload,
  ): Promise<void> {
    const body = JSON.stringify(payload);
    const timestamp = Math.floor(Date.now() / 1000).toString();
    const signedPayload = `${timestamp}.${body}`;
    const signature = crypto
      .createHmac("sha256", this.webhookSecret)
      .update(signedPayload)
      .digest("hex");

    let response: Response;
    try {
      response = await fetch(callbackUrl, {
        method: "POST",  // POST -- matches ShopStream's webhook endpoint
        headers: {
          "Content-Type": "application/json",
          "X-PingWave-Signature": signature,
          "X-PingWave-Timestamp": timestamp,
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
}

export { PingWaveClient };
export type {
  SendNotificationRequest, NotificationResponse,
  DeliveryCallbackPayload, NotificationResult,
};
```

### Expected Stricture Result

```
PASS  0 violations
```

Stricture must produce zero violations. All 8 inter-service contracts have matching types across boundaries. Currency is integer cents everywhere (except OceanBridge's declared_value which is correctly a string decimal per customs requirements, and the conversion is explicit). Weight is consistently in grams with proper kg conversion for OceanBridge. Dimensions convert from mm (internal) to cm (external) correctly. All timestamps are ISO 8601 with UTC timezone. Country codes are ISO 3166-1 alpha-2. Status enums match the documented contract per service. Webhook signatures use HMAC-SHA256 with constant-time comparison. Error handling wraps all fetch() calls in try/catch with status code checks.
