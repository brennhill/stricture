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
