@router.get("/{product_id}", response_model=Product)
async def get_product(product_id: UUID) -> Product:
    """Get a product by ID."""
    if product_id not in products_db:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Product with ID {product_id} not found",
        )
    return products_db[product_id]
