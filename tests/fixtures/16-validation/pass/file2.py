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
