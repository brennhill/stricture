def update_item(
    self,
    item_id: str,
    etag: str,  # Add ETag parameter
    request: UpdateItemRequest,
) -> Item:
    response = self.client.patch(
        f"/items/{item_id}",
        json=request.model_dump(exclude_none=True),
        headers={"If-Match": f'"{etag}"'},  # ✅ Include ETag
    )
    response.raise_for_status()
    return Item(**response.json())
