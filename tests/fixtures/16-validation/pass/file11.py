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
