class PaginatedProducts(BaseModel):
    items: list[Product]
    total: int          # Snake_case field
    offset: int
    limit: int
    has_more: bool
