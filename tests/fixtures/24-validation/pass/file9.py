"""test_client.py — Integration tests for Python client"""
import pytest
from datetime import datetime

from client import InventoryClient
from models import (
    CreateItemRequest,
    UpdateItemRequest,
    ItemStatus,
)


@pytest.fixture
def client():
    """Create client instance"""
    with InventoryClient() as c:
        yield c


def test_list_items(client):
    """Test listing items with pagination"""
    response = client.list_items(page=1, limit=10)

    assert len(response.items) > 0
    assert response.pagination.page == 1
    assert response.pagination.limit == 10
    assert response.pagination.total >= len(response.items)

    # Verify item structure
    for item in response.items:
        assert item.id
        assert item.sku
        assert item.name
        assert item.warehouse_id
        assert item.stock_count >= 0
        assert isinstance(item.status, ItemStatus)
        assert item.etag


def test_get_item_by_id(client):
    """Test getting a single item"""
    # First list to get a valid ID
    items = client.list_items(limit=1)
    item_id = items.items[0].id

    # Get specific item
    item = client.get_item(item_id)

    assert item.id == item_id
    assert item.sku
    assert item.etag


def test_get_item_not_found(client):
    """Test getting non-existent item"""
    with pytest.raises(Exception) as exc_info:
        client.get_item("nonexistent-id")

    assert exc_info.value.response.status_code == 404


def test_create_item(client):
    """Test creating a new item"""
    request = CreateItemRequest(
        sku="SKU-PYTEST-001",
        name="Python Test Item",
        description="Created by pytest",
        stock_count=100,
        warehouse_id="warehouse-test",
    )

    item = client.create_item(request)

    assert item.id
    assert item.sku == "SKU-PYTEST-001"
    assert item.name == "Python Test Item"
    assert item.description == "Created by pytest"
    assert item.stock_count == 100
    assert item.warehouse_id == "warehouse-test"
    assert item.status == ItemStatus.IN_STOCK
    assert item.etag
    assert isinstance(item.created_at, datetime)
    assert isinstance(item.updated_at, datetime)


def test_create_item_with_null_description(client):
    """Test creating item with null description"""
    request = CreateItemRequest(
        sku="SKU-PYTEST-NULL",
        name="Null Description Test",
        description=None,
        stock_count=50,
        warehouse_id="warehouse-test",
    )

    item = client.create_item(request)

    assert item.id
    assert item.description is None


def test_create_item_missing_required_field(client):
    """Test creating item without required field"""
    request = CreateItemRequest(
        sku="SKU-INCOMPLETE",
        name="Incomplete Item",
        stock_count=10,
        warehouse_id="",  # Invalid empty string
    )

    with pytest.raises(Exception) as exc_info:
        client.create_item(request)

    assert exc_info.value.response.status_code == 400


def test_update_item_with_etag(client):
    """Test updating an item with correct ETag"""
    # Create item
    create_req = CreateItemRequest(
        sku="SKU-UPDATE-TEST",
        name="Update Test",
        stock_count=100,
        warehouse_id="warehouse-test",
    )
    created = client.create_item(create_req)

    # Update item
    update_req = UpdateItemRequest(
        stock_count=150,
        status=ItemStatus.IN_STOCK,
    )
    updated = client.update_item(created.id, created.etag, update_req)

    assert updated.id == created.id
    assert updated.stock_count == 150
    assert updated.status == ItemStatus.IN_STOCK
    assert updated.etag != created.etag  # ETag should change


def test_update_item_without_etag(client):
    """Test updating item without ETag header"""
    # Create item first
    create_req = CreateItemRequest(
        sku="SKU-ETAG-TEST",
        name="ETag Test",
        stock_count=50,
        warehouse_id="warehouse-test",
    )
    created = client.create_item(create_req)

    # Try to update without ETag (by modifying client temporarily)
    update_req = UpdateItemRequest(stock_count=75)

    # This should fail with 428 Precondition Required
    with pytest.raises(Exception) as exc_info:
        # Pass empty string as etag
        client.client.patch(
            f"/items/{created.id}",
            json=update_req.model_dump(exclude_none=True),
            # Missing If-Match header
        )

    assert exc_info.value.response.status_code == 428


def test_update_item_etag_mismatch(client):
    """Test updating item with wrong ETag"""
    # Create item
    create_req = CreateItemRequest(
        sku="SKU-MISMATCH",
        name="Mismatch Test",
        stock_count=100,
        warehouse_id="warehouse-test",
    )
    created = client.create_item(create_req)

    # Update with wrong ETag
    update_req = UpdateItemRequest(stock_count=200)

    with pytest.raises(Exception) as exc_info:
        client.update_item(created.id, "wrong-etag", update_req)

    assert exc_info.value.response.status_code == 412  # Precondition Failed


def test_delete_item(client):
    """Test deleting an item"""
    # Create item
    create_req = CreateItemRequest(
        sku="SKU-DELETE",
        name="Delete Test",
        stock_count=10,
        warehouse_id="warehouse-test",
    )
    created = client.create_item(create_req)

    # Delete item
    client.delete_item(created.id)

    # Verify deletion
    with pytest.raises(Exception) as exc_info:
        client.get_item(created.id)

    assert exc_info.value.response.status_code == 404


def test_pagination(client):
    """Test pagination functionality"""
    # Get first page
    page1 = client.list_items(page=1, limit=2)

    assert len(page1.items) <= 2
    assert page1.pagination.page == 1
    assert page1.pagination.limit == 2

    # If there are more items, test second page
    if page1.pagination.has_next:
        page2 = client.list_items(page=2, limit=2)
        assert page2.pagination.page == 2
        assert page2.items[0].id != page1.items[0].id


def test_status_filter(client):
    """Test filtering by status"""
    response = client.list_items(status=ItemStatus.IN_STOCK)

    for item in response.items:
        assert item.status == ItemStatus.IN_STOCK


def test_nullable_field_handling(client):
    """Test proper handling of nullable description field"""
    # Get item with null description
    items = client.list_items(limit=10)

    # Find item with null description (item-003 from seed data)
    null_desc_items = [item for item in items.items if item.description is None]

    if null_desc_items:
        item = null_desc_items[0]
        assert item.description is None

        # Verify we can safely handle null description
        # This should NOT raise an error
        desc_upper = item.description.upper() if item.description else None
        assert desc_upper is None
