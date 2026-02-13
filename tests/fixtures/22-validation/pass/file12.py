class CreateProductRequest(ProductBase):
    product_name: str = Field(..., min_length=1, max_length=200)  # REQUIRED
    price: Decimal = Field(..., ge=0, max_digits=10, decimal_places=2)  # REQUIRED
    stock_quantity: int = Field(..., ge=0)  # REQUIRED
    category: str = Field(..., min_length=1, max_length=100)  # REQUIRED
