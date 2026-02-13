from fastapi import FastAPI, Request, status
from fastapi.responses import JSONResponse
from fastapi.middleware.cors import CORSMiddleware
from pydantic import ValidationError
from routers import products

app = FastAPI(
    title="Product Catalog API",
    version="1.0.0",
    description="Cross-language test: Python FastAPI + TypeScript client",
)

# CORS configuration for TypeScript client
app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:3000"],  # TypeScript dev server
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include routers
app.include_router(products.router)


@app.exception_handler(ValidationError)
async def validation_exception_handler(
    request: Request, exc: ValidationError
) -> JSONResponse:
    """Handle Pydantic validation errors."""
    return JSONResponse(
        status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
        content={
            "detail": exc.errors(),
            "body": exc.model.__name__ if hasattr(exc, "model") else None,
        },
    )


@app.get("/health")
async def health_check():
    """Health check endpoint."""
    return {"status": "healthy"}
