# 40 — Test Quality Patterns

**Why included:** Validation for all 10 TQ (Test Quality) rules. Includes realistic source files and comprehensive test suites demonstrating correct and incorrect patterns across Jest, Go testify, and Python pytest frameworks.

**Rules validated:**
- TQ-no-shallow-assertions
- TQ-return-type-verified
- TQ-schema-conformance
- TQ-assertion-depth
- TQ-boundary-tested
- TQ-mock-scope
- TQ-test-isolation
- TQ-test-naming
- TQ-negative-cases
- TQ-error-path-coverage

---

## Source Files Under Test

### 1. UserService (TypeScript)

```typescript
// src/services/user-service.ts
export interface User {
  id: string;
  name: string;
  email: string;
  role: 'admin' | 'user' | 'guest';
  createdAt: Date;
  address?: {
    street: string;
    city: string;
    country: string;
  };
}

export class UserService {
  private db: Database;

  async createUser(name: string, email: string, role: string): Promise<User> {
    if (!name || !email) {
      throw new Error('Name and email are required');
    }
    if (!['admin', 'user', 'guest'].includes(role)) {
      throw new Error('Invalid role');
    }
    const user = await this.db.insert({ name, email, role, createdAt: new Date() });
    return user;
  }

  async getUser(id: string): Promise<User | null> {
    if (!id) {
      throw new Error('User ID is required');
    }
    try {
      const user = await this.db.findById(id);
      return user || null;
    } catch (error) {
      if (error.code === 'NOT_FOUND') {
        return null;
      }
      throw new Error('Database error');
    }
  }

  async updateUser(id: string, updates: Partial<User>): Promise<User> {
    if (!id) {
      throw new Error('User ID is required');
    }
    try {
      const user = await this.db.update(id, updates);
      if (!user) {
        throw new Error('User not found');
      }
      return user;
    } catch (error) {
      if (error.code === 'VALIDATION_ERROR') {
        throw new Error('Invalid update data');
      }
      throw error;
    }
  }

  async listUsers(page: number, pageSize: number): Promise<User[]> {
    const users = await this.db.findAll({ page, pageSize });
    return users;
  }
}
```

### 2. PaginationUtil (TypeScript)

```typescript
// src/utils/pagination.ts
export interface PaginationParams {
  page: number;      // 1-100
  pageSize: number;  // 1-50
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

export class PaginationUtil {
  static validate(params: PaginationParams): void {
    if (params.page < 1 || params.page > 100) {
      throw new Error('Page must be between 1 and 100');
    }
    if (params.pageSize < 1 || params.pageSize > 50) {
      throw new Error('Page size must be between 1 and 50');
    }
  }

  static paginate<T>(items: T[], params: PaginationParams): PaginatedResponse<T> {
    this.validate(params);
    const total = items.length;
    const totalPages = Math.ceil(total / params.pageSize);
    const start = (params.page - 1) * params.pageSize;
    const end = start + params.pageSize;
    const pageItems = items.slice(start, end);

    return {
      items: pageItems,
      total,
      page: params.page,
      pageSize: params.pageSize,
      totalPages,
    };
  }
}
```

### 3. ConfigLoader (TypeScript)

```typescript
// src/config/loader.ts
export interface AppConfig {
  apiKey: string;
  apiUrl: string;
  timeout: number;
  debug?: boolean;
  features?: {
    analytics: boolean;
    logging: boolean;
  };
}

export class ConfigLoader {
  async loadFromEnv(): Promise<AppConfig> {
    const apiKey = process.env.API_KEY;
    const apiUrl = process.env.API_URL;
    const timeout = process.env.TIMEOUT ? parseInt(process.env.TIMEOUT, 10) : 5000;

    if (!apiKey || !apiUrl) {
      throw new Error('API_KEY and API_URL are required');
    }

    return {
      apiKey,
      apiUrl,
      timeout,
      debug: process.env.DEBUG === 'true',
      features: process.env.FEATURES ? JSON.parse(process.env.FEATURES) : undefined,
    };
  }

  async loadFromFile(path: string): Promise<AppConfig> {
    try {
      const content = await fs.readFile(path, 'utf-8');
      const config = JSON.parse(content);

      if (!config.apiKey || !config.apiUrl) {
        throw new Error('Invalid config file: missing required fields');
      }

      return {
        apiKey: config.apiKey,
        apiUrl: config.apiUrl,
        timeout: config.timeout || 5000,
        debug: config.debug,
        features: config.features,
      };
    } catch (error) {
      if (error.code === 'ENOENT') {
        throw new Error('Config file not found');
      }
      throw new Error('Failed to parse config file');
    }
  }
}
```

### 4. WebhookProcessor (TypeScript)

```typescript
// src/webhooks/processor.ts
export interface WebhookPayload {
  event: string;
  timestamp: number;
  data: Record<string, unknown>;
}

export class WebhookProcessor {
  private secret: string;
  private maxRetries: number = 3;

  verifySignature(payload: string, signature: string): boolean {
    const expectedSignature = crypto
      .createHmac('sha256', this.secret)
      .update(payload)
      .digest('hex');
    return signature === expectedSignature;
  }

  async process(payload: WebhookPayload, signature: string): Promise<void> {
    const payloadStr = JSON.stringify(payload);

    if (!this.verifySignature(payloadStr, signature)) {
      throw new Error('Invalid signature');
    }

    let attempt = 0;
    while (attempt < this.maxRetries) {
      try {
        await this.handleEvent(payload);
        return;
      } catch (error) {
        attempt++;
        if (attempt >= this.maxRetries) {
          throw new Error('Max retries exceeded');
        }
        await this.delay(1000 * attempt);
      }
    }
  }

  private async handleEvent(payload: WebhookPayload): Promise<void> {
    // Event handling logic
  }

  private delay(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}
```

---

## PERFECT — All 10 TQ Rules Pass

### UserService Tests (Jest) — PERFECT

```typescript
// tests/services/user-service.test.ts
import { UserService } from '../../src/services/user-service';
import { Database } from '../../src/db';

describe('UserService', () => {
  let userService: UserService;
  let mockDb: jest.Mocked<Database>;

  beforeEach(() => {
    // TQ-mock-scope: Fresh mocks in beforeEach
    mockDb = {
      insert: jest.fn(),
      findById: jest.fn(),
      update: jest.fn(),
      findAll: jest.fn(),
    } as any;
    userService = new UserService(mockDb);
  });

  afterEach(() => {
    // TQ-mock-scope: Cleanup in afterEach
    jest.clearAllMocks();
  });

  describe('createUser', () => {
    // TQ-test-naming: Clear "should [verb] when [condition]" format
    it('should create user when valid inputs provided', async () => {
      const mockUser = {
        id: 'user-123',
        name: 'John Doe',
        email: 'john@example.com',
        role: 'user' as const,
        createdAt: new Date('2026-01-01T00:00:00Z'),
      };
      mockDb.insert.mockResolvedValue(mockUser);

      const result = await userService.createUser('John Doe', 'john@example.com', 'user');

      // TQ-return-type-verified: All fields of User interface checked
      expect(result.id).toBe('user-123');
      expect(result.name).toBe('John Doe');
      expect(result.email).toBe('john@example.com');
      expect(result.role).toBe('user');
      expect(result.createdAt).toEqual(new Date('2026-01-01T00:00:00Z'));

      // TQ-schema-conformance: Verify no extra fields
      expect(Object.keys(result)).toEqual(['id', 'name', 'email', 'role', 'createdAt']);
    });

    // TQ-negative-cases: Error test for every happy path
    it('should throw error when name is empty', async () => {
      await expect(
        userService.createUser('', 'john@example.com', 'user')
      ).rejects.toThrow('Name and email are required');
    });

    it('should throw error when email is empty', async () => {
      await expect(
        userService.createUser('John Doe', '', 'user')
      ).rejects.toThrow('Name and email are required');
    });

    // TQ-boundary-tested: Invalid role boundary
    it('should throw error when role is invalid', async () => {
      await expect(
        userService.createUser('John Doe', 'john@example.com', 'superadmin')
      ).rejects.toThrow('Invalid role');
    });
  });

  describe('getUser', () => {
    it('should return user when user exists', async () => {
      const mockUser = {
        id: 'user-456',
        name: 'Jane Smith',
        email: 'jane@example.com',
        role: 'admin' as const,
        createdAt: new Date('2026-01-15T10:30:00Z'),
        address: {
          street: '123 Main St',
          city: 'New York',
          country: 'USA',
        },
      };
      mockDb.findById.mockResolvedValue(mockUser);

      const result = await userService.getUser('user-456');

      // TQ-return-type-verified: All required fields
      expect(result).not.toBeNull();
      expect(result!.id).toBe('user-456');
      expect(result!.name).toBe('Jane Smith');
      expect(result!.email).toBe('jane@example.com');
      expect(result!.role).toBe('admin');
      expect(result!.createdAt).toEqual(new Date('2026-01-15T10:30:00Z'));

      // TQ-assertion-depth: Nested object fields verified
      expect(result!.address).toBeDefined();
      expect(result!.address!.street).toBe('123 Main St');
      expect(result!.address!.city).toBe('New York');
      expect(result!.address!.country).toBe('USA');
    });

    it('should return null when user not found', async () => {
      mockDb.findById.mockResolvedValue(null);

      const result = await userService.getUser('nonexistent');

      expect(result).toBeNull();
    });

    // TQ-error-path-coverage: Test NOT_FOUND error branch
    it('should return null when database throws NOT_FOUND error', async () => {
      const notFoundError = new Error('Not found');
      (notFoundError as any).code = 'NOT_FOUND';
      mockDb.findById.mockRejectedValue(notFoundError);

      const result = await userService.getUser('user-789');

      expect(result).toBeNull();
    });

    // TQ-error-path-coverage: Test generic error branch
    it('should throw database error when unknown error occurs', async () => {
      mockDb.findById.mockRejectedValue(new Error('Connection failed'));

      await expect(
        userService.getUser('user-789')
      ).rejects.toThrow('Database error');
    });

    it('should throw error when id is empty', async () => {
      await expect(
        userService.getUser('')
      ).rejects.toThrow('User ID is required');
    });
  });

  describe('updateUser', () => {
    it('should update user when valid data provided', async () => {
      const mockUpdatedUser = {
        id: 'user-123',
        name: 'John Updated',
        email: 'john.updated@example.com',
        role: 'admin' as const,
        createdAt: new Date('2026-01-01T00:00:00Z'),
      };
      mockDb.update.mockResolvedValue(mockUpdatedUser);

      const result = await userService.updateUser('user-123', {
        name: 'John Updated',
        email: 'john.updated@example.com',
      });

      // TQ-return-type-verified: All fields verified
      expect(result.id).toBe('user-123');
      expect(result.name).toBe('John Updated');
      expect(result.email).toBe('john.updated@example.com');
      expect(result.role).toBe('admin');
      expect(result.createdAt).toEqual(new Date('2026-01-01T00:00:00Z'));
    });

    it('should throw error when user not found', async () => {
      mockDb.update.mockResolvedValue(null);

      await expect(
        userService.updateUser('nonexistent', { name: 'Updated' })
      ).rejects.toThrow('User not found');
    });

    // TQ-error-path-coverage: Test VALIDATION_ERROR branch
    it('should throw validation error when invalid data provided', async () => {
      const validationError = new Error('Invalid email format');
      (validationError as any).code = 'VALIDATION_ERROR';
      mockDb.update.mockRejectedValue(validationError);

      await expect(
        userService.updateUser('user-123', { email: 'invalid-email' })
      ).rejects.toThrow('Invalid update data');
    });

    // TQ-error-path-coverage: Test generic error rethrow branch
    it('should rethrow unknown errors', async () => {
      const unknownError = new Error('Database connection lost');
      mockDb.update.mockRejectedValue(unknownError);

      await expect(
        userService.updateUser('user-123', { name: 'Updated' })
      ).rejects.toThrow('Database connection lost');
    });

    it('should throw error when id is empty', async () => {
      await expect(
        userService.updateUser('', { name: 'Updated' })
      ).rejects.toThrow('User ID is required');
    });
  });

  describe('listUsers', () => {
    it('should return paginated users', async () => {
      const mockUsers = [
        {
          id: 'user-1',
          name: 'User One',
          email: 'one@example.com',
          role: 'user' as const,
          createdAt: new Date('2026-01-01T00:00:00Z'),
        },
        {
          id: 'user-2',
          name: 'User Two',
          email: 'two@example.com',
          role: 'admin' as const,
          createdAt: new Date('2026-01-02T00:00:00Z'),
        },
        {
          id: 'user-3',
          name: 'User Three',
          email: 'three@example.com',
          role: 'guest' as const,
          createdAt: new Date('2026-01-03T00:00:00Z'),
        },
      ];
      mockDb.findAll.mockResolvedValue(mockUsers);

      const result = await userService.listUsers(1, 10);

      // TQ-assertion-depth: Not just .toHaveLength, but verify each user's fields
      expect(result).toHaveLength(3);

      expect(result[0].id).toBe('user-1');
      expect(result[0].name).toBe('User One');
      expect(result[0].email).toBe('one@example.com');
      expect(result[0].role).toBe('user');
      expect(result[0].createdAt).toEqual(new Date('2026-01-01T00:00:00Z'));

      expect(result[1].id).toBe('user-2');
      expect(result[1].name).toBe('User Two');
      expect(result[1].email).toBe('two@example.com');
      expect(result[1].role).toBe('admin');

      expect(result[2].id).toBe('user-3');
      expect(result[2].name).toBe('User Three');
      expect(result[2].email).toBe('three@example.com');
      expect(result[2].role).toBe('guest');
    });
  });
});
```

### PaginationUtil Tests (Jest) — PERFECT

```typescript
// tests/utils/pagination.test.ts
import { PaginationUtil } from '../../src/utils/pagination';

describe('PaginationUtil', () => {
  // TQ-test-isolation: Create fresh data in each test
  const createTestData = () => {
    return Array.from({ length: 100 }, (_, i) => ({
      id: `item-${i + 1}`,
      name: `Item ${i + 1}`,
    }));
  };

  describe('validate', () => {
    // TQ-boundary-tested: Test all boundary values for page
    it('should accept page=1 (minimum valid)', () => {
      expect(() => {
        PaginationUtil.validate({ page: 1, pageSize: 10 });
      }).not.toThrow();
    });

    it('should accept page=100 (maximum valid)', () => {
      expect(() => {
        PaginationUtil.validate({ page: 100, pageSize: 10 });
      }).not.toThrow();
    });

    it('should reject page=0 (below minimum)', () => {
      expect(() => {
        PaginationUtil.validate({ page: 0, pageSize: 10 });
      }).toThrow('Page must be between 1 and 100');
    });

    it('should reject page=101 (above maximum)', () => {
      expect(() => {
        PaginationUtil.validate({ page: 101, pageSize: 10 });
      }).toThrow('Page must be between 1 and 100');
    });

    // TQ-boundary-tested: Test all boundary values for pageSize
    it('should accept pageSize=1 (minimum valid)', () => {
      expect(() => {
        PaginationUtil.validate({ page: 1, pageSize: 1 });
      }).not.toThrow();
    });

    it('should accept pageSize=50 (maximum valid)', () => {
      expect(() => {
        PaginationUtil.validate({ page: 1, pageSize: 50 });
      }).not.toThrow();
    });

    it('should reject pageSize=0 (below minimum)', () => {
      expect(() => {
        PaginationUtil.validate({ page: 1, pageSize: 0 });
      }).toThrow('Page size must be between 1 and 50');
    });

    it('should reject pageSize=51 (above maximum)', () => {
      expect(() => {
        PaginationUtil.validate({ page: 1, pageSize: 51 });
      }).toThrow('Page size must be between 1 and 50');
    });
  });

  describe('paginate', () => {
    it('should return first page of results', () => {
      const items = createTestData();

      const result = PaginationUtil.paginate(items, { page: 1, pageSize: 10 });

      // TQ-return-type-verified: All fields of PaginatedResponse verified
      expect(result.items).toHaveLength(10);
      expect(result.total).toBe(100);
      expect(result.page).toBe(1);
      expect(result.pageSize).toBe(10);
      expect(result.totalPages).toBe(10);

      // TQ-schema-conformance: Verify exact schema
      expect(Object.keys(result).sort()).toEqual(['items', 'page', 'pageSize', 'total', 'totalPages'].sort());

      // TQ-assertion-depth: Verify actual item content
      expect(result.items[0].id).toBe('item-1');
      expect(result.items[0].name).toBe('Item 1');
      expect(result.items[9].id).toBe('item-10');
      expect(result.items[9].name).toBe('Item 10');
    });

    it('should return middle page of results', () => {
      const items = createTestData();

      const result = PaginationUtil.paginate(items, { page: 5, pageSize: 10 });

      expect(result.items).toHaveLength(10);
      expect(result.page).toBe(5);
      expect(result.items[0].id).toBe('item-41');
      expect(result.items[9].id).toBe('item-50');
    });

    it('should return last page of results', () => {
      const items = createTestData();

      const result = PaginationUtil.paginate(items, { page: 10, pageSize: 10 });

      expect(result.items).toHaveLength(10);
      expect(result.page).toBe(10);
      expect(result.items[0].id).toBe('item-91');
      expect(result.items[9].id).toBe('item-100');
    });

    it('should handle partial last page', () => {
      const items = createTestData().slice(0, 25);

      const result = PaginationUtil.paginate(items, { page: 3, pageSize: 10 });

      expect(result.items).toHaveLength(5);
      expect(result.total).toBe(25);
      expect(result.totalPages).toBe(3);
      expect(result.items[0].id).toBe('item-21');
      expect(result.items[4].id).toBe('item-25');
    });

    it('should throw error when page is invalid', () => {
      const items = createTestData();

      expect(() => {
        PaginationUtil.paginate(items, { page: 0, pageSize: 10 });
      }).toThrow('Page must be between 1 and 100');
    });
  });
});
```

---

## VIOLATIONS — TQ Rule Failures

### V01 — Shallow API Response Assertion (TQ-no-shallow-assertions)
**Violation:** Using `toBeDefined()` on API response without checking actual fields
**Expected violation:** `TQ-no-shallow-assertions`

**Source:**
```typescript
// src/api/fetch-user.ts
export async function fetchUser(id: string): Promise<User> {
  const response = await fetch(`/api/users/${id}`);
  return response.json();
}
```

**Bad Test:**
```typescript
// tests/api/fetch-user.test.ts
it('should fetch user', async () => {
  const result = await fetchUser('user-123');

  // TQ-no-shallow-assertions: Shallow assertion doesn't verify API contract
  expect(result).toBeDefined();
  // Missing: result.id, result.name, result.email, etc.
});
```

**Why Stricture catches this:** `toBeDefined()` and `toBeTruthy()` assertions on complex return values (objects, arrays) don't verify the actual data structure, which means API contract changes won't be caught by tests.

---

### V02 — Shallow Boolean Result Assertion (TQ-no-shallow-assertions)
**Violation:** Using `toBeTruthy()` on operation result without checking state
**Expected violation:** `TQ-no-shallow-assertions`

**Source:**
```typescript
// src/auth/login.ts
export async function login(email: string, password: string): Promise<boolean> {
  const response = await authService.authenticate({ email, password });
  return response.success;
}
```

**Bad Test:**
```typescript
// tests/auth/login.test.ts
it('should login successfully', async () => {
  const result = await login('user@example.com', 'password123');

  // TQ-no-shallow-assertions: Doesn't verify side effects or state changes
  expect(result).toBeTruthy();
  // Missing: session state, auth token, user context, etc.
});
```

**Why Stricture catches this:** For operations with side effects (login, logout, save, etc.), only checking the return value is insufficient. Tests should verify the operation's full impact on system state.

---

### V03 — Return Type Not Verified (TQ-return-type-verified)
**Violation:** Test doesn't assert return type shape from function
**Expected violation:** `TQ-return-type-verified`

**Source:**
```typescript
// src/parsers/json-parser.ts
export interface ParsedData {
  id: string;
  metadata: {
    version: string;
    timestamp: number;
  };
  items: string[];
}

export function parseData(input: string): ParsedData {
  const raw = JSON.parse(input);
  return {
    id: raw.id,
    metadata: { version: raw.version, timestamp: raw.ts },
    items: raw.items || [],
  };
}
```

**Bad Test:**
```typescript
// tests/parsers/json-parser.test.ts
it('should parse JSON data', () => {
  const input = '{"id":"123","version":"1.0","ts":1234567890,"items":["a","b"]}';
  const result = parseData(input);

  // TQ-return-type-verified: Only checks one field, not full return type
  expect(result.id).toBe('123');
  // Missing: metadata.version, metadata.timestamp, items array content
});
```

**Why Stricture catches this:** Return type verification ensures the function contract is honored. Without checking all fields of the return type, interface changes or parser bugs can go undetected.

---

### V04 — Side Effect Checked But Not Return Value (TQ-return-type-verified)
**Violation:** Test checks side effect but not return value
**Expected violation:** `TQ-return-type-verified`

**Source:**
```typescript
// src/storage/cache.ts
export class Cache {
  private store = new Map<string, string>();

  set(key: string, value: string): { stored: boolean; key: string } {
    this.store.set(key, value);
    return { stored: true, key };
  }
}
```

**Bad Test:**
```typescript
// tests/storage/cache.test.ts
it('should store value in cache', () => {
  const cache = new Cache();
  const result = cache.set('key1', 'value1');

  // TQ-return-type-verified: Only checks side effect, not return value
  expect(cache.get('key1')).toBe('value1');
  // Missing: result.stored, result.key
});
```

**Why Stricture catches this:** Functions with both return values and side effects should have both tested. Ignoring return values means callers relying on that return data won't have test coverage.

---

### V05 — Hardcoded Mock Doesn't Match Schema (TQ-schema-conformance)
**Violation:** Test uses hardcoded mock that doesn't match actual API schema
**Expected violation:** `TQ-schema-conformance`

**Source:**
```typescript
// src/api/types.ts
export interface ApiResponse {
  status: number;
  data: {
    userId: string;
    userName: string;
    userEmail: string;
    createdAt: string;
  };
  meta: {
    requestId: string;
    timestamp: number;
  };
}
```

**Bad Test:**
```typescript
// tests/api/client.test.ts
it('should handle API response', async () => {
  // TQ-schema-conformance: Mock missing required 'meta' field
  const mockResponse: ApiResponse = {
    status: 200,
    data: {
      userId: 'user-123',
      userName: 'John Doe',
      userEmail: 'john@example.com',
      createdAt: '2026-01-01T00:00:00Z',
    },
    // Missing: meta field (required by ApiResponse interface)
  } as ApiResponse; // Type assertion bypasses TypeScript checking

  const result = processResponse(mockResponse);
  expect(result.userId).toBe('user-123');
});
```

**Why Stricture catches this:** Type assertions (`as ApiResponse`) bypass TypeScript's type checking. Tests should use fully-conformant mocks to catch schema violations at test time.

---

### V06 — Mock Response Missing Required Fields (TQ-schema-conformance)
**Violation:** Mock response missing required fields
**Expected violation:** `TQ-schema-conformance`

**Source:**
```typescript
// src/webhooks/types.ts
export interface WebhookEvent {
  eventId: string;
  eventType: string;
  payload: Record<string, unknown>;
  timestamp: number;
  signature: string;
}
```

**Bad Test:**
```typescript
// tests/webhooks/handler.test.ts
it('should process webhook event', async () => {
  // TQ-schema-conformance: Mock missing 'signature' field
  const mockEvent = {
    eventId: 'evt-123',
    eventType: 'user.created',
    payload: { userId: 'user-456' },
    timestamp: 1234567890,
    // Missing: signature (required by WebhookEvent)
  };

  await handleWebhook(mockEvent as WebhookEvent);
  expect(eventProcessor.received).toHaveBeenCalled();
});
```

**Why Stricture catches this:** Incomplete mocks create false confidence. If production code depends on the missing field (e.g., signature validation), tests will pass but production will fail.

---

### V07 — No Network Error Test (TQ-error-path-coverage)
**Violation:** No test for network error case
**Expected violation:** `TQ-error-path-coverage`

**Source:**
```typescript
// src/api/client.ts
export async function fetchData(url: string): Promise<Data> {
  try {
    const response = await fetch(url);
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`);
    }
    return response.json();
  } catch (error) {
    if (error instanceof TypeError) {
      throw new Error('Network error');
    }
    throw error;
  }
}
```

**Bad Test:**
```typescript
// tests/api/client.test.ts
describe('fetchData', () => {
  it('should fetch data successfully', async () => {
    global.fetch = jest.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ id: '123', name: 'Test' }),
    });

    const result = await fetchData('/api/data');
    expect(result.id).toBe('123');
  });

  // TQ-error-path-coverage: Missing test for TypeError/network error branch
  // Missing: Test for fetch() throwing TypeError (network failure)
});
```

**Why Stricture catches this:** Network errors are common in production. The `catch (error)` branch that handles `TypeError` is completely untested, meaning network error handling is unverified.

---

### V08 — No Malformed Response Test (TQ-error-path-coverage)
**Violation:** No test for malformed response body
**Expected violation:** `TQ-error-path-coverage`

**Source:**
```typescript
// src/parsers/response-parser.ts
export async function parseResponse(response: Response): Promise<ParsedData> {
  try {
    const json = await response.json();
    return {
      id: json.id,
      value: json.value,
      timestamp: new Date(json.timestamp),
    };
  } catch (error) {
    throw new Error('Failed to parse response: malformed JSON');
  }
}
```

**Bad Test:**
```typescript
// tests/parsers/response-parser.test.ts
it('should parse valid response', async () => {
  const mockResponse = new Response(
    JSON.stringify({ id: '123', value: 'test', timestamp: '2026-01-01T00:00:00Z' })
  );

  const result = await parseResponse(mockResponse);
  expect(result.id).toBe('123');
  expect(result.value).toBe('test');
});

// TQ-error-path-coverage: Missing test for .json() throwing error
// Missing: Test for malformed JSON (triggers catch block)
```

**Why Stricture catches this:** The `catch` block handles JSON parsing errors but has zero test coverage. Malformed responses in production will trigger untested error paths.

---

### V09 — Shallow Object Assertion (TQ-assertion-depth)
**Violation:** Only checks top-level fields, not nested objects
**Expected violation:** `TQ-assertion-depth`

**Source:**
```typescript
// src/services/profile-service.ts
export interface UserProfile {
  userId: string;
  name: string;
  contact: {
    email: string;
    phone: string;
    address: {
      street: string;
      city: string;
      zip: string;
    };
  };
}

export function getProfile(userId: string): UserProfile {
  return database.fetchProfile(userId);
}
```

**Bad Test:**
```typescript
// tests/services/profile-service.test.ts
it('should get user profile', () => {
  const profile = getProfile('user-123');

  // TQ-assertion-depth: Only checks top-level fields
  expect(profile.userId).toBe('user-123');
  expect(profile.name).toBe('John Doe');
  expect(profile.contact).toBeDefined();
  // Missing: contact.email, contact.phone, contact.address.street, etc.
});
```

**Why Stricture catches this:** Using `toBeDefined()` on nested objects without asserting their fields means changes to nested structure won't be caught. Should verify all levels of the object hierarchy.

---

### V10 — Array Length Without Element Verification (TQ-assertion-depth)
**Violation:** Checks array length but not array element structure
**Expected violation:** `TQ-assertion-depth`

**Source:**
```typescript
// src/services/order-service.ts
export interface OrderItem {
  productId: string;
  quantity: number;
  price: number;
  discount?: number;
}

export function getOrderItems(orderId: string): OrderItem[] {
  return database.fetchOrderItems(orderId);
}
```

**Bad Test:**
```typescript
// tests/services/order-service.test.ts
it('should get order items', () => {
  const items = getOrderItems('order-456');

  // TQ-assertion-depth: Only checks array length
  expect(items).toHaveLength(3);
  // Missing: items[0].productId, items[0].quantity, items[0].price, etc.
});
```

**Why Stricture catches this:** Checking only array length doesn't verify the structure or content of elements. Tests should assert on actual element values to catch data corruption or mapping errors.

---

### V11 — No Empty String Boundary Test (TQ-boundary-tested)
**Violation:** No test for empty string input
**Expected violation:** `TQ-boundary-tested`

**Source:**
```typescript
// src/validators/email-validator.ts
export function validateEmail(email: string): boolean {
  if (email.length === 0) {
    return false;
  }
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
}
```

**Bad Test:**
```typescript
// tests/validators/email-validator.test.ts
describe('validateEmail', () => {
  it('should validate correct email', () => {
    expect(validateEmail('user@example.com')).toBe(true);
  });

  it('should reject invalid email', () => {
    expect(validateEmail('notanemail')).toBe(false);
  });

  // TQ-boundary-tested: Missing empty string test
  // Missing: validateEmail('') test (boundary condition)
});
```

**Why Stricture catches this:** Empty string is a critical boundary for string inputs. The code has an explicit `email.length === 0` check that's completely untested.

---

### V12 — No Maximum Value Boundary Test (TQ-boundary-tested)
**Violation:** No test for maximum value boundary
**Expected violation:** `TQ-boundary-tested`

**Source:**
```typescript
// src/validators/age-validator.ts
export function validateAge(age: number): boolean {
  if (age < 0 || age > 150) {
    throw new Error('Age must be between 0 and 150');
  }
  return true;
}
```

**Bad Test:**
```typescript
// tests/validators/age-validator.test.ts
describe('validateAge', () => {
  it('should accept valid age', () => {
    expect(validateAge(25)).toBe(true);
  });

  it('should reject negative age', () => {
    expect(() => validateAge(-1)).toThrow('Age must be between 0 and 150');
  });

  // TQ-boundary-tested: Missing maximum boundary test
  // Missing: validateAge(150) and validateAge(151) tests
});
```

**Why Stricture catches this:** Testing only the minimum boundary (negative) but not the maximum (150/151) leaves half the validation logic unverified. Boundary values are where off-by-one errors occur.

---

### V13 — Global Mock Leak (TQ-mock-scope)
**Violation:** Global mock leaks between tests (no afterEach cleanup)
**Expected violation:** `TQ-mock-scope`

**Source:**
```typescript
// src/utils/time.ts
export function getCurrentTime(): number {
  return Date.now();
}
```

**Bad Test:**
```typescript
// tests/utils/time.test.ts
describe('getCurrentTime', () => {
  it('should return current timestamp', () => {
    // TQ-mock-scope: Global Date.now() mocked but never cleaned up
    const mockNow = jest.fn(() => 1234567890);
    global.Date.now = mockNow;

    expect(getCurrentTime()).toBe(1234567890);
    // Missing: Restore Date.now() in afterEach
  });

  it('should return different timestamp', () => {
    // This test will FAIL because Date.now is still mocked from previous test
    expect(getCurrentTime()).toBeGreaterThan(0);
  });
});
```

**Why Stricture catches this:** Global mocks that aren't cleaned up in `afterEach` cause test pollution. Subsequent tests inherit the mock, leading to unexpected failures and flaky tests.

---

### V14 — Permanent Module Override (TQ-mock-scope)
**Violation:** Mock overrides real module permanently
**Expected violation:** `TQ-mock-scope`

**Source:**
```typescript
// src/config/env.ts
export const config = {
  apiUrl: process.env.API_URL || 'https://api.example.com',
  apiKey: process.env.API_KEY || '',
};
```

**Bad Test:**
```typescript
// tests/config/env.test.ts
jest.mock('../../src/config/env', () => ({
  config: {
    apiUrl: 'https://mock.example.com',
    apiKey: 'mock-key',
  },
}));

describe('config', () => {
  it('should use mocked config', () => {
    const { config } = require('../../src/config/env');
    expect(config.apiUrl).toBe('https://mock.example.com');
  });
});

// TQ-mock-scope: Module mock persists for entire test suite
// All other tests in this file will get the mock, not the real module
```

**Why Stricture catches this:** Top-level `jest.mock()` calls apply to the entire test file. If tests need different mock configurations, this creates conflicts. Use `jest.doMock()` in individual tests or factory functions for per-test mocks.

---

### V15 — Shared Mutable State (TQ-test-isolation)
**Violation:** Tests share mutable state via module-level variable
**Expected violation:** `TQ-test-isolation`

**Source:**
```typescript
// src/services/counter-service.ts
export class CounterService {
  private count = 0;

  increment(): number {
    this.count++;
    return this.count;
  }

  getCount(): number {
    return this.count;
  }
}
```

**Bad Test:**
```typescript
// tests/services/counter-service.test.ts
// TQ-test-isolation: Shared instance across tests
const counterService = new CounterService();

describe('CounterService', () => {
  it('should increment to 1', () => {
    const result = counterService.increment();
    expect(result).toBe(1);
  });

  it('should get current count', () => {
    // This test FAILS if run after first test (expects 0, gets 1)
    const count = counterService.getCount();
    expect(count).toBe(0); // FAILS: count is 1 from previous test
  });
});
```

**Why Stricture catches this:** Module-level instances persist across tests, creating order-dependent failures. Each test should create fresh instances in `beforeEach` to ensure isolation.

---

### V16 — Order-Dependent Test (TQ-test-isolation)
**Violation:** Test depends on execution order (passes alone, fails in suite)
**Expected violation:** `TQ-test-isolation`

**Source:**
```typescript
// src/storage/store.ts
export class Store {
  private data: Record<string, string> = {};

  set(key: string, value: string): void {
    this.data[key] = value;
  }

  get(key: string): string | undefined {
    return this.data[key];
  }
}
```

**Bad Test:**
```typescript
// tests/storage/store.test.ts
describe('Store', () => {
  const store = new Store();

  it('should set value', () => {
    store.set('key1', 'value1');
    expect(store.get('key1')).toBe('value1');
  });

  it('should get previously set value', () => {
    // TQ-test-isolation: Depends on previous test running first
    expect(store.get('key1')).toBe('value1'); // FAILS if run in isolation
  });

  it('should handle multiple keys', () => {
    store.set('key2', 'value2');
    // Assumes key1 still exists from first test
    expect(store.get('key1')).toBe('value1');
    expect(store.get('key2')).toBe('value2');
  });
});
```

**Why Stricture catches this:** Tests that depend on previous tests running create brittle suites. Running tests in isolation (`.only`) or in different orders will cause failures. Each test should be self-contained.

---

### V17 — Only Happy Path Tested (TQ-negative-cases)
**Violation:** Only happy path tested, no invalid input test
**Expected violation:** `TQ-negative-cases`

**Source:**
```typescript
// src/math/calculator.ts
export function divide(a: number, b: number): number {
  if (b === 0) {
    throw new Error('Division by zero');
  }
  return a / b;
}
```

**Bad Test:**
```typescript
// tests/math/calculator.test.ts
describe('divide', () => {
  it('should divide two numbers', () => {
    expect(divide(10, 2)).toBe(5);
    expect(divide(15, 3)).toBe(5);
    expect(divide(100, 10)).toBe(10);
  });

  // TQ-negative-cases: Missing test for division by zero
  // Missing: Test for divide(10, 0) throwing error
});
```

**Why Stricture catches this:** The error handling logic (`if (b === 0)`) has zero test coverage. Production code with error branches must have tests for both success and failure paths.

---

### V18 — No Unauthorized Access Test (TQ-negative-cases)
**Violation:** No test for unauthorized access
**Expected violation:** `TQ-negative-cases`

**Source:**
```typescript
// src/auth/resource-guard.ts
export async function accessResource(userId: string, resourceId: string): Promise<Resource> {
  const user = await userService.getUser(userId);
  if (!user) {
    throw new Error('User not found');
  }

  const resource = await resourceService.getResource(resourceId);
  if (resource.ownerId !== userId) {
    throw new Error('Unauthorized: You do not own this resource');
  }

  return resource;
}
```

**Bad Test:**
```typescript
// tests/auth/resource-guard.test.ts
describe('accessResource', () => {
  it('should return resource when user is owner', async () => {
    const mockUser = { id: 'user-123', name: 'John' };
    const mockResource = { id: 'res-456', ownerId: 'user-123', data: 'test' };

    userService.getUser = jest.fn().mockResolvedValue(mockUser);
    resourceService.getResource = jest.fn().mockResolvedValue(mockResource);

    const result = await accessResource('user-123', 'res-456');
    expect(result.id).toBe('res-456');
  });

  // TQ-negative-cases: Missing test for unauthorized access
  // Missing: Test where resource.ownerId !== userId (throws 'Unauthorized')
});
```

**Why Stricture catches this:** Security-critical logic (authorization checks) must have negative test cases. The `resource.ownerId !== userId` branch is completely untested, meaning authorization bugs won't be caught.

---

### V19 — Generic Test Names (TQ-test-naming)
**Violation:** Test name `test1`, `test2` (not descriptive)
**Expected violation:** `TQ-test-naming`

**Source:**
```typescript
// src/formatters/date-formatter.ts
export function formatDate(date: Date, format: string): string {
  if (format === 'short') {
    return `${date.getMonth() + 1}/${date.getDate()}/${date.getFullYear()}`;
  }
  return date.toISOString();
}
```

**Bad Test:**
```typescript
// tests/formatters/date-formatter.test.ts
describe('formatDate', () => {
  // TQ-test-naming: Non-descriptive test names
  it('test1', () => {
    const date = new Date('2026-01-15T00:00:00Z');
    const result = formatDate(date, 'short');
    expect(result).toBe('1/15/2026');
  });

  it('test2', () => {
    const date = new Date('2026-01-15T00:00:00Z');
    const result = formatDate(date, 'iso');
    expect(result).toBe('2026-01-15T00:00:00.000Z');
  });
});
```

**Why Stricture catches this:** Test names like `test1`, `test2`, `testCase`, `scenario1` provide no information about what's being tested. When tests fail in CI, developers can't determine the failure cause without reading the code.

---

### V20 — Test Name Doesn't Describe Behavior (TQ-test-naming)
**Violation:** Test name doesn't describe expected behavior
**Expected violation:** `TQ-test-naming`

**Source:**
```typescript
// src/validators/password-validator.ts
export function validatePassword(password: string): { valid: boolean; errors: string[] } {
  const errors: string[] = [];

  if (password.length < 8) {
    errors.push('Password must be at least 8 characters');
  }
  if (!/[A-Z]/.test(password)) {
    errors.push('Password must contain at least one uppercase letter');
  }
  if (!/[0-9]/.test(password)) {
    errors.push('Password must contain at least one number');
  }

  return { valid: errors.length === 0, errors };
}
```

**Bad Test:**
```typescript
// tests/validators/password-validator.test.ts
describe('validatePassword', () => {
  // TQ-test-naming: Test names don't describe behavior
  it('validates password', () => {
    const result = validatePassword('Abc12345');
    expect(result.valid).toBe(true);
  });

  it('checks password', () => {
    const result = validatePassword('abc');
    expect(result.valid).toBe(false);
  });

  it('password validation', () => {
    const result = validatePassword('abcdefgh');
    expect(result.valid).toBe(false);
    expect(result.errors).toContain('Password must contain at least one uppercase letter');
  });
});
```

**Why Stricture catches this:** Names like "validates password", "checks password", "password validation" don't specify what aspect is being tested. Better: "should accept password with uppercase and number", "should reject password shorter than 8 characters", "should reject password without uppercase letter".

