"""Test suite missing negative cases."""

import pytest
from unittest import mock
from src.stripe_client import StripeClient


@pytest.fixture
def stripe_client():
    return StripeClient(api_key="sk_test_12345")


def test_create_charge_success(stripe_client):
    """Test only the happy path."""
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
        assert charge.id == "ch_123"


# Missing tests for:
# - 402 Payment Required
# - 429 Rate Limit Exceeded
# - 400 Invalid Request
# - 401 Unauthorized
# - 500 Internal Server Error
