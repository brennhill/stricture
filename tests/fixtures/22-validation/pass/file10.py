@router.post("", response_model=Product, status_code=status.HTTP_201_CREATED,
    responses={
        201: {"description": "Product created successfully"},
        422: {"description": "Validation error"},
    }
)
async def create_product(request: CreateProductRequest) -> Product:
    product = Product(**request.model_dump())
    products_db[product.id] = product
    return product
