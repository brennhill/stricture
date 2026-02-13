# B05-field-naming-mismatch/python-client/models.py
class Item(BaseModel):
    """Inventory item model with WRONG field naming"""
    id: str
    sku: str
    name: str
    description: Optional[str] = None
    stockCount: int = Field(ge=0)  # ❌ Should be stock_count
    warehouse_id: str
    status: ItemStatus
    created_at: datetime
    updated_at: datetime
    etag: str
