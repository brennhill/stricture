from typing import Optional
from uuid import UUID
from fastapi import APIRouter, HTTPException, status, Query
from models import (
    Product,
    CreateProductRequest,
    UpdateProductRequest,
    PaginatedProducts,
    ProductStatus,
)

router = APIRouter(prefix="/api/products", tags=["products"])

# In-memory storage (replace with database in production)
products_db: dict[UUID, Product] = {}


@router.post(
    "",
    response_model=Product,
    status_code=status.HTTP_201_CREATED,
    responses={
        201: {"description": "Product created successfully"},
        400: {"description": "Invalid request body"},
        422: {"description": "Validation error"},
    },
)
async def create_product(request: CreateProductRequest) -> Product:
    """Create a new product."""
    product = Product(**request.model_dump())
    products_db[product.id] = product
    return product


@router.get(
    "/{product_id}",
    response_model=Product,
    responses={
        200: {"description": "Product found"},
        404: {"description": "Product not found"},
        422: {"description": "Invalid product ID format"},
    },
)
async def get_product(product_id: UUID) -> Product:
    """Get a product by ID."""
    if product_id not in products_db:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Product with ID {product_id} not found",
        )
    return products_db[product_id]


@router.get(
    "",
    response_model=PaginatedProducts,
    responses={
        200: {"description": "Products retrieved successfully"},
        400: {"description": "Invalid pagination parameters"},
    },
)
async def list_products(
    offset: int = Query(0, ge=0, description="Number of items to skip"),
    limit: int = Query(20, ge=1, le=100, description="Number of items to return"),
    status_filter: Optional[ProductStatus] = Query(
        None, alias="status", description="Filter by product status"
    ),
    category: Optional[str] = Query(None, description="Filter by category"),
) -> PaginatedProducts:
    """List products with pagination and optional filters."""
    # Apply filters
    filtered_products = list(products_db.values())

    if status_filter:
        filtered_products = [
            p for p in filtered_products if p.status == status_filter
        ]

    if category:
        filtered_products = [
            p for p in filtered_products if p.category.lower() == category.lower()
        ]

    # Sort by created_at descending
    filtered_products.sort(key=lambda p: p.created_at, reverse=True)

    total = len(filtered_products)
    items = filtered_products[offset : offset + limit]
    has_more = offset + limit < total

    return PaginatedProducts(
        items=items,
        total=total,
        offset=offset,
        limit=limit,
        has_more=has_more,
    )


@router.put(
    "/{product_id}",
    response_model=Product,
    responses={
        200: {"description": "Product updated successfully"},
        404: {"description": "Product not found"},
        409: {"description": "Version conflict (optimistic locking)"},
        422: {"description": "Validation error"},
    },
)
async def update_product(
    product_id: UUID, request: UpdateProductRequest
) -> Product:
    """Update a product with optimistic locking."""
    if product_id not in products_db:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Product with ID {product_id} not found",
        )

    existing_product = products_db[product_id]

    # Optimistic locking check
    if existing_product.version != request.version:
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail=(
                f"Version conflict: expected {existing_product.version}, "
                f"got {request.version}"
            ),
        )

    # Apply partial update
    update_data = request.model_dump(exclude_unset=True, exclude={"version"})
    for field, value in update_data.items():
        setattr(existing_product, field, value)

    # Increment version and update timestamp
    existing_product.version += 1
    existing_product.updated_at = datetime.utcnow()

    return existing_product
