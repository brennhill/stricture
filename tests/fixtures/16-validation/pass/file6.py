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
