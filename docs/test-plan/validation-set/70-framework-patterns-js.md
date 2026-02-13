# 70 — JavaScript/TypeScript Framework Patterns

**Purpose:** Framework-specific route detection for Stricture. Shows how Stricture identifies route handlers, middleware, request parsing across 5 major JavaScript frameworks.

**Coverage:**
- Express.js (req, res, next patterns)
- Fastify (schema validation)
- NestJS (decorators, DTOs, pipes)
- Next.js API Routes (Pages Router + App Router)
- Hono (lightweight edge runtime)

**Format:** For each framework, 1 PERFECT example + 3 violations showing what Stricture catches.

---

## Express.js

### PERFECT: Validated Route with Error Handler

```typescript
// routes/users.ts
import express, { Request, Response, NextFunction } from 'express';
import { z } from 'zod';

const router = express.Router();

const CreateUserSchema = z.object({
  email: z.string().email(),
  name: z.string().min(1).max(100),
  age: z.number().int().min(13).max(120),
});

type CreateUserRequest = z.infer<typeof CreateUserSchema>;

// ✅ PERFECT: Complete validation, error handling, correct status codes
router.post('/users', async (req: Request, res: Response, next: NextFunction) => {
  try {
    // Input validation
    const parsed = CreateUserSchema.safeParse(req.body);
    if (!parsed.success) {
      res.status(400).json({
        error: 'Validation failed',
        details: parsed.error.errors,
      });
      return;
    }

    const userData: CreateUserRequest = parsed.data;

    // Business logic
    const user = await createUser(userData);

    // Correct status code for resource creation
    res.status(201).json({
      id: user.id,
      email: user.email,
      name: user.name,
      createdAt: user.createdAt,
    });
  } catch (error) {
    next(error); // Pass errors to error handler middleware
  }
});

// Error handler middleware
router.use((err: Error, req: Request, res: Response, next: NextFunction) => {
  console.error('Route error:', err);
  res.status(500).json({ error: 'Internal server error' });
});

async function createUser(data: CreateUserRequest) {
  // Simulated DB operation
  return {
    id: Math.random().toString(36).substring(7),
    ...data,
    createdAt: new Date().toISOString(),
  };
}

export default router;
```

**Why PERFECT:**
- ✅ Schema validation before processing
- ✅ safeParse with error handling (not .parse which throws)
- ✅ Correct HTTP status codes (400 for validation, 201 for creation, 500 for errors)
- ✅ try/catch with next(error) to pass errors to middleware
- ✅ Type safety with z.infer
- ✅ Error handler middleware registered

---

### VIOLATION 1: Missing Input Validation

```typescript
// routes/users-bad-1.ts
import express, { Request, Response } from 'express';

const router = express.Router();

// ❌ VIOLATION: No input validation on request body
router.post('/users', async (req: Request, res: Response) => {
  // Direct access to req.body without validation
  const { email, name, age } = req.body;

  const user = await createUser({ email, name, age });

  res.status(201).json(user);
});

async function createUser(data: any) {
  // This will fail if data is malformed
  return {
    id: Math.random().toString(36).substring(7),
    ...data,
    createdAt: new Date().toISOString(),
  };
}

export default router;
```

**Stricture Detection:**
```
ERROR [EXPRESS_MISSING_VALIDATION] routes/users-bad-1.ts:7:1
  POST route handler accesses req.body without validation

  Found: req.body destructuring without schema validation
  Expected: Schema validation (zod, joi, yup) before accessing req.body

  Fix: Add input validation:
    const parsed = CreateUserSchema.safeParse(req.body);
    if (!parsed.success) {
      return res.status(400).json({ error: parsed.error });
    }
```

---

### VIOLATION 2: Wrong Status Code

```typescript
// routes/users-bad-2.ts
import express, { Request, Response } from 'express';
import { z } from 'zod';

const router = express.Router();

const CreateUserSchema = z.object({
  email: z.string().email(),
  name: z.string().min(1),
});

// ❌ VIOLATION: Returns 200 instead of 201 for resource creation
router.post('/users', async (req: Request, res: Response) => {
  const parsed = CreateUserSchema.safeParse(req.body);
  if (!parsed.success) {
    res.status(400).json({ error: parsed.error });
    return;
  }

  const user = await createUser(parsed.data);

  // Wrong status code - should be 201 for POST creating a resource
  res.status(200).json(user);
});

async function createUser(data: any) {
  return {
    id: Math.random().toString(36).substring(7),
    ...data,
  };
}

export default router;
```

**Stricture Detection:**
```
WARNING [EXPRESS_WRONG_STATUS_CODE] routes/users-bad-2.ts:22:3
  POST route creating resource returns 200 instead of 201

  Found: res.status(200) after successful resource creation
  Expected: res.status(201) for POST routes that create resources

  Context: Route pattern '/users' with POST method suggests resource creation
  Fix: Change to res.status(201).json(user)
```

---

### VIOLATION 3: Unhandled Middleware Error

```typescript
// routes/users-bad-3.ts
import express, { Request, Response, NextFunction } from 'express';
import { z } from 'zod';

const router = express.Router();

const CreateUserSchema = z.object({
  email: z.string().email(),
  name: z.string().min(1),
});

// ❌ VIOLATION: No try/catch and no error handler middleware
router.post('/users', async (req: Request, res: Response, next: NextFunction) => {
  const parsed = CreateUserSchema.safeParse(req.body);
  if (!parsed.success) {
    res.status(400).json({ error: parsed.error });
    return;
  }

  // If createUser throws, it will cause an unhandled promise rejection
  const user = await createUser(parsed.data);

  res.status(201).json(user);
});

async function createUser(data: any): Promise<any> {
  // This might throw if DB connection fails
  throw new Error('Database connection failed');
}

// ❌ No error handler middleware registered

export default router;
```

**Stricture Detection:**
```
ERROR [EXPRESS_UNHANDLED_ASYNC_ERROR] routes/users-bad-3.ts:12:1
  Async route handler missing try/catch block

  Found: async handler with await but no try/catch
  Expected: try/catch with next(error) OR error handler middleware

  Risk: Unhandled promise rejections will crash the server

  Fix options:
    1. Wrap in try/catch and call next(error)
    2. Register error handler middleware:
       router.use((err, req, res, next) => {
         res.status(500).json({ error: 'Internal error' });
       });
```

---

## Fastify

### PERFECT: Schema-Validated Route

```typescript
// routes/products.ts
import { FastifyInstance, FastifyRequest, FastifyReply } from 'fastify';
import { Type, Static } from '@sinclair/typebox';

const CreateProductSchema = Type.Object({
  name: Type.String({ minLength: 1, maxLength: 200 }),
  price: Type.Number({ minimum: 0 }),
  category: Type.Union([
    Type.Literal('electronics'),
    Type.Literal('clothing'),
    Type.Literal('food'),
  ]),
});

type CreateProductBody = Static<typeof CreateProductSchema>;

const ProductResponseSchema = Type.Object({
  id: Type.String(),
  name: Type.String(),
  price: Type.Number(),
  category: Type.String(),
  createdAt: Type.String({ format: 'date-time' }),
});

// ✅ PERFECT: Schema validation, error handling, correct response type
export default async function productRoutes(fastify: FastifyInstance) {
  fastify.post<{ Body: CreateProductBody }>(
    '/products',
    {
      schema: {
        body: CreateProductSchema,
        response: {
          201: ProductResponseSchema,
          400: Type.Object({
            error: Type.String(),
            details: Type.Optional(Type.Array(Type.Any())),
          }),
        },
      },
    },
    async (request: FastifyRequest<{ Body: CreateProductBody }>, reply: FastifyReply) => {
      try {
        // Schema is automatically validated by Fastify
        const product = await createProduct(request.body);

        // Correct status code and response shape
        return reply.status(201).send(product);
      } catch (error) {
        request.log.error(error);
        return reply.status(500).send({ error: 'Failed to create product' });
      }
    }
  );
}

async function createProduct(data: CreateProductBody) {
  return {
    id: crypto.randomUUID(),
    ...data,
    createdAt: new Date().toISOString(),
  };
}
```

**Why PERFECT:**
- ✅ TypeBox schema for request body validation
- ✅ Response schema with status code mapping
- ✅ Type safety with FastifyRequest generics
- ✅ try/catch error handling
- ✅ Correct HTTP status codes (201 for creation)
- ✅ Proper error logging with request.log

---

### VIOLATION 1: Schema Mismatch

```typescript
// routes/products-bad-1.ts
import { FastifyInstance, FastifyRequest, FastifyReply } from 'fastify';
import { Type, Static } from '@sinclair/typebox';

const CreateProductSchema = Type.Object({
  name: Type.String(),
  price: Type.Number(),
  category: Type.String(),
});

type CreateProductBody = Static<typeof CreateProductSchema>;

// ❌ VIOLATION: Response doesn't match declared schema
export default async function productRoutes(fastify: FastifyInstance) {
  fastify.post<{ Body: CreateProductBody }>(
    '/products',
    {
      schema: {
        body: CreateProductSchema,
        response: {
          201: Type.Object({
            id: Type.String(),
            name: Type.String(),
            price: Type.Number(),
          }),
        },
      },
    },
    async (request: FastifyRequest<{ Body: CreateProductBody }>, reply: FastifyReply) => {
      const product = await createProduct(request.body);

      // Response includes fields not in schema (category, createdAt)
      return reply.status(201).send(product);
    }
  );
}

async function createProduct(data: CreateProductBody) {
  return {
    id: crypto.randomUUID(),
    name: data.name,
    price: data.price,
    category: data.category, // ❌ Not in response schema
    createdAt: new Date().toISOString(), // ❌ Not in response schema
  };
}
```

**Stricture Detection:**
```
ERROR [FASTIFY_SCHEMA_MISMATCH] routes/products-bad-1.ts:28:7
  Response object doesn't match declared schema

  Declared schema fields: id, name, price
  Actual response fields: id, name, price, category, createdAt

  Extra fields: category, createdAt

  Fix: Either update response schema to include all fields, or filter response:
    return reply.status(201).send({
      id: product.id,
      name: product.name,
      price: product.price,
    });
```

---

### VIOLATION 2: Missing Error Handler

```typescript
// routes/products-bad-2.ts
import { FastifyInstance, FastifyRequest, FastifyReply } from 'fastify';
import { Type, Static } from '@sinclair/typebox';

const CreateProductSchema = Type.Object({
  name: Type.String(),
  price: Type.Number(),
});

type CreateProductBody = Static<typeof CreateProductSchema>;

// ❌ VIOLATION: No error handling for async operations
export default async function productRoutes(fastify: FastifyInstance) {
  fastify.post<{ Body: CreateProductBody }>(
    '/products',
    {
      schema: {
        body: CreateProductSchema,
      },
    },
    async (request: FastifyRequest<{ Body: CreateProductBody }>, reply: FastifyReply) => {
      // No try/catch - if createProduct throws, it will be unhandled
      const product = await createProduct(request.body);
      return reply.status(201).send(product);
    }
  );
}

async function createProduct(data: CreateProductBody) {
  // This might throw
  throw new Error('Database unavailable');
}
```

**Stricture Detection:**
```
ERROR [FASTIFY_MISSING_ERROR_HANDLER] routes/products-bad-2.ts:19:5
  Async route handler missing error handling

  Found: async handler with await but no try/catch
  Expected: try/catch block or error handler hook

  Fix: Add error handling:
    try {
      const product = await createProduct(request.body);
      return reply.status(201).send(product);
    } catch (error) {
      request.log.error(error);
      return reply.status(500).send({ error: 'Internal error' });
    }
```

---

### VIOLATION 3: Wrong Content Type

```typescript
// routes/products-bad-3.ts
import { FastifyInstance, FastifyRequest, FastifyReply } from 'fastify';
import { Type, Static } from '@sinclair/typebox';

const CreateProductSchema = Type.Object({
  name: Type.String(),
  price: Type.Number(),
});

type CreateProductBody = Static<typeof CreateProductSchema>;

// ❌ VIOLATION: Sends text instead of JSON
export default async function productRoutes(fastify: FastifyInstance) {
  fastify.post<{ Body: CreateProductBody }>(
    '/products',
    {
      schema: {
        body: CreateProductSchema,
      },
    },
    async (request: FastifyRequest<{ Body: CreateProductBody }>, reply: FastifyReply) => {
      const product = await createProduct(request.body);

      // Wrong: Sends plain text instead of JSON
      return reply
        .status(201)
        .type('text/plain')
        .send(`Product created: ${product.id}`);
    }
  );
}

async function createProduct(data: CreateProductBody) {
  return {
    id: crypto.randomUUID(),
    ...data,
  };
}
```

**Stricture Detection:**
```
WARNING [FASTIFY_WRONG_CONTENT_TYPE] routes/products-bad-3.ts:25:9
  POST route returns non-JSON content type

  Found: .type('text/plain')
  Expected: application/json (default)

  Context: Fastify routes typically return JSON
  Fix: Remove .type() call or use reply.send(jsonObject)
```

---

## NestJS

### PERFECT: Controller with DTO + ValidationPipe

```typescript
// dto/create-order.dto.ts
import { IsString, IsNumber, IsArray, ArrayMinSize, Min, Max } from 'class-validator';
import { Type } from 'class-transformer';

class OrderItemDto {
  @IsString()
  productId: string;

  @IsNumber()
  @Min(1)
  @Max(1000)
  quantity: number;
}

export class CreateOrderDto {
  @IsString()
  customerId: string;

  @IsArray()
  @ArrayMinSize(1)
  @Type(() => OrderItemDto)
  items: OrderItemDto[];
}

// orders.controller.ts
import { Controller, Post, Body, HttpCode, HttpStatus, ValidationPipe } from '@nestjs/common';
import { CreateOrderDto } from './dto/create-order.dto';
import { OrdersService } from './orders.service';

@Controller('orders')
export class OrdersController {
  constructor(private readonly ordersService: OrdersService) {}

  // ✅ PERFECT: DTO validation, correct status code, type safety
  @Post()
  @HttpCode(HttpStatus.CREATED)
  async create(
    @Body(new ValidationPipe({ whitelist: true, forbidNonWhitelisted: true }))
    createOrderDto: CreateOrderDto,
  ) {
    try {
      const order = await this.ordersService.create(createOrderDto);

      return {
        id: order.id,
        customerId: order.customerId,
        total: order.total,
        createdAt: order.createdAt,
      };
    } catch (error) {
      // NestJS exception filters handle this
      throw error;
    }
  }
}

// orders.service.ts
import { Injectable } from '@nestjs/common';
import { CreateOrderDto } from './dto/create-order.dto';

@Injectable()
export class OrdersService {
  async create(createOrderDto: CreateOrderDto) {
    const total = createOrderDto.items.reduce(
      (sum, item) => sum + item.quantity * 10, // Simplified pricing
      0
    );

    return {
      id: crypto.randomUUID(),
      customerId: createOrderDto.customerId,
      items: createOrderDto.items,
      total,
      createdAt: new Date().toISOString(),
    };
  }
}
```

**Why PERFECT:**
- ✅ DTO with class-validator decorators
- ✅ ValidationPipe with whitelist (strips unknown properties)
- ✅ forbidNonWhitelisted prevents extra fields
- ✅ Correct @HttpCode decorator (201 for POST)
- ✅ Type safety with TypeScript
- ✅ Service layer separation
- ✅ @Injectable decorator for dependency injection

---

### VIOLATION 1: DTO Mismatch

```typescript
// dto/create-order-bad.dto.ts
import { IsString, IsNumber } from 'class-validator';

export class CreateOrderDto {
  @IsString()
  customerId: string;

  @IsNumber()
  amount: number;
}

// orders-bad-1.controller.ts
import { Controller, Post, Body } from '@nestjs/common';
import { CreateOrderDto } from './dto/create-order-bad.dto';

@Controller('orders')
export class OrdersController {
  // ❌ VIOLATION: Accesses fields not defined in DTO
  @Post()
  async create(@Body() createOrderDto: CreateOrderDto) {
    // DTO only has customerId and amount
    // But code tries to access items (not in DTO)
    const order = {
      id: crypto.randomUUID(),
      customerId: createOrderDto.customerId,
      // @ts-ignore
      items: createOrderDto.items, // ❌ Not in DTO
      // @ts-ignore
      shippingAddress: createOrderDto.shippingAddress, // ❌ Not in DTO
    };

    return order;
  }
}
```

**Stricture Detection:**
```
ERROR [NESTJS_DTO_MISMATCH] orders-bad-1.controller.ts:15:12
  Accessing property not defined in DTO

  DTO type: CreateOrderDto (customerId, amount)
  Accessed property: items

  Additional violations:
    Line 17: shippingAddress (not in DTO)

  Fix: Add properties to DTO with validation decorators:
    @IsArray()
    items: OrderItemDto[];

    @IsObject()
    shippingAddress: AddressDto;
```

---

### VIOLATION 2: Missing ValidationPipe

```typescript
// orders-bad-2.controller.ts
import { Controller, Post, Body, HttpCode, HttpStatus } from '@nestjs/common';
import { CreateOrderDto } from './dto/create-order.dto';

@Controller('orders')
export class OrdersController {
  // ❌ VIOLATION: No ValidationPipe on @Body decorator
  @Post()
  @HttpCode(HttpStatus.CREATED)
  async create(@Body() createOrderDto: CreateOrderDto) {
    // DTO validation decorators are defined but never applied
    // Malformed data will pass through unchecked
    return {
      id: crypto.randomUUID(),
      customerId: createOrderDto.customerId,
      items: createOrderDto.items,
    };
  }
}
```

**Stricture Detection:**
```
ERROR [NESTJS_MISSING_VALIDATION_PIPE] orders-bad-2.controller.ts:9:3
  DTO parameter missing ValidationPipe

  Found: @Body() createOrderDto: CreateOrderDto
  Expected: @Body(new ValidationPipe(...)) createOrderDto: CreateOrderDto

  Risk: DTO validation decorators won't be enforced

  Fix: Add ValidationPipe:
    @Body(new ValidationPipe({
      whitelist: true,
      forbidNonWhitelisted: true,
    }))
    createOrderDto: CreateOrderDto
```

---

### VIOLATION 3: Wrong Decorator Type

```typescript
// orders-bad-3.controller.ts
import { Controller, Get, Body, HttpCode, HttpStatus } from '@nestjs/common';
import { CreateOrderDto } from './dto/create-order.dto';

@Controller('orders')
export class OrdersController {
  // ❌ VIOLATION: Uses @Get for a mutation operation
  @Get()
  @HttpCode(HttpStatus.CREATED)
  async create(@Body() createOrderDto: CreateOrderDto) {
    // Creating a resource with GET violates REST conventions
    return {
      id: crypto.randomUUID(),
      customerId: createOrderDto.customerId,
    };
  }
}
```

**Stricture Detection:**
```
ERROR [NESTJS_WRONG_HTTP_METHOD] orders-bad-3.controller.ts:7:3
  Route mutates state but uses GET method

  Found: @Get() with method name 'create'
  Expected: @Post() for resource creation

  Context: Method name suggests mutation (create/update/delete)
  Fix: Change to @Post() decorator
```

---

## Next.js API Routes

### PERFECT: Validated API Route (App Router)

```typescript
// app/api/comments/route.ts
import { NextRequest, NextResponse } from 'next/server';
import { z } from 'zod';

const CreateCommentSchema = z.object({
  postId: z.string().uuid(),
  content: z.string().min(1).max(5000),
  authorEmail: z.string().email(),
});

type CreateCommentRequest = z.infer<typeof CreateCommentSchema>;

// ✅ PERFECT: Schema validation, error handling, correct status codes
export async function POST(request: NextRequest) {
  try {
    // Parse and validate request body
    const body = await request.json();
    const parsed = CreateCommentSchema.safeParse(body);

    if (!parsed.success) {
      return NextResponse.json(
        {
          error: 'Validation failed',
          details: parsed.error.errors,
        },
        { status: 400 }
      );
    }

    const commentData: CreateCommentRequest = parsed.data;

    // Business logic
    const comment = await createComment(commentData);

    // Correct status code for resource creation
    return NextResponse.json(comment, { status: 201 });
  } catch (error) {
    console.error('Failed to create comment:', error);

    return NextResponse.json(
      { error: 'Failed to create comment' },
      { status: 500 }
    );
  }
}

async function createComment(data: CreateCommentRequest) {
  return {
    id: crypto.randomUUID(),
    ...data,
    createdAt: new Date().toISOString(),
  };
}
```

**Why PERFECT:**
- ✅ Named export (POST function) for App Router
- ✅ Schema validation with safeParse
- ✅ Proper error handling with try/catch
- ✅ Correct HTTP status codes (400, 201, 500)
- ✅ Type safety with z.infer
- ✅ Proper NextResponse.json usage

---

### VIOLATION 1: Wrong HTTP Method (Pages Router)

```typescript
// pages/api/comments-bad-1.ts
import type { NextApiRequest, NextApiResponse } from 'next';
import { z } from 'zod';

const CreateCommentSchema = z.object({
  postId: z.string(),
  content: z.string(),
});

// ❌ VIOLATION: Uses GET to create a resource
export default async function handler(req: NextApiRequest, res: NextApiResponse) {
  // Should check req.method === 'POST'
  if (req.method === 'GET') {
    const parsed = CreateCommentSchema.safeParse(req.body);

    if (!parsed.success) {
      return res.status(400).json({ error: parsed.error });
    }

    const comment = await createComment(parsed.data);
    return res.status(201).json(comment);
  }

  return res.status(405).json({ error: 'Method not allowed' });
}

async function createComment(data: any) {
  return {
    id: crypto.randomUUID(),
    ...data,
  };
}
```

**Stricture Detection:**
```
ERROR [NEXTJS_WRONG_METHOD] pages/api/comments-bad-1.ts:12:1
  Route creates resource using GET method

  Found: req.method === 'GET' with createComment() call
  Expected: req.method === 'POST' for resource creation

  Context: Function name 'createComment' suggests mutation
  Fix: Change to if (req.method === 'POST')
```

---

### VIOLATION 2: Missing Validation

```typescript
// app/api/comments/route-bad-2.ts
import { NextRequest, NextResponse } from 'next/server';

// ❌ VIOLATION: No schema validation on request body
export async function POST(request: NextRequest) {
  try {
    // Directly uses request body without validation
    const body = await request.json();

    // No validation - malformed data can cause errors
    const comment = await createComment(body);

    return NextResponse.json(comment, { status: 201 });
  } catch (error) {
    return NextResponse.json(
      { error: 'Failed to create comment' },
      { status: 500 }
    );
  }
}

async function createComment(data: any) {
  // This assumes data is well-formed
  return {
    id: crypto.randomUUID(),
    postId: data.postId,
    content: data.content,
    authorEmail: data.authorEmail,
  };
}
```

**Stricture Detection:**
```
ERROR [NEXTJS_MISSING_VALIDATION] app/api/comments/route-bad-2.ts:7:5
  Request body used without validation

  Found: await request.json() directly passed to business logic
  Expected: Schema validation before processing

  Fix: Add schema validation:
    const parsed = CreateCommentSchema.safeParse(body);
    if (!parsed.success) {
      return NextResponse.json(
        { error: parsed.error },
        { status: 400 }
      );
    }
```

---

### VIOLATION 3: Wrong Response Shape

```typescript
// app/api/comments/[id]/route-bad-3.ts
import { NextRequest, NextResponse } from 'next/server';

// ❌ VIOLATION: Returns inconsistent response shapes
export async function GET(
  request: NextRequest,
  { params }: { params: { id: string } }
) {
  try {
    const comment = await getComment(params.id);

    if (!comment) {
      // Inconsistent: error case returns string
      return NextResponse.json('Comment not found', { status: 404 });
    }

    // Success case returns object
    return NextResponse.json(comment);
  } catch (error) {
    // Another inconsistent format
    return NextResponse.json(error, { status: 500 });
  }
}

async function getComment(id: string) {
  if (id === 'missing') {
    return null;
  }
  return {
    id,
    content: 'Test comment',
  };
}
```

**Stricture Detection:**
```
ERROR [NEXTJS_INCONSISTENT_RESPONSE] app/api/comments/[id]/route-bad-3.ts:13:7
  API route returns inconsistent response shapes

  Success: object with properties
  Error (404): string
  Error (500): Error object

  Expected: All responses should follow consistent schema:
    { data?: T, error?: string }

  Fix: Standardize error responses:
    return NextResponse.json(
      { error: 'Comment not found' },
      { status: 404 }
    );
```

---

## Hono

### PERFECT: Validated Hono Route

```typescript
// routes/webhooks.ts
import { Hono } from 'hono';
import { validator } from 'hono/validator';
import { z } from 'zod';

const WebhookPayloadSchema = z.object({
  event: z.enum(['user.created', 'user.updated', 'user.deleted']),
  userId: z.string().uuid(),
  timestamp: z.string().datetime(),
  data: z.record(z.unknown()),
});

type WebhookPayload = z.infer<typeof WebhookPayloadSchema>;

const app = new Hono();

// ✅ PERFECT: Validator middleware, error handling, correct status
app.post(
  '/webhooks',
  validator('json', (value, c) => {
    const parsed = WebhookPayloadSchema.safeParse(value);
    if (!parsed.success) {
      return c.json(
        {
          error: 'Invalid webhook payload',
          details: parsed.error.errors,
        },
        400
      );
    }
    return parsed.data;
  }),
  async (c) => {
    try {
      const payload = c.req.valid('json') as WebhookPayload;

      // Process webhook
      const result = await processWebhook(payload);

      return c.json(
        {
          received: true,
          eventId: result.eventId,
        },
        202 // Accepted (async processing)
      );
    } catch (error) {
      console.error('Webhook processing failed:', error);
      return c.json({ error: 'Processing failed' }, 500);
    }
  }
);

async function processWebhook(payload: WebhookPayload) {
  return {
    eventId: crypto.randomUUID(),
    event: payload.event,
    processedAt: new Date().toISOString(),
  };
}

export default app;
```

**Why PERFECT:**
- ✅ Hono validator middleware with zod schema
- ✅ safeParse with error handling
- ✅ Type-safe with c.req.valid('json')
- ✅ Correct HTTP status (202 for async processing)
- ✅ try/catch error handling
- ✅ Proper Hono context (c) usage

---

### VIOLATION 1: Missing Validator Middleware

```typescript
// routes/webhooks-bad-1.ts
import { Hono } from 'hono';

const app = new Hono();

// ❌ VIOLATION: No validation middleware
app.post('/webhooks', async (c) => {
  // Directly accesses body without validation
  const payload = await c.req.json();

  // Assumes payload is well-formed
  const result = await processWebhook(payload);

  return c.json({ received: true, eventId: result.eventId }, 202);
});

async function processWebhook(payload: any) {
  return {
    eventId: crypto.randomUUID(),
    event: payload.event, // Might be undefined or wrong type
  };
}

export default app;
```

**Stricture Detection:**
```
ERROR [HONO_MISSING_VALIDATOR] routes/webhooks-bad-1.ts:6:1
  POST route missing validator middleware

  Found: Direct c.req.json() without validation
  Expected: validator('json', schema) middleware

  Fix: Add validator middleware:
    app.post(
      '/webhooks',
      validator('json', (value, c) => {
        const parsed = WebhookPayloadSchema.safeParse(value);
        if (!parsed.success) {
          return c.json({ error: parsed.error }, 400);
        }
        return parsed.data;
      }),
      async (c) => { ... }
    );
```

---

### VIOLATION 2: Unsafe Type Assertion

```typescript
// routes/webhooks-bad-2.ts
import { Hono } from 'hono';
import { validator } from 'hono/validator';
import { z } from 'zod';

const WebhookPayloadSchema = z.object({
  event: z.string(),
  userId: z.string(),
});

const app = new Hono();

// ❌ VIOLATION: Uses .parse() which throws instead of safeParse
app.post(
  '/webhooks',
  validator('json', (value, c) => {
    // This will throw on invalid data, causing 500 instead of 400
    const parsed = WebhookPayloadSchema.parse(value); // ❌ Throws
    return parsed;
  }),
  async (c) => {
    const payload = c.req.valid('json');
    return c.json({ received: true }, 202);
  }
);

export default app;
```

**Stricture Detection:**
```
ERROR [HONO_UNSAFE_PARSE] routes/webhooks-bad-2.ts:17:5
  Validator uses .parse() which throws on invalid input

  Found: WebhookPayloadSchema.parse(value)
  Expected: WebhookPayloadSchema.safeParse(value)

  Risk: Unhandled validation errors become 500 instead of 400

  Fix: Use safeParse and handle errors:
    const parsed = WebhookPayloadSchema.safeParse(value);
    if (!parsed.success) {
      return c.json({ error: parsed.error }, 400);
    }
    return parsed.data;
```

---

### VIOLATION 3: Wrong Status Code

```typescript
// routes/webhooks-bad-3.ts
import { Hono } from 'hono';
import { validator } from 'hono/validator';
import { z } from 'zod';

const WebhookPayloadSchema = z.object({
  event: z.string(),
  userId: z.string(),
});

const app = new Hono();

// ❌ VIOLATION: Returns 200 for async operation instead of 202
app.post(
  '/webhooks',
  validator('json', (value, c) => {
    const parsed = WebhookPayloadSchema.safeParse(value);
    if (!parsed.success) {
      return c.json({ error: parsed.error }, 400);
    }
    return parsed.data;
  }),
  async (c) => {
    const payload = c.req.valid('json');

    // Queues async processing
    await queueWebhook(payload);

    // Wrong: Should be 202 (Accepted) for async processing
    return c.json({ received: true }, 200);
  }
);

async function queueWebhook(payload: any) {
  // Simulated async queue
  console.log('Queued webhook:', payload);
}

export default app;
```

**Stricture Detection:**
```
WARNING [HONO_WRONG_STATUS_CODE] routes/webhooks-bad-3.ts:28:5
  Async webhook processing returns 200 instead of 202

  Found: c.json({ received: true }, 200)
  Expected: c.json({ received: true }, 202)

  Context: Function 'queueWebhook' suggests async/deferred processing
  Fix: Use 202 (Accepted) for webhooks that process asynchronously
```

---

## Summary

**Covered Frameworks:** 5
**Total Examples:** 20 (5 frameworks × 4 examples each)
**Pattern Categories:**
- Route handler detection (all frameworks)
- Schema validation (Express/Fastify/NestJS/Next.js/Hono)
- Error handling (async/await patterns)
- HTTP status codes (REST conventions)
- Type safety (TypeScript generics, DTOs, validators)

**Key Stricture Capabilities Demonstrated:**
1. Framework-specific route detection (decorators, exports, method names)
2. Validation library integration (zod, class-validator, TypeBox)
3. HTTP method/status code conventions
4. Response shape consistency
5. Error handling patterns (try/catch, middleware, pipes)

**LOC:** ~1,150 lines

