# B15-missing-etag/python-client/client.py
def update_item(
    self,
    item_id: str,
    request: UpdateItemRequest,
) -> Item:
    """Update item WITHOUT ETag (wrong signature)"""
    response = self.client.patch(
        f"/items/{item_id}",
        json=request.model_dump(exclude_none=True),
        # ❌ Missing If-Match header with ETag
    )
    response.raise_for_status()
    return Item(**response.json())
