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
