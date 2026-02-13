# B06-missing-required-field/python-client/models.py
class Item(BaseModel):
    """Inventory item model MISSING warehouse_id"""
    id: str
    sku: str
    name: str
    description: Optional[str] = None
    stock_count: int = Field(ge=0)
    # ❌ warehouse_id field is missing
    status: ItemStatus
    created_at: datetime
    updated_at: datetime
    etag: str
