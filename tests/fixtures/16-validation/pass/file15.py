"""Stripe client with incomplete pagination."""

import httpx
from typing import List
from pydantic import BaseModel


class ChargeModel(BaseModel):
    id: str
    amount: int
    currency: str
    status: str


class ChargeListResponse(BaseModel):
    data: List[ChargeModel]
    has_more: bool
    url: str


class StripeClient:
    def __init__(self, api_key: str):
        self.client = httpx.Client(
            base_url="https://api.stripe.com",
            headers={"Authorization": f"Bearer {api_key}"},
        )

    def list_charges(self, limit: int = 10) -> List[ChargeModel]:
        """List charges with broken pagination."""
        try:
            resp = self.client.get("/v1/charges", params={"limit": limit})
            resp.raise_for_status()
            page = ChargeListResponse(**resp.json())

            # Returns only first page, ignoring has_more
            return page.data
        except httpx.HTTPStatusError as e:
            raise ValueError(f"Failed to list charges: {e}")


# If account has 100 charges but limit=10, only returns 10 instead of 100
