# B07-type-mismatch-int-float/python-client/models.py
class Item(BaseModel):
    """Inventory item with WRONG type for stock_count"""
    id: str
    sku: str
    name: str
    description: Optional[str] = None
    stock_count: float = Field(ge=0)  # ❌ Should be int to match Go int64
    warehouse_id: str
    status: ItemStatus
    created_at: datetime
    updated_at: datetime
    etag: str
