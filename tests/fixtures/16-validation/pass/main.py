"""Stripe API client with type-safe models and error handling."""

import hashlib
import hmac
import os
from typing import Generator, Optional
from decimal import Decimal

import httpx
from pydantic import BaseModel, Field, field_validator


class ChargeModel(BaseModel):
    """Stripe charge object with validation."""

    id: str = Field(..., pattern=r"^ch_[a-zA-Z0-9]+$")
    amount: int = Field(..., ge=50, le=99999999)  # Stripe cents limits
    currency: str = Field(..., pattern=r"^[a-z]{3}$")
    status: str
    customer: Optional[str] = None
    description: Optional[str] = None
    failure_message: Optional[str] = None
    balance_transaction: Optional[str] = None

    @field_validator("status")
    @classmethod
    def validate_status(cls, v: str) -> str:
        """Ensure status is a valid Stripe charge status."""
        valid_statuses = {"succeeded", "pending", "failed"}
        if v not in valid_statuses:
            raise ValueError(f"Invalid status: {v}")
        return v

    @field_validator("currency")
    @classmethod
    def validate_currency(cls, v: str) -> str:
        """Ensure currency is lowercase ISO 4217."""
        return v.lower()


class ChargeListResponse(BaseModel):
    """Paginated list of charges."""

    object: str = Field(..., pattern=r"^list$")
    data: list[ChargeModel]
    has_more: bool
    url: str


class WebhookEvent(BaseModel):
    """Stripe webhook event."""

    id: str
    type: str
    data: dict
    created: int


class StripeClient:
    """Type-safe Stripe API client with error handling."""

    def __init__(self, api_key: str, base_url: str = "https://api.stripe.com"):
        """Initialize Stripe client.

        Args:
            api_key: Stripe secret key (sk_test_* or sk_live_*)
            base_url: API base URL (default: production)
        """
        self.api_key = api_key
        self.base_url = base_url
        self.client = httpx.Client(
            base_url=base_url,
            headers={
                "Authorization": f"Bearer {api_key}",
                "Content-Type": "application/x-www-form-urlencoded",
            },
            timeout=30.0,
        )

    def create_charge(
        self,
        amount: int,
        currency: str,
        source: str,
        description: Optional[str] = None,
        idempotency_key: Optional[str] = None,
    ) -> ChargeModel:
        """Create a new charge.

        Args:
            amount: Amount in cents (50 to 99999999)
            currency: Three-letter ISO currency code
            source: Payment source (card token or source ID)
            description: Optional charge description
            idempotency_key: Optional idempotency key for safe retries

        Returns:
            ChargeModel: Created charge object

        Raises:
            httpx.HTTPStatusError: On API error (400, 402, 404, etc.)
            ValueError: On invalid parameters
        """
        if not (50 <= amount <= 99999999):
            raise ValueError(f"Amount must be between 50 and 99999999, got {amount}")

        if len(currency) != 3 or not currency.isalpha():
            raise ValueError(f"Invalid currency code: {currency}")

        payload = {
            "amount": amount,
            "currency": currency.lower(),
            "source": source,
        }

        if description:
            payload["description"] = description

        headers = {}
        if idempotency_key:
            headers["Idempotency-Key"] = idempotency_key

        try:
            resp = self.client.post(
                "/v1/charges",
                data=payload,
                headers=headers if headers else None,
            )
            resp.raise_for_status()
            return ChargeModel(**resp.json())
        except httpx.HTTPStatusError as e:
            # Re-raise with context
            raise httpx.HTTPStatusError(
                f"Failed to create charge: {e.response.text}",
                request=e.request,
                response=e.response,
            ) from e

    def get_charge(self, charge_id: str) -> ChargeModel:
        """Retrieve a charge by ID.

        Args:
            charge_id: Charge ID (ch_*)

        Returns:
            ChargeModel: Charge object

        Raises:
            httpx.HTTPStatusError: On API error
            ValueError: On invalid charge_id format
        """
        if not charge_id.startswith("ch_"):
            raise ValueError(f"Invalid charge ID format: {charge_id}")

        try:
            resp = self.client.get(f"/v1/charges/{charge_id}")
            resp.raise_for_status()
            return ChargeModel(**resp.json())
        except httpx.HTTPStatusError as e:
            raise httpx.HTTPStatusError(
                f"Failed to retrieve charge {charge_id}: {e.response.text}",
                request=e.request,
                response=e.response,
            ) from e

    def list_charges(
        self,
        limit: int = 10,
        starting_after: Optional[str] = None,
    ) -> Generator[ChargeModel, None, None]:
        """List charges with automatic pagination.

        Args:
            limit: Number of charges per page (1-100)
            starting_after: Cursor for pagination

        Yields:
            ChargeModel: Individual charge objects

        Raises:
            httpx.HTTPStatusError: On API error
        """
        if not (1 <= limit <= 100):
            raise ValueError(f"Limit must be between 1 and 100, got {limit}")

        cursor = starting_after

        while True:
            params = {"limit": limit}
            if cursor:
                params["starting_after"] = cursor

            try:
                resp = self.client.get("/v1/charges", params=params)
                resp.raise_for_status()
                page = ChargeListResponse(**resp.json())
            except httpx.HTTPStatusError as e:
                raise httpx.HTTPStatusError(
                    f"Failed to list charges: {e.response.text}",
                    request=e.request,
                    response=e.response,
                ) from e

            for charge in page.data:
                yield charge

            if not page.has_more:
                break

            # Set cursor to last item's ID
            if page.data:
                cursor = page.data[-1].id
            else:
                break

    def verify_webhook(
        self,
        payload: bytes,
        signature: str,
        secret: str,
    ) -> WebhookEvent:
        """Verify and parse a Stripe webhook event.

        Args:
            payload: Raw request body (bytes)
            signature: Stripe-Signature header value
            secret: Webhook signing secret (whsec_*)

        Returns:
            WebhookEvent: Parsed and verified event

        Raises:
            ValueError: On signature verification failure
        """
        # Parse signature header
        sig_parts = {}
        for part in signature.split(","):
            key, value = part.split("=", 1)
            sig_parts[key] = value

        if "t" not in sig_parts or "v1" not in sig_parts:
            raise ValueError("Invalid signature format")

        timestamp = sig_parts["t"]
        received_sig = sig_parts["v1"]

        # Compute expected signature
        signed_payload = f"{timestamp}.{payload.decode('utf-8')}"
        expected_sig = hmac.new(
            secret.encode("utf-8"),
            signed_payload.encode("utf-8"),
            hashlib.sha256,
        ).hexdigest()

        # Constant-time comparison
        if not hmac.compare_digest(expected_sig, received_sig):
            raise ValueError("Signature verification failed")

        # Parse event
        import json
        event_data = json.loads(payload)
        return WebhookEvent(**event_data)

    def calculate_total_with_tax(
        self,
        amount_cents: int,
        tax_rate: Decimal,
    ) -> int:
        """Calculate total amount with tax (no float arithmetic).

        Args:
            amount_cents: Base amount in cents
            tax_rate: Tax rate as Decimal (e.g., Decimal("0.08") for 8%)

        Returns:
            int: Total amount in cents (rounded)
        """
        amount_decimal = Decimal(amount_cents)
        total = amount_decimal * (Decimal("1") + tax_rate)
        return int(total.quantize(Decimal("1")))  # Round to nearest cent

    def close(self) -> None:
        """Close the HTTP client."""
        self.client.close()

    def __enter__(self):
        """Context manager entry."""
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit."""
        self.close()


def get_failure_message_safe(charge: ChargeModel) -> str:
    """Safely get failure message with null check.

    Args:
        charge: Charge object

    Returns:
        str: Uppercase failure message or "N/A"
    """
    if charge.failure_message is None:
        return "N/A"
    return charge.failure_message.upper()
