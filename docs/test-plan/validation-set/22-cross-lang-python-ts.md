# 22 — Cross-Language Contract: Python Server (FastAPI) + TypeScript Client

**Status:** Validation Set
**Languages:** Python (FastAPI/Pydantic) + TypeScript (fetch client)
**Why included:** Tests Contract-to-Reality (CTR) rules across language boundary with inherent naming convention conflicts (snake_case ↔ camelCase). FastAPI/Pydantic server with TypeScript fetch client represents common full-stack architecture with built-in mismatch sources.

---

## Overview

### API: Product Catalog Service

**Endpoints:**
- `POST /api/products` — Create product
- `GET /api/products/:id` — Get product by ID
- `GET /api/products` — List products (paginated)
- `PUT /api/products/:id` — Update product

**Key Cross-Language Challenges:**
1. **Naming conventions:** Python `snake_case` ↔ TypeScript `camelCase`
2. **Type systems:** Pydantic validation ↔ TypeScript interfaces
3. **Numeric types:** Python `Decimal` ↔ TypeScript `number`
4. **Nullability:** Python `Optional[T]` ↔ TypeScript `T | null | undefined`
5. **Enums:** Python Enum classes ↔ TypeScript string unions
6. **Validation:** Server-side Pydantic validators ↔ Client-side runtime checks
7. **Serialization:** FastAPI JSON encoding ↔ fetch() JSON parsing

---

## Architecture

### Server (Python)
```
server/
├── main.py                    # FastAPI app, CORS, exception handlers
├── models.py                  # Pydantic models (Product, CreateProductRequest, etc.)
├── routers/
│   └── products.py            # Product endpoints
├── tests/
│   ├── test_products.py       # Pytest tests for all endpoints
│   └── test_models.py         # Pydantic model validation tests
└── requirements.txt           # fastapi, pydantic, uvicorn, pytest
```

### Client (TypeScript)
```
client/
├── src/
│   ├── types.ts               # TypeScript interfaces matching Pydantic models
│   ├── product-client.ts      # API client class with fetch() calls
│   └── config.ts              # API base URL, headers
├── tests/
│   ├── product-client.test.ts # Jest tests for all operations
│   └── types.test.ts          # Type validation tests
├── package.json
└── tsconfig.json              # strict: true
```

---

## Manifest Fragment

### Directory Structure
```
22-cross-lang-python-ts/
├── server/
│   ├── main.py
│   ├── models.py
│   ├── routers/
│   │   └── products.py
│   ├── tests/
│   │   ├── test_products.py
│   │   └── test_models.py
│   └── requirements.txt
├── client/
│   ├── src/
│   │   ├── types.ts
│   │   ├── product-client.ts
│   │   └── config.ts
│   ├── tests/
│   │   ├── product-client.test.ts
│   │   └── types.test.ts
│   ├── package.json
│   └── tsconfig.json
└── README.md
```

### Test Inventory

**Python (server/) — 38 tests:**
- `test_models.py` — 12 tests (Pydantic validation)
- `test_products.py` — 26 tests (endpoint behavior, error handling, pagination)

**TypeScript (client/) — 32 tests:**
- `types.test.ts` — 8 tests (interface validation, type guards)
- `product-client.test.ts` — 24 tests (API operations, error handling, pagination)

**Total:** 70 tests across both languages

---

## PERFECT Implementation

### Server: models.py (Pydantic)
```python
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
```

### Server: routers/products.py
```python
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
```

### Server: main.py
```python
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
```

---

## BUG Cases (B01–B15)

### B01 — No Error Handling (TQ-error-path-coverage)

**Bug:** TypeScript client makes fetch() call without try/catch, crashes on network errors.
**Expected violation:** `TQ-error-path-coverage`

**Python Server:**
```python
@router.get("/{product_id}", response_model=Product)
async def get_product(product_id: UUID) -> Product:
    """Get a product by ID."""
    if product_id not in products_db:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Product with ID {product_id} not found",
        )
    return products_db[product_id]
```

**TypeScript Client:**
```typescript
// BUG: No try/catch — network errors crash the app
async function getProduct(productId: string): Promise<Product> {
    const response = await fetch(`${API_BASE}/api/products/${productId}`);
    const data = await response.json();
    return data;
}
```

**Why Stricture catches this:** `TQ-error-path-coverage` requires every I/O operation (fetch, file access, network) to have explicit error handling. This code has zero error paths tested, violating the rule.

---

### B02 — No Status Code Check (CTR-status-code-handling)

**Bug:** TypeScript client doesn't check `response.ok`, treats 404 as success and returns error JSON as if it were a Product.
**Expected violation:** `CTR-status-code-handling`

**Python Server:**
```python
@router.get("/{product_id}", response_model=Product)
async def get_product(product_id: UUID) -> Product:
    if product_id not in products_db:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Product with ID {product_id} not found",
        )
    return products_db[product_id]
```

**TypeScript Client:**
```typescript
// BUG: No response.ok check — 404 error treated as valid Product
async function getProduct(productId: string): Promise<Product> {
    try {
        const response = await fetch(`${API_BASE}/api/products/${productId}`);
        // Missing: if (!response.ok) { throw new Error(...) }
        const data = await response.json();
        return data;  // Could be {detail: "not found"} instead of Product
    } catch (error) {
        throw new Error(`Failed to fetch product: ${error.message}`);
    }
}
```

**Why Stricture catches this:** `CTR-status-code-handling` requires HTTP clients to validate `response.ok` (or status codes) before parsing JSON. Server spec says 404 returns error shape, not Product shape.

---

### B03 — Shallow Test Assertions (TQ-no-shallow-assertions)

**Bug:** Test only checks `expect(user).toBeDefined()` instead of validating all required fields match the contract.
**Expected violation:** `TQ-no-shallow-assertions`

**Python Server:**
```python
class Product(ProductBase):
    id: UUID = Field(default_factory=uuid4)
    created_at: datetime = Field(default_factory=datetime.utcnow)
    updated_at: datetime = Field(default_factory=datetime.utcnow)
    version: int = Field(default=1)
```

**TypeScript Client Test:**
```typescript
// BUG: Shallow assertion — doesn't verify contract fields
test("getProduct returns product", async () => {
    const product = await client.getProduct("123e4567-e89b-12d3-a456-426614174000");
    expect(product).toBeDefined();  // SHALLOW — what about id, created_at, version?
});
```

**Why Stricture catches this:** `TQ-no-shallow-assertions` requires deep validation of contract fields. The Product schema has 13+ fields (id, product_name, price, created_at, etc.) — test must verify all required fields exist and have correct types.

---

### B04 — Missing Negative Tests (TQ-negative-cases)

**Bug:** No tests for 422 validation errors, only happy path tests exist.
**Expected violation:** `TQ-negative-cases`

**Python Server:**
```python
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
```

**TypeScript Client Test:**
```typescript
// BUG: Only tests happy path — no 422 validation error tests
describe("createProduct", () => {
    test("creates product successfully", async () => {
        const product = await client.createProduct({
            product_name: "Widget",
            price: 19.99,
            stock_quantity: 100,
            category: "Tools",
        });
        expect(product.id).toBeDefined();
    });

    // MISSING: test("returns 422 for invalid price", ...)
    // MISSING: test("returns 422 for empty product_name", ...)
    // MISSING: test("returns 422 for negative stock_quantity", ...)
});
```

**Why Stricture catches this:** `TQ-negative-cases` requires at least one negative test per error condition in the API contract. Server spec lists 422 validation error — tests must verify client handles it.

---

### B05 — Request Missing Required Fields (CTR-request-shape)

**Bug:** TypeScript client sends request body missing required field that FastAPI validates.
**Expected violation:** `CTR-request-shape`

**Python Server:**
```python
class CreateProductRequest(ProductBase):
    product_name: str = Field(..., min_length=1, max_length=200)  # REQUIRED
    price: Decimal = Field(..., ge=0, max_digits=10, decimal_places=2)  # REQUIRED
    stock_quantity: int = Field(..., ge=0)  # REQUIRED
    category: str = Field(..., min_length=1, max_length=100)  # REQUIRED
```

**TypeScript Client:**
```typescript
// BUG: Missing required field "category"
interface CreateProductRequest {
    product_name: string;
    price: number;
    stock_quantity: number;
    // MISSING: category (required by Pydantic)
}

async function createProduct(req: CreateProductRequest): Promise<Product> {
    const response = await fetch(`${API_BASE}/api/products`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(req),  // Server will return 422 — category missing
    });
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    return response.json();
}
```

**Why Stricture catches this:** `CTR-request-shape` requires all required fields in server schema to be present in client request type. Pydantic requires `category`, TypeScript interface omits it.

---

### B06 — Response Type Mismatch (CTR-response-shape)

**Bug:** TypeScript interface declares field that Pydantic model doesn't include in response.
**Expected violation:** `CTR-response-shape`

**Python Server:**
```python
class Product(ProductBase):
    id: UUID = Field(default_factory=uuid4)
    created_at: datetime = Field(default_factory=datetime.utcnow)
    updated_at: datetime = Field(default_factory=datetime.utcnow)
    version: int = Field(default=1)
    # NOTE: No "sku" field in Pydantic model
```

**TypeScript Client:**
```typescript
// BUG: "sku" field doesn't exist in Python response
interface Product {
    id: string;
    product_name: string;
    price: number;
    stock_quantity: number;
    category: string;
    created_at: string;
    updated_at: string;
    version: number;
    sku: string;  // NOT in Pydantic model — will be undefined at runtime
}

async function getProduct(productId: string): Promise<Product> {
    const response = await fetch(`${API_BASE}/api/products/${productId}`);
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    return response.json();  // Runtime data missing "sku"
}
```

**Why Stricture catches this:** `CTR-response-shape` requires client response types to be a subset of server response schema. TypeScript expects `sku`, Pydantic doesn't provide it — contract violation.

---

### B07 — Wrong Field Types (CTR-manifest-conformance)

**Bug:** FastAPI returns ISO datetime string, TypeScript treats as Date object without parsing.
**Expected violation:** `CTR-manifest-conformance`

**Python Server:**
```python
class Product(ProductBase):
    created_at: datetime = Field(default_factory=datetime.utcnow)

    class Config:
        json_encoders = {
            datetime: lambda v: v.isoformat(),  # Returns "2024-01-15T12:34:56.789Z"
        }
```

**TypeScript Client:**
```typescript
// BUG: created_at is Date type but server sends string
interface Product {
    id: string;
    product_name: string;
    created_at: Date;  // WRONG — server sends ISO string, not Date object
}

async function getProduct(productId: string): Promise<Product> {
    const response = await fetch(`${API_BASE}/api/products/${productId}`);
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    const data = await response.json();
    return data;  // created_at is "2024-01-15T12:34:56.789Z" (string), not Date
}

// Usage:
const product = await getProduct("...");
console.log(product.created_at.getFullYear());  // CRASH — created_at is string
```

**Why Stricture catches this:** `CTR-manifest-conformance` requires client types to match server JSON wire format. Server sends ISO string, client expects Date object — type mismatch.

---

### B08 — Incomplete Enum Handling (CTR-strictness-parity)

**Bug:** TypeScript Literal type missing enum values from Python Enum.
**Expected violation:** `CTR-strictness-parity`

**Python Server:**
```python
class ProductStatus(str, Enum):
    AVAILABLE = "available"
    OUT_OF_STOCK = "out_of_stock"
    DISCONTINUED = "discontinued"
    PRE_ORDER = "pre_order"  # 4 values
```

**TypeScript Client:**
```typescript
// BUG: Missing "pre_order" enum value
type ProductStatus = "available" | "out_of_stock" | "discontinued";
// MISSING: "pre_order"

interface Product {
    status: ProductStatus;
}

async function filterByStatus(status: ProductStatus): Promise<Product[]> {
    const response = await fetch(`${API_BASE}/api/products?status=${status}`);
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    const data = await response.json();
    return data.items;
}

// Usage:
const products = await filterByStatus("pre_order");
// TypeScript error: "pre_order" not assignable to ProductStatus
// But server DOES accept "pre_order" — client/server enum mismatch
```

**Why Stricture catches this:** `CTR-strictness-parity` requires client enums to exactly match server enum values. Python has 4 values, TypeScript has 3 — incomplete coverage.

---

### B09 — Missing Range Validation (CTR-strictness-parity)

**Bug:** TypeScript doesn't enforce Pydantic Field range constraints (ge, le).
**Expected violation:** `CTR-strictness-parity`

**Python Server:**
```python
class ProductBase(BaseModel):
    discount_percentage: Optional[Decimal] = Field(
        None, ge=0, le=100, max_digits=5, decimal_places=2
    )  # Server enforces 0-100 range
```

**TypeScript Client:**
```typescript
// BUG: No runtime validation for discount_percentage range
interface CreateProductRequest {
    product_name: string;
    price: number;
    stock_quantity: number;
    category: string;
    discount_percentage?: number;  // No ge=0, le=100 constraint
}

async function createProduct(req: CreateProductRequest): Promise<Product> {
    // BUG: Sends invalid discount_percentage without client-side validation
    const response = await fetch(`${API_BASE}/api/products`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(req),  // Could send discount_percentage: 150
    });
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    return response.json();
}

// Usage:
await createProduct({
    product_name: "Widget",
    price: 99.99,
    stock_quantity: 50,
    category: "Tools",
    discount_percentage: 150,  // Server will reject with 422 — should fail client-side
});
```

**Why Stricture catches this:** `CTR-strictness-parity` requires client validation to mirror server constraints. Pydantic enforces `ge=0, le=100`, TypeScript has no validation — strictness mismatch.

---

### B10 — Format Not Validated (CTR-strictness-parity)

**Bug:** TypeScript accepts any string for email field, doesn't enforce EmailStr validation.
**Expected violation:** `CTR-strictness-parity`

**Python Server:**
```python
from pydantic import BaseModel, EmailStr

class ContactInfo(BaseModel):
    email: EmailStr  # Pydantic validates email format
    phone: str
```

**TypeScript Client:**
```typescript
// BUG: No email format validation
interface ContactInfo {
    email: string;  // Should validate email format like Pydantic
    phone: string;
}

async function updateContact(productId: string, contact: ContactInfo): Promise<void> {
    const response = await fetch(`${API_BASE}/api/products/${productId}/contact`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(contact),
    });
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
}

// Usage:
await updateContact("123", {
    email: "not-an-email",  // Invalid format — server will reject, client allows
    phone: "555-1234",
});
```

**Why Stricture catches this:** `CTR-strictness-parity` requires client to enforce same format constraints as server. Pydantic validates EmailStr, TypeScript treats as plain string.

---

### B11 — snake_case Mismatch (CTR-json-tag-match)

**Bug:** Python uses `created_at` (snake_case), TypeScript uses `createdAt` (camelCase) without Pydantic alias.
**Expected violation:** `CTR-json-tag-match`

**Python Server:**
```python
class Product(ProductBase):
    id: UUID = Field(default_factory=uuid4)
    created_at: datetime = Field(default_factory=datetime.utcnow)  # snake_case
    updated_at: datetime = Field(default_factory=datetime.utcnow)

    class Config:
        json_encoders = {
            datetime: lambda v: v.isoformat(),
        }
    # NO alias configured — JSON uses "created_at"
```

**TypeScript Client:**
```typescript
// BUG: camelCase field names don't match server JSON (snake_case)
interface Product {
    id: string;
    productName: string;  // Server sends "product_name"
    createdAt: string;    // Server sends "created_at"
    updatedAt: string;    // Server sends "updated_at"
}

async function getProduct(productId: string): Promise<Product> {
    const response = await fetch(`${API_BASE}/api/products/${productId}`);
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    const data = await response.json();
    return data;  // Runtime: {created_at: "...", product_name: "..."} — keys don't match
}

// Usage:
const product = await getProduct("123");
console.log(product.createdAt);  // undefined — actual key is "created_at"
```

**Why Stricture catches this:** `CTR-json-tag-match` requires client JSON field names to exactly match server JSON keys. Server sends `created_at`, client expects `createdAt` — naming mismatch.

---

### B12 — Nullable Field Crashes (CTR-response-shape)

**Bug:** Python Optional[str] = None, TypeScript doesn't handle null, crashes on access.
**Expected violation:** `CTR-response-shape`

**Python Server:**
```python
class ProductBase(BaseModel):
    description: Optional[str] = Field(None, max_length=2000)  # Can be None
```

**TypeScript Client:**
```typescript
// BUG: description is string, not string | null
interface Product {
    id: string;
    product_name: string;
    description: string;  // WRONG — should be "string | null"
}

async function getProduct(productId: string): Promise<Product> {
    const response = await fetch(`${API_BASE}/api/products/${productId}`);
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    return response.json();
}

// Usage:
const product = await getProduct("123");
console.log(product.description.toUpperCase());  // CRASH if description is null
```

**Why Stricture catches this:** `CTR-response-shape` requires client to handle all nullable fields from server schema. Pydantic allows `None`, TypeScript doesn't mark as nullable.

---

### B13 — Missing Auth (CTR-request-shape)

**Bug:** TypeScript doesn't send Bearer token that FastAPI Depends requires.
**Expected violation:** `CTR-request-shape`

**Python Server:**
```python
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
```

**TypeScript Client:**
```typescript
// BUG: No Authorization header — server requires Bearer token
async function createProduct(req: CreateProductRequest): Promise<Product> {
    const response = await fetch(`${API_BASE}/api/products`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            // MISSING: "Authorization": "Bearer valid-token-123"
        },
        body: JSON.stringify(req),
    });
    if (!response.ok) throw new Error(`HTTP ${response.status}`);  // Will get 401
    return response.json();
}
```

**Why Stricture catches this:** `CTR-request-shape` requires client to include all required headers from server contract. FastAPI Depends(security) requires Authorization header, client omits it.

---

### B14 — Pagination Mismatch (CTR-response-shape)

**Bug:** Python returns `total`, TypeScript expects `totalCount` — field name mismatch.
**Expected violation:** `CTR-response-shape`

**Python Server:**
```python
class PaginatedProducts(BaseModel):
    items: list[Product]
    total: int          # Snake_case field
    offset: int
    limit: int
    has_more: bool
```

**TypeScript Client:**
```typescript
// BUG: Field name mismatch (total vs totalCount)
interface PaginatedProducts {
    items: Product[];
    totalCount: number;  // Server sends "total", not "totalCount"
    offset: number;
    limit: number;
    hasMore: boolean;    // Server sends "has_more", not "hasMore"
}

async function listProducts(offset: number, limit: number): Promise<PaginatedProducts> {
    const response = await fetch(`${API_BASE}/api/products?offset=${offset}&limit=${limit}`);
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
    return response.json();  // Runtime: {total: 42, has_more: true}
}

// Usage:
const result = await listProducts(0, 20);
console.log(`Total: ${result.totalCount}`);  // undefined — actual key is "total"
```

**Why Stricture catches this:** `CTR-response-shape` requires exact JSON key match. Server uses `total` and `has_more`, client expects `totalCount` and `hasMore`.

---

### B15 — Race Condition (CTR-request-shape)

**Bug:** No If-Match header for optimistic locking, client bypasses version check.
**Expected violation:** `CTR-request-shape`

**Python Server:**
```python
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
```

**TypeScript Client:**
```typescript
// BUG: version sent in body, but should use If-Match header per REST best practices
interface UpdateProductRequest {
    product_name?: string;
    price?: number;
    version: number;  // Should be If-Match header, not body field
}

async function updateProduct(
    productId: string,
    updates: UpdateProductRequest
): Promise<Product> {
    const response = await fetch(`${API_BASE}/api/products/${productId}`, {
        method: "PUT",
        headers: {
            "Content-Type": "application/json",
            // MISSING: "If-Match": String(updates.version)
        },
        body: JSON.stringify(updates),  // version in body works, but not idiomatic
    });
    if (!response.ok) {
        if (response.status === 409) {
            throw new Error("Concurrent modification — retry with latest version");
        }
        throw new Error(`HTTP ${response.status}`);
    }
    return response.json();
}
```

**Why Stricture catches this:** `CTR-request-shape` requires client to follow server's concurrency control pattern. While version-in-body works, best practice is If-Match header for optimistic locking. Contract mismatch if server expects header but client sends body field (or vice versa).

---

## Test Manifest Summary

**Bug Distribution:**
- TQ (Test Quality): B01, B03, B04 (3 bugs)
- CTR (Contract-to-Reality): B02, B05, B06, B07, B08, B09, B10, B11, B12, B13, B14, B15 (12 bugs)

**Cross-Language Challenges Covered:**
1. **Naming conventions** (B11, B14) — snake_case vs camelCase
2. **Type systems** (B06, B07, B12) — Pydantic vs TypeScript
3. **Validation parity** (B09, B10) — Field constraints
4. **Enum completeness** (B08) — Python Enum vs TS union
5. **Nullability** (B12) — Optional[T] vs T | null
6. **Request contracts** (B05, B13, B15) — Required fields, headers, versioning
7. **Error handling** (B01, B02) — Network failures, HTTP status codes
8. **Test coverage** (B03, B04) — Shallow assertions, missing negative tests

**Why this validates Stricture:**
- Every bug represents a REAL production failure mode in Python↔TypeScript integration
- Automatic snake_case/camelCase issues are the #1 cross-language bug source
- Stricture must detect these WITHOUT requiring developers to manually write contract validators
- Tests prove Stricture's CTR rules catch API contract violations before runtime

