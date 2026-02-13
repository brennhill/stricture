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
