# B14-pagination-mismatch/python-client/models.py
class Pagination(BaseModel):
    """WRONG pagination structure"""
    current_page: int     # ❌ Go uses "page"
    per_page: int         # ❌ Go uses "limit"
    total_items: int      # ❌ Go uses "total"
    next_page: bool       # ❌ Go uses "has_next"
