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
