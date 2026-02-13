"""Stripe model with incomplete status enum."""

import httpx
from pydantic import BaseModel, field_validator


class ChargeModel(BaseModel):
    id: str
    amount: int
    currency: str
    status: str

    @field_validator("status")
    @classmethod
    def validate_status(cls, v: str) -> str:
        """Incomplete status validation."""
        valid_statuses = {"succeeded", "pending"}  # Missing: "failed"
        if v not in valid_statuses:
            raise ValueError(f"Invalid status: {v}")
        return v


class StripeClient:
    def __init__(self, api_key: str):
        self.client = httpx.Client(
            base_url="https://api.stripe.com",
            headers={"Authorization": f"Bearer {api_key}"},
        )

    def create_charge(self, amount: int, currency: str, source: str) -> ChargeModel:
        """Create charge with incomplete status enum."""
        try:
            resp = self.client.post(
                "/v1/charges",
                data={"amount": amount, "currency": currency, "source": source},
            )
            resp.raise_for_status()
            # Will raise ValidationError if charge fails (status="failed")
            return ChargeModel(**resp.json())
        except httpx.HTTPStatusError as e:
            raise ValueError(f"Charge failed: {e}")
