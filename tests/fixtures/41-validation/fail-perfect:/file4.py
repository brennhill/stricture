# test_payment_processor_anti1.py (BAD)
import pytest
from payment_processor import PaymentProcessor, PaymentRequest

def test_process_payment_carpet_bomb():
    processor = PaymentProcessor()
    result = processor.process(PaymentRequest(
        amount=100.00,
        currency='USD',
        user_id='user_123',
        payment_method_id='pm_456'
    ))

    assert result is not None
    assert result.transaction_id is not None
    assert result.amount is not None
    assert result.currency is not None
    assert result.status is not None
    assert result.fee is not None
    assert result.net_amount is not None
    assert result.processed_at is not None
    assert result.metadata is not None

def test_refund_defined():
    processor = PaymentProcessor()
    result = processor.refund('txn_123', 50.00)

    assert result is not None
    assert result.transaction_id is not None
