# 16 — Stripe Payments API (Python)

**Why included:** Python dataclasses/pydantic, httpx, pytest fixtures, type hints, HMAC webhook verification, generator-based pagination, Field validators.

**Language:** Python 3.10+
**Framework:** pytest, httpx, pydantic
**Domain:** Payment processing API client

---

## Manifest Fragment

```json
{
  "id": "16-stripe-python",
  "language": "python",
  "domain": "payments",
  "test_count": 16,
  "endpoints": [
    "POST /v1/charges",
    "GET /v1/charges/:id",
    "GET /v1/charges",
    "POST /webhook"
  ],
  "perfect_patterns": [
    "pydantic_models_with_validators",
    "httpx_client_with_error_handling",
    "hmac_webhook_verification",
    "generator_pagination",
    "pytest_fixtures",
    "pytest_parametrize",
    "pytest_raises",
    "type_hints_full_coverage"
  ],
  "bug_categories": [
    "no_error_handling",
    "unchecked_status_codes",
    "weak_assertions",
    "missing_exception_tests",
    "incomplete_payloads",
    "missing_model_fields",
    "type_errors",
    "incomplete_conditionals",
    "missing_validators",
    "missing_prefix_validation",
    "float_arithmetic",
    "missing_null_checks",
    "no_hmac_verification",
    "pagination_ignored",
    "missing_idempotency"
  ]
}
```

---

## PERFECT Implementation

**File:** `src/stripe_client.py`

```python
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
```

---

**File:** `tests/test_stripe_client.py`

```python
"""Tests for Stripe API client."""

import hashlib
import hmac
import json
from decimal import Decimal
from unittest import mock

import httpx
import pytest
from pydantic import ValidationError

from src.stripe_client import (
    ChargeModel,
    StripeClient,
    WebhookEvent,
    get_failure_message_safe,
)


@pytest.fixture
def stripe_client():
    """Create a StripeClient instance for testing."""
    return StripeClient(api_key="sk_test_12345", base_url="https://api.stripe.test")


@pytest.fixture
def mock_charge_response():
    """Mock successful charge response."""
    return {
        "id": "ch_1ABC23def456",
        "amount": 5000,
        "currency": "usd",
        "status": "succeeded",
        "customer": "cus_123",
        "description": "Test charge",
        "failure_message": None,
        "balance_transaction": "txn_789",
    }


@pytest.fixture
def webhook_secret():
    """Webhook signing secret."""
    return "whsec_test_secret_12345"


class TestChargeModel:
    """Test ChargeModel validation."""

    def test_valid_charge(self):
        """Valid charge should parse successfully."""
        charge = ChargeModel(
            id="ch_1ABC23def456",
            amount=5000,
            currency="usd",
            status="succeeded",
        )
        assert charge.id == "ch_1ABC23def456"
        assert charge.amount == 5000
        assert charge.currency == "usd"
        assert charge.status == "succeeded"

    @pytest.mark.parametrize("charge_id,should_pass", [
        ("ch_1ABC23def456", True),
        ("ch_", False),  # Too short
        ("ch_abc", True),
        ("pi_123", False),  # Wrong prefix
        ("123", False),  # No prefix
    ])
    def test_charge_id_validation(self, charge_id, should_pass):
        """Charge ID must match ch_* pattern."""
        if should_pass:
            charge = ChargeModel(
                id=charge_id,
                amount=5000,
                currency="usd",
                status="succeeded",
            )
            assert charge.id.startswith("ch_")
        else:
            with pytest.raises(ValidationError) as exc_info:
                ChargeModel(
                    id=charge_id,
                    amount=5000,
                    currency="usd",
                    status="succeeded",
                )
            assert "id" in str(exc_info.value)

    @pytest.mark.parametrize("amount,should_pass", [
        (50, True),  # Minimum
        (5000, True),  # Normal
        (99999999, True),  # Maximum
        (49, False),  # Below minimum
        (100000000, False),  # Above maximum
        (0, False),  # Zero
        (-100, False),  # Negative
    ])
    def test_amount_validation(self, amount, should_pass):
        """Amount must be between 50 and 99999999 cents."""
        if should_pass:
            charge = ChargeModel(
                id="ch_123",
                amount=amount,
                currency="usd",
                status="succeeded",
            )
            assert 50 <= charge.amount <= 99999999
        else:
            with pytest.raises(ValidationError) as exc_info:
                ChargeModel(
                    id="ch_123",
                    amount=amount,
                    currency="usd",
                    status="succeeded",
                )
            assert "amount" in str(exc_info.value)

    @pytest.mark.parametrize("currency,expected", [
        ("USD", "usd"),  # Uppercase normalized
        ("usd", "usd"),  # Already lowercase
        ("EUR", "eur"),
        ("GBP", "gbp"),
    ])
    def test_currency_normalization(self, currency, expected):
        """Currency should be normalized to lowercase."""
        charge = ChargeModel(
            id="ch_123",
            amount=5000,
            currency=currency,
            status="succeeded",
        )
        assert charge.currency == expected

    @pytest.mark.parametrize("currency,should_pass", [
        ("usd", True),
        ("eur", True),
        ("gbp", True),
        ("us", False),  # Too short
        ("usdd", False),  # Too long
        ("123", False),  # Not alpha
    ])
    def test_currency_format(self, currency, should_pass):
        """Currency must be 3-letter ISO code."""
        if should_pass:
            charge = ChargeModel(
                id="ch_123",
                amount=5000,
                currency=currency,
                status="succeeded",
            )
            assert len(charge.currency) == 3
        else:
            with pytest.raises(ValidationError):
                ChargeModel(
                    id="ch_123",
                    amount=5000,
                    currency=currency,
                    status="succeeded",
                )

    @pytest.mark.parametrize("status,should_pass", [
        ("succeeded", True),
        ("pending", True),
        ("failed", True),
        ("unknown", False),
        ("cancelled", False),
    ])
    def test_status_validation(self, status, should_pass):
        """Status must be valid Stripe charge status."""
        if should_pass:
            charge = ChargeModel(
                id="ch_123",
                amount=5000,
                currency="usd",
                status=status,
            )
            assert charge.status in {"succeeded", "pending", "failed"}
        else:
            with pytest.raises(ValidationError) as exc_info:
                ChargeModel(
                    id="ch_123",
                    amount=5000,
                    currency="usd",
                    status=status,
                )
            assert "status" in str(exc_info.value)


class TestStripeClientCreateCharge:
    """Test StripeClient.create_charge()."""

    def test_create_charge_success(self, stripe_client, mock_charge_response):
        """Successful charge creation with all fields."""
        with mock.patch.object(stripe_client.client, "post") as mock_post:
            mock_resp = mock.Mock()
            mock_resp.json.return_value = mock_charge_response
            mock_resp.raise_for_status = mock.Mock()
            mock_post.return_value = mock_resp

            charge = stripe_client.create_charge(
                amount=5000,
                currency="USD",
                source="tok_visa",
                description="Test charge",
                idempotency_key="idem_123",
            )

            assert charge.id == "ch_1ABC23def456"
            assert charge.amount == 5000
            assert charge.currency == "usd"
            assert charge.status == "succeeded"
            assert charge.description == "Test charge"

            # Verify request
            mock_post.assert_called_once()
            call_args = mock_post.call_args
            assert call_args[0][0] == "/v1/charges"
            assert call_args[1]["data"]["amount"] == 5000
            assert call_args[1]["data"]["currency"] == "usd"
            assert call_args[1]["data"]["source"] == "tok_visa"
            assert call_args[1]["data"]["description"] == "Test charge"
            assert call_args[1]["headers"]["Idempotency-Key"] == "idem_123"

    def test_create_charge_minimal(self, stripe_client, mock_charge_response):
        """Create charge with only required fields."""
        with mock.patch.object(stripe_client.client, "post") as mock_post:
            mock_resp = mock.Mock()
            mock_resp.json.return_value = mock_charge_response
            mock_resp.raise_for_status = mock.Mock()
            mock_post.return_value = mock_resp

            charge = stripe_client.create_charge(
                amount=5000,
                currency="usd",
                source="tok_visa",
            )

            assert charge.id.startswith("ch_")
            assert charge.amount == 5000

            # Verify no description or idempotency key
            call_args = mock_post.call_args
            assert "description" not in call_args[1]["data"]
            assert call_args[1]["headers"] is None

```

---

## Bug Cases

### B01 — No Error Handling on HTTP Call (TQ-error-path-coverage)

**Bug:** No try/except block around httpx.post() call, leaving HTTP errors unhandled.

**Expected violation:** `TQ-error-path-coverage`

```python
"""Stripe client without error handling."""

import httpx
from pydantic import BaseModel, Field


class ChargeModel(BaseModel):
    id: str
    amount: int
    currency: str
    status: str


class StripeClient:
    def __init__(self, api_key: str):
        self.client = httpx.Client(
            base_url="https://api.stripe.com",
            headers={"Authorization": f"Bearer {api_key}"},
        )

    def create_charge(self, amount: int, currency: str, source: str) -> ChargeModel:
        """Create charge without error handling."""
        resp = self.client.post(
            "/v1/charges",
            data={"amount": amount, "currency": currency, "source": source},
        )
        resp.raise_for_status()
        return ChargeModel(**resp.json())
```

**Why Stricture catches this:** `TQ-error-path-coverage` detects that `httpx.post()` can raise `httpx.HTTPStatusError`, `httpx.ConnectError`, or `httpx.TimeoutException`, but there's no try/except block handling these exceptions. The function propagates network errors to callers without context or recovery options.

---

### B02 — No Status Code Check (CTR-status-code-handling)

**Bug:** Missing `response.raise_for_status()` or manual status code validation, silently accepting 4xx/5xx responses.

**Expected violation:** `CTR-status-code-handling`

```python
"""Stripe client without status validation."""

import httpx
from pydantic import BaseModel


class ChargeModel(BaseModel):
    id: str
    amount: int
    currency: str
    status: str


class StripeClient:
    def __init__(self, api_key: str):
        self.client = httpx.Client(
            base_url="https://api.stripe.com",
            headers={"Authorization": f"Bearer {api_key}"},
        )

    def create_charge(self, amount: int, currency: str, source: str) -> ChargeModel:
        """Create charge without status check."""
        try:
            resp = self.client.post(
                "/v1/charges",
                data={"amount": amount, "currency": currency, "source": source},
            )
            # No raise_for_status() call
            return ChargeModel(**resp.json())
        except httpx.HTTPError as e:
            raise ValueError(f"HTTP error: {e}")
```

**Why Stricture catches this:** `CTR-status-code-handling` detects that the httpx Response object is used without calling `.raise_for_status()` or checking `.status_code`. A 402 (Payment Required) or 429 (Rate Limit) response would be treated as success, causing Pydantic to fail parsing the error JSON.

---

### B03 — Shallow Test Assertions (TQ-no-shallow-assertions)

**Bug:** Test only checks `assert result is not None` instead of validating actual field values.

**Expected violation:** `TQ-no-shallow-assertions`

```python
"""Test with shallow assertions."""

import pytest
from unittest import mock
from src.stripe_client import StripeClient, ChargeModel


@pytest.fixture
def stripe_client():
    return StripeClient(api_key="sk_test_12345")


def test_create_charge_shallow(stripe_client):
    """Test with shallow assertion (anti-pattern)."""
    with mock.patch.object(stripe_client.client, "post") as mock_post:
        mock_resp = mock.Mock()
        mock_resp.json.return_value = {
            "id": "ch_123",
            "amount": 5000,
            "currency": "usd",
            "status": "succeeded",
        }
        mock_resp.raise_for_status = mock.Mock()
        mock_post.return_value = mock_resp

        charge = stripe_client.create_charge(
            amount=5000,
            currency="usd",
            source="tok_visa",
        )

        # Shallow assertion - doesn't verify fields
        assert charge is not None
        # Missing: assert charge.amount == 5000, etc.
```

**Why Stricture catches this:** `TQ-no-shallow-assertions` flags tests that only assert truthiness (`assert x is not None`, `assert result`, `assert len(x) > 0`) without validating specific field values. These tests pass even when returned data is completely wrong.

---

### B04 — Missing Negative Tests (TQ-negative-cases)

**Bug:** No tests for error responses like 402 Payment Required, 429 Rate Limit, or 400 Bad Request.

**Expected violation:** `TQ-negative-cases`

```python
"""Test suite missing negative cases."""

import pytest
from unittest import mock
from src.stripe_client import StripeClient


@pytest.fixture
def stripe_client():
    return StripeClient(api_key="sk_test_12345")


def test_create_charge_success(stripe_client):
    """Test only the happy path."""
    with mock.patch.object(stripe_client.client, "post") as mock_post:
        mock_resp = mock.Mock()
        mock_resp.json.return_value = {
            "id": "ch_123",
            "amount": 5000,
            "currency": "usd",
            "status": "succeeded",
        }
        mock_resp.raise_for_status = mock.Mock()
        mock_post.return_value = mock_resp

        charge = stripe_client.create_charge(
            amount=5000,
            currency="usd",
            source="tok_visa",
        )
        assert charge.id == "ch_123"


# Missing tests for:
# - 402 Payment Required
# - 429 Rate Limit Exceeded
# - 400 Invalid Request
# - 401 Unauthorized
# - 500 Internal Server Error
```

**Why Stricture catches this:** `TQ-negative-cases` detects test suites that only cover success paths without testing exception cases. For a payment API, missing tests for payment failures (402), rate limits (429), and validation errors (400) leaves critical error handling unverified.

---

### B05 — Request Missing Required Fields (CTR-request-shape)

**Bug:** Omits "currency" from charge creation payload, violating Stripe API contract.

**Expected violation:** `CTR-request-shape`

```python
"""Stripe client with incomplete request payload."""

import httpx
from pydantic import BaseModel


class ChargeModel(BaseModel):
    id: str
    amount: int
    currency: str
    status: str


class StripeClient:
    def __init__(self, api_key: str):
        self.client = httpx.Client(
            base_url="https://api.stripe.com",
            headers={"Authorization": f"Bearer {api_key}"},
        )

    def create_charge(self, amount: int, currency: str, source: str) -> ChargeModel:
        """Create charge with incomplete payload."""
        payload = {
            "amount": amount,
            # Missing: "currency": currency
            "source": source,
        }

        try:
            resp = self.client.post("/v1/charges", data=payload)
            resp.raise_for_status()
            return ChargeModel(**resp.json())
        except httpx.HTTPStatusError as e:
            raise ValueError(f"Charge failed: {e}")
```

**Why Stricture catches this:** `CTR-request-shape` validates that request payloads include all required fields per API contract. The Stripe charges endpoint requires "amount", "currency", and "source". Omitting "currency" causes a 400 Bad Request, but the code doesn't reflect this invariant.

---

### B06 — Response Type Mismatch (CTR-response-shape)

**Bug:** Pydantic model missing "livemode" field that exists in actual Stripe API responses.

**Expected violation:** `CTR-response-shape`

```python
"""Stripe model with incomplete response shape."""

import httpx
from pydantic import BaseModel


class ChargeModel(BaseModel):
    """Incomplete charge model missing livemode field."""
    id: str
    amount: int
    currency: str
    status: str
    # Missing: livemode field (bool) present in real API responses


class StripeClient:
    def __init__(self, api_key: str):
        self.client = httpx.Client(
            base_url="https://api.stripe.com",
            headers={"Authorization": f"Bearer {api_key}"},
        )

    def create_charge(self, amount: int, currency: str, source: str) -> ChargeModel:
        """Create charge with incomplete model."""
        try:
            resp = self.client.post(
                "/v1/charges",
                data={"amount": amount, "currency": currency, "source": source},
            )
            resp.raise_for_status()
            # Pydantic will fail if response contains unexpected "livemode" field
            # (unless using extra="allow")
            return ChargeModel(**resp.json())
        except httpx.HTTPStatusError as e:
            raise ValueError(f"Charge failed: {e}")
```

**Why Stricture catches this:** `CTR-response-shape` validates that response models match the actual API schema. Stripe charge responses include a "livemode" boolean field, but the ChargeModel doesn't define it. This causes Pydantic ValidationError when parsing real responses (unless `extra="allow"` is set, which weakens type safety).

---

### B07 — Wrong Field Types (CTR-manifest-conformance)

**Bug:** Using `float` for amount instead of `int`, violating Stripe's cent-based currency model.

**Expected violation:** `CTR-manifest-conformance`

```python
"""Stripe client with incorrect type for amount."""

import httpx
from pydantic import BaseModel


class ChargeModel(BaseModel):
    id: str
    amount: float  # Wrong: should be int (cents)
    currency: str
    status: str


class StripeClient:
    def __init__(self, api_key: str):
        self.client = httpx.Client(
            base_url="https://api.stripe.com",
            headers={"Authorization": f"Bearer {api_key}"},
        )

    def create_charge(self, amount: float, currency: str, source: str) -> ChargeModel:
        """Create charge with wrong amount type."""
        try:
            resp = self.client.post(
                "/v1/charges",
                data={
                    "amount": int(amount),  # Manual conversion hides type mismatch
                    "currency": currency,
                    "source": source,
                },
            )
            resp.raise_for_status()
            return ChargeModel(**resp.json())
        except httpx.HTTPStatusError as e:
            raise ValueError(f"Charge failed: {e}")
```

**Why Stricture catches this:** `CTR-manifest-conformance` validates that field types match API specifications. Stripe amounts are always integers representing the smallest currency unit (cents for USD). Using `float` introduces floating-point precision errors and contradicts the API contract (e.g., 50.75 cents is invalid).

---

### B08 — Incomplete Enum Handling (CTR-strictness-parity)

**Bug:** Status validator only checks for "succeeded" and "pending", missing "failed" state.

**Expected violation:** `CTR-strictness-parity`

```python
"""Stripe model with incomplete status enum."""

import httpx
from pydantic import BaseModel, field_validator


class ChargeModel(BaseModel):
    id: str
    amount: int
    currency: str
    status: str

    @field_validator("status")
    @classmethod
    def validate_status(cls, v: str) -> str:
        """Incomplete status validation."""
        valid_statuses = {"succeeded", "pending"}  # Missing: "failed"
        if v not in valid_statuses:
            raise ValueError(f"Invalid status: {v}")
        return v


class StripeClient:
    def __init__(self, api_key: str):
        self.client = httpx.Client(
            base_url="https://api.stripe.com",
            headers={"Authorization": f"Bearer {api_key}"},
        )

    def create_charge(self, amount: int, currency: str, source: str) -> ChargeModel:
        """Create charge with incomplete status enum."""
        try:
            resp = self.client.post(
                "/v1/charges",
                data={"amount": amount, "currency": currency, "source": source},
            )
            resp.raise_for_status()
            # Will raise ValidationError if charge fails (status="failed")
            return ChargeModel(**resp.json())
        except httpx.HTTPStatusError as e:
            raise ValueError(f"Charge failed: {e}")
```

**Why Stricture catches this:** `CTR-strictness-parity` validates that enum validators include all possible values per API documentation. Stripe charges can have status "succeeded", "pending", or "failed". Omitting "failed" causes Pydantic to reject legitimate API responses when a charge fails.

---

### B09 — Missing Range Validation (CTR-strictness-parity)

**Bug:** No validation that amount is >= 50 cents (Stripe's minimum charge).

**Expected violation:** `CTR-strictness-parity`

```python
"""Stripe client without minimum amount validation."""

import httpx
from pydantic import BaseModel


class ChargeModel(BaseModel):
    id: str
    amount: int  # Missing: ge=50 constraint
    currency: str
    status: str


class StripeClient:
    def __init__(self, api_key: str):
        self.client = httpx.Client(
            base_url="https://api.stripe.com",
            headers={"Authorization": f"Bearer {api_key}"},
        )

    def create_charge(self, amount: int, currency: str, source: str) -> ChargeModel:
        """Create charge without amount range validation."""
        # No check for amount >= 50
        try:
            resp = self.client.post(
                "/v1/charges",
                data={"amount": amount, "currency": currency, "source": source},
            )
            resp.raise_for_status()
            return ChargeModel(**resp.json())
        except httpx.HTTPStatusError as e:
            raise ValueError(f"Charge failed: {e}")


# Invalid call (should fail validation):
client = StripeClient(api_key="sk_test_123")
client.create_charge(amount=10, currency="usd", source="tok_visa")  # 10 cents < 50
```

**Why Stricture catches this:** `CTR-strictness-parity` validates that numeric fields enforce API-specified constraints. Stripe requires minimum charge of 50 cents (amount >= 50). Without Pydantic's `ge=50` constraint, invalid amounts are sent to the API, resulting in 400 errors that could be prevented client-side.

---

### B10 — Format Not Validated (CTR-strictness-parity)

**Bug:** Charge ID used without validating "ch_" prefix, accepting invalid IDs.

**Expected violation:** `CTR-strictness-parity`

```python
"""Stripe client without ID format validation."""

import httpx
from pydantic import BaseModel


class ChargeModel(BaseModel):
    id: str  # Missing: pattern=r"^ch_[a-zA-Z0-9]+$"
    amount: int
    currency: str
    status: str


class StripeClient:
    def __init__(self, api_key: str):
        self.client = httpx.Client(
            base_url="https://api.stripe.com",
            headers={"Authorization": f"Bearer {api_key}"},
        )

    def get_charge(self, charge_id: str) -> ChargeModel:
        """Get charge without ID format validation."""
        # No check for ch_* prefix
        try:
            resp = self.client.get(f"/v1/charges/{charge_id}")
            resp.raise_for_status()
            return ChargeModel(**resp.json())
        except httpx.HTTPStatusError as e:
            raise ValueError(f"Failed to get charge: {e}")


# Invalid call (should fail validation):
client = StripeClient(api_key="sk_test_123")
client.get_charge("pi_123")  # Payment Intent ID, not Charge ID
```

**Why Stricture catches this:** `CTR-strictness-parity` validates that string fields enforce format constraints. Stripe charge IDs always start with "ch_" prefix. Without Pydantic's `pattern` constraint, invalid IDs (like payment intent IDs "pi_*") are accepted, causing 404 errors instead of client-side validation failures.

---

### B11 — Precision Loss on Currency (CTR-strictness-parity)

**Bug:** Using `float` for currency calculations instead of `Decimal`, causing rounding errors.

**Expected violation:** `CTR-strictness-parity`

```python
"""Stripe client with float arithmetic (precision loss)."""

import httpx
from pydantic import BaseModel


class ChargeModel(BaseModel):
    id: str
    amount: int
    currency: str
    status: str


class StripeClient:
    def __init__(self, api_key: str):
        self.client = httpx.Client(
            base_url="https://api.stripe.com",
            headers={"Authorization": f"Bearer {api_key}"},
        )

    def calculate_total_with_tax(self, amount_cents: int, tax_rate: float) -> int:
        """Calculate total with float arithmetic (precision loss)."""
        # Using float causes rounding errors
        total = amount_cents * (1.0 + tax_rate)
        return int(total)  # Truncates instead of proper rounding


# Example precision loss:
client = StripeClient(api_key="sk_test_123")
# 5000 cents * 1.08 = 5400.0, but float arithmetic may give 5399.999999
total = client.calculate_total_with_tax(5000, 0.08)
# Expected: 5400, Actual: 5399 (off by 1 cent due to truncation)
```

**Why Stricture catches this:** `CTR-strictness-parity` validates that monetary calculations use `Decimal` instead of `float` to avoid precision loss. Financial APIs require exact arithmetic. Using float for tax calculations can cause rounding errors (e.g., 5000 * 1.08 = 5399.999... truncates to 5399 instead of 5400).

---

### B12 — Nullable Field Crashes (CTR-response-shape)

**Bug:** Accessing Optional[str] field without None check, causing AttributeError.

**Expected violation:** `CTR-response-shape`

```python
"""Stripe client with missing null check."""

import httpx
from typing import Optional
from pydantic import BaseModel


class ChargeModel(BaseModel):
    id: str
    amount: int
    currency: str
    status: str
    failure_message: Optional[str] = None


class StripeClient:
    def __init__(self, api_key: str):
        self.client = httpx.Client(
            base_url="https://api.stripe.com",
            headers={"Authorization": f"Bearer {api_key}"},
        )

    def get_failure_summary(self, charge: ChargeModel) -> str:
        """Get failure message without null check."""
        # Crashes if failure_message is None
        return charge.failure_message.upper()  # AttributeError: 'NoneType' has no attribute 'upper'


# Usage that crashes:
charge = ChargeModel(
    id="ch_123",
    amount=5000,
    currency="usd",
    status="succeeded",
    failure_message=None,  # Successful charge has no failure message
)
summary = StripeClient.get_failure_summary(None, charge)  # Crashes
```

**Why Stricture catches this:** `CTR-response-shape` validates that Optional fields are checked for None before method calls. Stripe charge responses have nullable fields like `failure_message` (None for successful charges). Calling `.upper()` without a None check causes AttributeError at runtime.

---

### B13 — Missing Webhook Verification (CTR-request-shape)

**Bug:** Processing webhook payload without HMAC signature verification, accepting forged events.

**Expected violation:** `CTR-request-shape`

```python
"""Stripe webhook handler without signature verification."""

import json
import httpx
from pydantic import BaseModel


class WebhookEvent(BaseModel):
    id: str
    type: str
    data: dict


class StripeWebhookHandler:
    def __init__(self, api_key: str):
        self.api_key = api_key

    def process_webhook(self, payload: bytes, signature: str) -> WebhookEvent:
        """Process webhook without signature verification."""
        # Missing: HMAC signature verification
        # Directly parses payload without validating authenticity
        event_data = json.loads(payload)
        event = WebhookEvent(**event_data)

        # Process event (vulnerable to forged requests)
        if event.type == "charge.succeeded":
            self.handle_charge_success(event.data)

        return event

    def handle_charge_success(self, data: dict):
        """Handle successful charge (called on unverified webhook)."""
        print(f"Charge succeeded: {data.get('id')}")


# Attacker can forge webhook events without signature check
```

**Why Stricture catches this:** `CTR-request-shape` validates that webhook endpoints verify HMAC signatures before processing payloads. Stripe webhooks include a `Stripe-Signature` header with an HMAC-SHA256 signature. Without verification using `hmac.compare_digest()`, attackers can send forged events to trigger unauthorized actions.

---

### B14 — Pagination Terminated Early (CTR-response-shape)

**Bug:** Ignoring `has_more` field in list response, returning only first page.

**Expected violation:** `CTR-response-shape`

```python
"""Stripe client with incomplete pagination."""

import httpx
from typing import List
from pydantic import BaseModel


class ChargeModel(BaseModel):
    id: str
    amount: int
    currency: str
    status: str


class ChargeListResponse(BaseModel):
    data: List[ChargeModel]
    has_more: bool
    url: str


class StripeClient:
    def __init__(self, api_key: str):
        self.client = httpx.Client(
            base_url="https://api.stripe.com",
            headers={"Authorization": f"Bearer {api_key}"},
        )

    def list_charges(self, limit: int = 10) -> List[ChargeModel]:
        """List charges with broken pagination."""
        try:
            resp = self.client.get("/v1/charges", params={"limit": limit})
            resp.raise_for_status()
            page = ChargeListResponse(**resp.json())

            # Returns only first page, ignoring has_more
            return page.data
        except httpx.HTTPStatusError as e:
            raise ValueError(f"Failed to list charges: {e}")


# If account has 100 charges but limit=10, only returns 10 instead of 100
```

**Why Stricture catches this:** `CTR-response-shape` validates that paginated responses check the `has_more` field and fetch subsequent pages. Stripe list endpoints return `has_more: true` when additional pages exist. Ignoring this field causes data loss—users with 100 charges only see the first 10.

---

### B15 — Race Condition on Retry (CTR-request-shape)

**Bug:** No `Idempotency-Key` header on charge creation, causing duplicate charges on network retry.

**Expected violation:** `CTR-request-shape`

```python
"""Stripe client without idempotency protection."""

import httpx
from pydantic import BaseModel


class ChargeModel(BaseModel):
    id: str
    amount: int
    currency: str
    status: str


class StripeClient:
    def __init__(self, api_key: str):
        self.client = httpx.Client(
            base_url="https://api.stripe.com",
            headers={"Authorization": f"Bearer {api_key}"},
            timeout=httpx.Timeout(5.0, connect=2.0),
        )

    def create_charge_with_retry(
        self,
        amount: int,
        currency: str,
        source: str,
        max_retries: int = 3,
    ) -> ChargeModel:
        """Create charge with retry (no idempotency key)."""
        for attempt in range(max_retries):
            try:
                resp = self.client.post(
                    "/v1/charges",
                    data={"amount": amount, "currency": currency, "source": source},
                    # Missing: headers={"Idempotency-Key": unique_key}
                )
                resp.raise_for_status()
                return ChargeModel(**resp.json())
            except httpx.TimeoutException:
                if attempt == max_retries - 1:
                    raise
                continue  # Retry without idempotency key → duplicate charge


# Network timeout after charge created but before response received
# Retry creates second charge → customer billed twice
```

**Why Stricture catches this:** `CTR-request-shape` validates that mutation requests (POST, PATCH, DELETE) include idempotency keys when retrying. Stripe's `Idempotency-Key` header prevents duplicate charges if a request times out after the charge was created but before the response was received. Without this header, retries create multiple charges, double-billing customers.
