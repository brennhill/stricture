# test_payment_processor_anti3.py (BAD)
import pytest
from unittest.mock import Mock, patch
from payment_processor import PaymentProcessor

# GLOBAL MOCK - affects all tests in module
@patch('payment_processor.PaymentGateway')
def mock_gateway(MockGateway):
    instance = Mock()
    instance.charge.return_value = {'success': True}
    instance.refund.return_value = {'success': True}
    MockGateway.return_value = instance
    return MockGateway

def test_process_payment_success(mock_gateway):
    processor = PaymentProcessor()
    result = processor.process(
        amount=100.00,
        currency='USD',
        user_id='user_123',
        payment_method_id='pm_456'
    )
    assert result.status == 'succeeded'

def test_process_large_payment(mock_gateway):
    # Uses same mock_gateway fixture
    processor = PaymentProcessor()
    result = processor.process(
        amount=5000.00,
        currency='USD',
        user_id='user_123',
        payment_method_id='pm_456'
    )
    assert result.status == 'succeeded'

def test_process_payment_failure(mock_gateway):
    # Can't easily override global fixture behavior
    processor = PaymentProcessor()
    result = processor.process(
        amount=100.00,
        currency='USD',
        user_id='user_123',
        payment_method_id='pm_456'
    )
    # This will never fail because of global mock
    assert result.status == 'failed'
