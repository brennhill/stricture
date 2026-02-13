"""client.py — Httpx client for inventory API"""
import httpx
from typing import Optional

from models import (
    Item,
    CreateItemRequest,
    UpdateItemRequest,
    ListItemsResponse,
    ErrorResponse,
    ItemStatus,
)


class InventoryClient:
    """HTTP client for inventory management API"""

    def __init__(self, base_url: str = "http://localhost:8080/api/v1"):
        self.base_url = base_url
        self.client = httpx.Client(base_url=base_url, timeout=30.0)

    def close(self):
        """Close the HTTP client"""
        self.client.close()

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.close()

    def list_items(
        self,
        page: int = 1,
        limit: int = 20,
        status: Optional[ItemStatus] = None,
    ) -> ListItemsResponse:
        """List items with pagination and optional status filter"""
        params = {"page": page, "limit": limit}
        if status:
            params["status"] = status.value

        response = self.client.get("/items", params=params)
        response.raise_for_status()

        return ListItemsResponse(**response.json())

    def get_item(self, item_id: str) -> Item:
        """Get a single item by ID"""
        response = self.client.get(f"/items/{item_id}")
        response.raise_for_status()

        return Item(**response.json())

    def create_item(self, request: CreateItemRequest) -> Item:
        """Create a new item"""
        response = self.client.post(
            "/items",
            json=request.model_dump(exclude_none=True),
        )
        response.raise_for_status()

        return Item(**response.json())

    def update_item(
        self,
        item_id: str,
        etag: str,
        request: UpdateItemRequest,
    ) -> Item:
        """Update an existing item (requires ETag)"""
        response = self.client.patch(
            f"/items/{item_id}",
            json=request.model_dump(exclude_none=True),
            headers={"If-Match": f'"{etag}"'},
        )
        response.raise_for_status()

        return Item(**response.json())

    def delete_item(self, item_id: str) -> None:
        """Delete an item"""
        response = self.client.delete(f"/items/{item_id}")
        response.raise_for_status()

    def handle_error(self, response: httpx.Response) -> ErrorResponse:
        """Parse error response"""
        try:
            return ErrorResponse(**response.json())
        except Exception:
            return ErrorResponse(
                error="UNKNOWN_ERROR",
                message=f"HTTP {response.status_code}: {response.text}",
            )
