class ProductBase(BaseModel):
    discount_percentage: Optional[Decimal] = Field(
        None, ge=0, le=100, max_digits=5, decimal_places=2
    )  # Server enforces 0-100 range
