"""Test with shallow assertions."""

import pytest
from unittest import mock
from src.stripe_client import StripeClient, ChargeModel


@pytest.fixture
def stripe_client():
    return StripeClient(api_key="sk_test_12345")


def test_create_charge_shallow(stripe_client):
    """Test with shallow assertion (anti-pattern)."""
    with mock.patch.object(stripe_client.client, "post") as mock_post:
        mock_resp = mock.Mock()
        mock_resp.json.return_value = {
            "id": "ch_123",
            "amount": 5000,
            "currency": "usd",
            "status": "succeeded",
        }
        mock_resp.raise_for_status = mock.Mock()
        mock_post.return_value = mock_resp

        charge = stripe_client.create_charge(
            amount=5000,
            currency="usd",
            source="tok_visa",
        )

        # Shallow assertion - doesn't verify fields
        assert charge is not None
        # Missing: assert charge.amount == 5000, etc.
