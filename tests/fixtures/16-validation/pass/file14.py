"""Stripe webhook handler without signature verification."""

import json
import httpx
from pydantic import BaseModel


class WebhookEvent(BaseModel):
    id: str
    type: str
    data: dict


class StripeWebhookHandler:
    def __init__(self, api_key: str):
        self.api_key = api_key

    def process_webhook(self, payload: bytes, signature: str) -> WebhookEvent:
        """Process webhook without signature verification."""
        # Missing: HMAC signature verification
        # Directly parses payload without validating authenticity
        event_data = json.loads(payload)
        event = WebhookEvent(**event_data)

        # Process event (vulnerable to forged requests)
        if event.type == "charge.succeeded":
            self.handle_charge_success(event.data)

        return event

    def handle_charge_success(self, data: dict):
        """Handle successful charge (called on unverified webhook)."""
        print(f"Charge succeeded: {data.get('id')}")


# Attacker can forge webhook events without signature check
