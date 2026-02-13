class UpdateProductRequest(BaseModel):
    product_name: Optional[str] = None
    price: Optional[Decimal] = None
    version: int = Field(..., description="Optimistic locking version")  # REQUIRED

@router.put(
    "/{product_id}",
    response_model=Product,
    responses={
        409: {"description": "Version conflict (optimistic locking)"},
    },
)
async def update_product(product_id: UUID, request: UpdateProductRequest) -> Product:
    existing_product = products_db[product_id]

    # Optimistic locking check
    if existing_product.version != request.version:
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail=f"Version conflict: expected {existing_product.version}, got {request.version}",
        )

    # Apply update
    for field, value in request.model_dump(exclude_unset=True, exclude={"version"}).items():
        setattr(existing_product, field, value)

    existing_product.version += 1
    return existing_product
