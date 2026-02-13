"""models.py — Pydantic models matching Go server schema"""
from datetime import datetime
from enum import Enum
from typing import Optional

from pydantic import BaseModel, Field


class ItemStatus(str, Enum):
    """Item availability status enum"""
    IN_STOCK = "in_stock"
    LOW_STOCK = "low_stock"
    OUT_OF_STOCK = "out_of_stock"
    DISCONTINUED = "discontinued"
    BACKORDER = "backorder"


class Item(BaseModel):
    """Inventory item model"""
    id: str
    sku: str
    name: str
    description: Optional[str] = None
    stock_count: int = Field(ge=0)  # Non-negative integer
    warehouse_id: str
    status: ItemStatus
    created_at: datetime
    updated_at: datetime
    etag: str

    class Config:
        use_enum_values = True


class CreateItemRequest(BaseModel):
    """Request model for creating a new item"""
    sku: str
    name: str
    description: Optional[str] = None
    stock_count: int = Field(ge=0)
    warehouse_id: str


class UpdateItemRequest(BaseModel):
    """Request model for updating an item"""
    stock_count: Optional[int] = Field(None, ge=0)
    status: Optional[ItemStatus] = None
    description: Optional[str] = None

    class Config:
        use_enum_values = True


class Pagination(BaseModel):
    """Pagination metadata"""
    page: int
    limit: int
    total: int
    has_next: bool


class ListItemsResponse(BaseModel):
    """Response model for listing items"""
    items: list[Item]
    pagination: Pagination


class ErrorResponse(BaseModel):
    """Error response model"""
    error: str
    message: str
