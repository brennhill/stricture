from fastapi import APIRouter, Depends, HTTPException, status
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials

security = HTTPBearer()

@router.post("", response_model=Product, status_code=status.HTTP_201_CREATED,
    responses={
        201: {"description": "Product created successfully"},
        401: {"description": "Unauthorized — missing or invalid token"},
    }
)
async def create_product(
    request: CreateProductRequest,
    credentials: HTTPAuthorizationCredentials = Depends(security),  # REQUIRES Bearer token
) -> Product:
    if credentials.credentials != "valid-token-123":
        raise HTTPException(status_code=status.HTTP_401_UNAUTHORIZED)
    product = Product(**request.model_dump())
    products_db[product.id] = product
    return product
