# B08-incomplete-enum/python-client/models.py
class ItemStatus(str, Enum):
    """INCOMPLETE item status enum"""
    IN_STOCK = "in_stock"
    LOW_STOCK = "low_stock"
    OUT_OF_STOCK = "out_of_stock"
    DISCONTINUED = "discontinued"
    # ❌ Missing BACKORDER = "backorder"
