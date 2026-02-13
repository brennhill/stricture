from decimal import Decimal
from enum import Enum
from typing import Optional
from uuid import UUID, uuid4
from datetime import datetime
from pydantic import BaseModel, Field, field_validator


class ProductStatus(str, Enum):
    """Product availability status."""
    AVAILABLE = "available"
    OUT_OF_STOCK = "out_of_stock"
    DISCONTINUED = "discontinued"
    PRE_ORDER = "pre_order"


class ProductBase(BaseModel):
    """Base product fields for create/update operations."""
    product_name: str = Field(..., min_length=1, max_length=200)
    description: Optional[str] = Field(None, max_length=2000)
    price: Decimal = Field(..., ge=0, max_digits=10, decimal_places=2)
    discount_percentage: Optional[Decimal] = Field(
        None, ge=0, le=100, max_digits=5, decimal_places=2
    )
    status: ProductStatus = Field(default=ProductStatus.AVAILABLE)
    stock_quantity: int = Field(..., ge=0)
    category: str = Field(..., min_length=1, max_length=100)
    tags: list[str] = Field(default_factory=list)

    @field_validator("tags")
    @classmethod
    def validate_tags(cls, v: list[str]) -> list[str]:
        """Ensure all tags are non-empty strings."""
        if not all(isinstance(tag, str) and tag.strip() for tag in v):
            raise ValueError("All tags must be non-empty strings")
        return [tag.strip() for tag in v]


class CreateProductRequest(ProductBase):
    """Request body for POST /api/products."""
    pass


class UpdateProductRequest(BaseModel):
    """Request body for PUT /api/products/:id (partial updates allowed)."""
    product_name: Optional[str] = Field(None, min_length=1, max_length=200)
    description: Optional[str] = Field(None, max_length=2000)
    price: Optional[Decimal] = Field(None, ge=0, max_digits=10, decimal_places=2)
    discount_percentage: Optional[Decimal] = Field(
        None, ge=0, le=100, max_digits=5, decimal_places=2
    )
    status: Optional[ProductStatus] = None
    stock_quantity: Optional[int] = Field(None, ge=0)
    category: Optional[str] = Field(None, min_length=1, max_length=100)
    tags: Optional[list[str]] = None
    version: int = Field(..., description="Optimistic locking version")

    @field_validator("tags")
    @classmethod
    def validate_tags(cls, v: Optional[list[str]]) -> Optional[list[str]]:
        if v is None:
            return None
        if not all(isinstance(tag, str) and tag.strip() for tag in v):
            raise ValueError("All tags must be non-empty strings")
        return [tag.strip() for tag in v]


class Product(ProductBase):
    """Full product model with server-generated fields."""
    id: UUID = Field(default_factory=uuid4)
    created_at: datetime = Field(default_factory=datetime.utcnow)
    updated_at: datetime = Field(default_factory=datetime.utcnow)
    version: int = Field(default=1)

    class Config:
        json_encoders = {
            Decimal: lambda v: float(v),  # Convert Decimal to float for JSON
            UUID: lambda v: str(v),       # Convert UUID to string
            datetime: lambda v: v.isoformat(),
        }


class PaginatedProducts(BaseModel):
    """Paginated product list response."""
    items: list[Product]
    total: int
    offset: int
    limit: int
    has_more: bool
