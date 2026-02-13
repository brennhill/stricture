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
