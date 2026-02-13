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
