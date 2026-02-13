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
