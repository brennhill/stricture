# 50 — Convention Patterns

**Why included:** Validates all 6 CONV rules across TypeScript, Go, Python, and Java for file naming, headers, error formats, export naming, test locations, and required exports.

## Rules Tested

- **CONV-file-naming**: kebab-case (TS), snake_case (Go/Python), PascalCase overrides for UI components
- **CONV-file-header**: Required format `// filename — purpose` for Go/TS, docstrings for Python, javadoc for Java
- **CONV-error-format**: `{OPERATION}: {ROOT_CAUSE}. {RECOVERY_ACTION}`
- **CONV-export-naming**: Functions camelCase (TS), PascalCase (Go), types PascalCase, constants SCREAMING_SNAKE
- **CONV-test-file-location**: Mirror strategy (tests/ mirrors src/)
- **CONV-required-exports**: Modules must have index.ts with named exports

## Perfect Example

```
project/
├── src/
│   ├── index.ts                    ✓ Required root export
│   ├── user-service.ts             ✓ kebab-case TS file
│   ├── components/
│   │   ├── index.ts                ✓ Required module export
│   │   └── UserCard.tsx            ✓ PascalCase override for component
│   ├── utils/
│   │   ├── index.ts                ✓ Required module export
│   │   └── format-date.ts          ✓ kebab-case utility
│   └── constants/
│       ├── index.ts                ✓ Required module export
│       └── api-config.ts           ✓ kebab-case constants file
├── internal/
│   └── user_service.go             ✓ snake_case Go file with header
├── pkg/
│   └── formatter.go                ✓ snake_case Go file with header
├── scripts/
│   └── user_migration.py           ✓ snake_case Python file
├── src/main/java/
│   └── UserService.java            ✓ PascalCase Java file
└── tests/
    ├── user-service.test.ts        ✓ Mirror location for TS
    ├── components/
    │   └── UserCard.test.tsx       ✓ Mirror location for component
    └── utils/
        └── format-date.test.ts     ✓ Mirror location for utility
```

### src/index.ts (Root Export - PERFECT)

```typescript
// index.ts — Root module exports for user service

export { createUserService, type UserServiceConfig } from './user-service';
export { UserCard } from './components';
export { formatDate, parseDate } from './utils';
export { API_BASE_URL, DEFAULT_TIMEOUT } from './constants';
```

**Why perfect:**
- ✓ File header present with correct format
- ✓ Named exports from all modules
- ✓ Types exported with `type` keyword
- ✓ Clean barrel export pattern

### src/user-service.ts (TypeScript Service - PERFECT)

```typescript
// user-service.ts — User service with CRUD operations and error handling

export interface UserServiceConfig {
  apiUrl: string;
  timeout: number;
  retryAttempts: number;
}

export interface User {
  id: string;
  name: string;
  email: string;
}

export const MAX_RETRY_ATTEMPTS = 3;
export const DEFAULT_TIMEOUT_MS = 5000;

export function createUserService(config: UserServiceConfig) {
  return {
    async getUser(id: string): Promise<User> {
      try {
        const response = await fetch(`${config.apiUrl}/users/${id}`);
        if (!response.ok) {
          throw new Error(
            `FETCH_USER: Network response failed with status ${response.status}. Verify user ID and retry.`
          );
        }
        return await response.json();
      } catch (error) {
        if (error instanceof TypeError) {
          throw new Error(
            `FETCH_USER: Network connection failed. Check API URL configuration and network connectivity.`
          );
        }
        throw error;
      }
    },

    async createUser(user: Omit<User, 'id'>): Promise<User> {
      try {
        const response = await fetch(`${config.apiUrl}/users`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(user),
        });
        if (!response.ok) {
          throw new Error(
            `CREATE_USER: Server rejected request with status ${response.status}. Validate user data format.`
          );
        }
        return await response.json();
      } catch (error) {
        if (error instanceof SyntaxError) {
          throw new Error(
            `CREATE_USER: Invalid JSON response from server. Contact API support.`
          );
        }
        throw error;
      }
    },
  };
}
```

**Why perfect:**
- ✓ File name: kebab-case `user-service.ts`
- ✓ File header: `// user-service.ts — User service with CRUD operations...`
- ✓ Export naming: Functions `createUserService` (camelCase), types `UserServiceConfig` (PascalCase), constants `MAX_RETRY_ATTEMPTS` (SCREAMING_SNAKE)
- ✓ Error format: `{OPERATION}: {ROOT_CAUSE}. {RECOVERY_ACTION}`

### src/components/index.ts (Component Barrel - PERFECT)

```typescript
// index.ts — UI components module exports

export { UserCard, type UserCardProps } from './UserCard';
```

**Why perfect:**
- ✓ Required module export
- ✓ File header present
- ✓ Named exports with types

### src/components/UserCard.tsx (React Component - PERFECT)

```typescript
// UserCard.tsx — User profile card component with avatar and details

import React from 'react';

export interface UserCardProps {
  userId: string;
  name: string;
  email: string;
  avatarUrl?: string;
}

export function UserCard({ userId, name, email, avatarUrl }: UserCardProps) {
  const [isLoading, setIsLoading] = React.useState(false);

  const handleClick = async () => {
    try {
      setIsLoading(true);
      // Simulate API call
      await fetch(`/api/users/${userId}`);
    } catch (error) {
      throw new Error(
        `LOAD_USER_DETAILS: Failed to fetch user data. Check network connection and try again.`
      );
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="user-card" onClick={handleClick}>
      {avatarUrl && <img src={avatarUrl} alt={`${name} avatar`} />}
      <h3>{name}</h3>
      <p>{email}</p>
      {isLoading && <span>Loading...</span>}
    </div>
  );
}
```

**Why perfect:**
- ✓ File name: PascalCase `UserCard.tsx` (component override)
- ✓ File header: `// UserCard.tsx — User profile card component...`
- ✓ Export naming: Component `UserCard` (PascalCase), props interface `UserCardProps` (PascalCase)
- ✓ Error format: Follows `{OPERATION}: {ROOT_CAUSE}. {RECOVERY_ACTION}`

### src/utils/index.ts (Utils Barrel - PERFECT)

```typescript
// index.ts — Date formatting and parsing utilities

export { formatDate, parseDate, isValidDate } from './format-date';
```

**Why perfect:**
- ✓ Required module export
- ✓ File header present
- ✓ Named exports

### src/utils/format-date.ts (Utility Functions - PERFECT)

```typescript
// format-date.ts — Date formatting and validation utilities

export const DATE_FORMAT_ISO = 'YYYY-MM-DD';
export const DATE_FORMAT_US = 'MM/DD/YYYY';

export function formatDate(date: Date, format: string = DATE_FORMAT_ISO): string {
  try {
    if (!isValidDate(date)) {
      throw new Error(
        `FORMAT_DATE: Invalid date object provided. Pass a valid Date instance.`
      );
    }
    // Implementation
    return date.toISOString().split('T')[0];
  } catch (error) {
    if (error instanceof RangeError) {
      throw new Error(
        `FORMAT_DATE: Date value out of valid range. Use dates between 1970-2100.`
      );
    }
    throw error;
  }
}

export function parseDate(dateString: string): Date {
  const parsed = new Date(dateString);
  if (!isValidDate(parsed)) {
    throw new Error(
      `PARSE_DATE: Date string '${dateString}' is not valid ISO 8601 format. Use YYYY-MM-DD format.`
    );
  }
  return parsed;
}

export function isValidDate(date: Date): boolean {
  return date instanceof Date && !isNaN(date.getTime());
}
```

**Why perfect:**
- ✓ File name: kebab-case `format-date.ts`
- ✓ File header: `// format-date.ts — Date formatting and validation utilities`
- ✓ Export naming: Functions `formatDate`, `parseDate` (camelCase), constants `DATE_FORMAT_ISO` (SCREAMING_SNAKE)
- ✓ Error format: All errors follow `{OPERATION}: {ROOT_CAUSE}. {RECOVERY_ACTION}`

### src/constants/index.ts (Constants Barrel - PERFECT)

```typescript
// index.ts — API configuration constants

export { API_BASE_URL, API_TIMEOUT, DEFAULT_TIMEOUT, MAX_RETRIES } from './api-config';
```

**Why perfect:**
- ✓ Required module export
- ✓ File header present
- ✓ Named exports

### src/constants/api-config.ts (Constants - PERFECT)

```typescript
// api-config.ts — API endpoint and timeout configuration

export const API_BASE_URL = process.env.API_URL || 'https://api.example.com';
export const API_TIMEOUT = 30000;
export const DEFAULT_TIMEOUT = 5000;
export const MAX_RETRIES = 3;
```

**Why perfect:**
- ✓ File name: kebab-case `api-config.ts`
- ✓ File header: `// api-config.ts — API endpoint and timeout configuration`
- ✓ Export naming: All constants SCREAMING_SNAKE_CASE

---

## Violation Examples

### V01 — TypeScript PascalCase filename (CONV-file-naming)

**Violation:** TypeScript file using PascalCase instead of required kebab-case
**Expected violation:** `CONV-file-naming` on file `UserService.ts`
**Language:** TypeScript

```typescript
// UserService.ts — User service implementation

export function createUserService() {
  return {
    getUser: (id: string) => fetch(`/users/${id}`),
  };
}
```

**Why Stricture catches this:** TypeScript files must use kebab-case (`user-service.ts`), not PascalCase. PascalCase is only allowed for React/UI components (`.tsx` files like `UserCard.tsx`).

**Fix:**
```typescript
// Rename file: UserService.ts → user-service.ts
```

---

### V02 — Go camelCase filename (CONV-file-naming)

**Violation:** Go file using camelCase instead of required snake_case
**Expected violation:** `CONV-file-naming` on file `userService.go`
**Language:** Go

```go
// userService.go — User service implementation

package service

func CreateUserService() *UserService {
    return &UserService{}
}
```

**Why Stricture catches this:** Go files must use snake_case (`user_service.go`), not camelCase. This follows standard Go package naming conventions.

**Fix:**
```go
// Rename file: userService.go → user_service.go
```

---

### V03 — Python PascalCase filename (CONV-file-naming)

**Violation:** Python file using PascalCase instead of required snake_case
**Expected violation:** `CONV-file-naming` on file `UserModel.py`
**Language:** Python

```python
# UserModel.py — User model class

class UserModel:
    def __init__(self, id: str, name: str):
        self.id = id
        self.name = name
```

**Why Stricture catches this:** Python files must use snake_case (`user_model.py`), not PascalCase. This follows PEP 8 module naming conventions.

**Fix:**
```python
# Rename file: UserModel.py → user_model.py
```

---

### V04 — TypeScript missing file header (CONV-file-header)

**Violation:** TypeScript file missing required `// filename — purpose` header
**Expected violation:** `CONV-file-header` on file `user-service.ts`
**Language:** TypeScript

```typescript
export interface User {
  id: string;
  name: string;
}

export function getUser(id: string): Promise<User> {
  return fetch(`/users/${id}`).then(res => res.json());
}
```

**Why Stricture catches this:** All TypeScript files must start with a single-line comment header in the format `// filename — purpose description`.

**Fix:**
```typescript
// user-service.ts — User service with fetch operations

export interface User {
  id: string;
  name: string;
}

export function getUser(id: string): Promise<User> {
  return fetch(`/users/${id}`).then(res => res.json());
}
```

---

### V05 — Go missing file header (CONV-file-header)

**Violation:** Go file missing required `// filename — purpose` header
**Expected violation:** `CONV-file-header` on file `user_service.go`
**Language:** Go

```go
package service

import "fmt"

func GetUser(id string) string {
    return fmt.Sprintf("User %s", id)
}
```

**Why Stricture catches this:** All Go files must start with a single-line comment header in the format `// filename — purpose description`, placed before the package declaration.

**Fix:**
```go
// user_service.go — User service operations

package service

import "fmt"

func GetUser(id string) string {
    return fmt.Sprintf("User %s", id)
}
```

---

### V06 — Python missing file header (CONV-file-header)

**Violation:** Python file missing required `# filename — purpose` docstring
**Expected violation:** `CONV-file-header` on file `user_model.py`
**Language:** Python

```python
from typing import Protocol

class UserModel(Protocol):
    id: str
    name: str
```

**Why Stricture catches this:** All Python files must start with a module-level docstring in the format `"""filename — purpose description"""`.

**Fix:**
```python
"""user_model.py — User model protocol definitions"""

from typing import Protocol

class UserModel(Protocol):
    id: str
    name: str
```

---

### V07 — TypeScript unstructured error message (CONV-error-format)

**Violation:** Error message missing required `{OPERATION}: {ROOT_CAUSE}. {RECOVERY_ACTION}` format
**Expected violation:** `CONV-error-format` on line with `throw new Error`
**Language:** TypeScript

```typescript
// user-service.ts — User service operations

export async function getUser(id: string) {
  const response = await fetch(`/users/${id}`);
  if (!response.ok) {
    throw new Error("something went wrong");
  }
  return response.json();
}
```

**Why Stricture catches this:** Error messages must follow the structured format `{OPERATION}: {ROOT_CAUSE}. {RECOVERY_ACTION}` to provide actionable context.

**Fix:**
```typescript
// user-service.ts — User service operations

export async function getUser(id: string) {
  const response = await fetch(`/users/${id}`);
  if (!response.ok) {
    throw new Error(
      `FETCH_USER: HTTP request failed with status ${response.status}. Verify user ID and retry.`
    );
  }
  return response.json();
}
```

---

### V08 — Go unstructured error message (CONV-error-format)

**Violation:** Go error missing required structured format
**Expected violation:** `CONV-error-format` on line with `fmt.Errorf`
**Language:** Go

```go
// user_service.go — User service operations

package service

import "fmt"

func GetUser(id string) (string, error) {
    if id == "" {
        return "", fmt.Errorf("error happened")
    }
    return id, nil
}
```

**Why Stricture catches this:** Go error messages must follow `{OPERATION}: {ROOT_CAUSE}. {RECOVERY_ACTION}` format for consistency and debuggability.

**Fix:**
```go
// user_service.go — User service operations

package service

import "fmt"

func GetUser(id string) (string, error) {
    if id == "" {
        return "", fmt.Errorf("GET_USER: Empty user ID provided. Pass a non-empty string")
    }
    return id, nil
}
```

---

### V09 — Python unstructured error message (CONV-error-format)

**Violation:** Python exception missing required structured format
**Expected violation:** `CONV-error-format` on line with `raise ValueError`
**Language:** Python

```python
"""user_service.py — User service operations"""

def get_user(user_id: str) -> dict:
    if not user_id:
        raise ValueError("bad input")
    return {"id": user_id}
```

**Why Stricture catches this:** Python exception messages must follow `{OPERATION}: {ROOT_CAUSE}. {RECOVERY_ACTION}` format for consistent error handling.

**Fix:**
```python
"""user_service.py — User service operations"""

def get_user(user_id: str) -> dict:
    if not user_id:
        raise ValueError(
            "GET_USER: Empty user_id provided. Pass a non-empty string"
        )
    return {"id": user_id}
```

---

### V10 — TypeScript generic export name (CONV-export-naming)

**Violation:** Export function using generic name instead of specific, descriptive name
**Expected violation:** `CONV-export-naming` on export `getData`
**Language:** TypeScript

```typescript
// user-service.ts — User service operations

export const getData = async (id: string) => {
  return fetch(`/users/${id}`).then(res => res.json());
};
```

**Why Stricture catches this:** Export names must be specific and descriptive. `getData` is too generic — it should be `getUser`, `getUserById`, or similar to indicate what data is being fetched.

**Fix:**
```typescript
// user-service.ts — User service operations

export const getUserById = async (id: string) => {
  return fetch(`/users/${id}`).then(res => res.json());
};
```

---

### V11 — TypeScript default export (CONV-export-naming)

**Violation:** Using default export instead of required named export
**Expected violation:** `CONV-export-naming` on `export default`
**Language:** TypeScript

```typescript
// user-service.ts — User service operations

class UserService {
  async getUser(id: string) {
    return fetch(`/users/${id}`).then(res => res.json());
  }
}

export default UserService;
```

**Why Stricture catches this:** Default exports are forbidden. All exports must be named exports to enable better tree-shaking, refactoring, and IDE support.

**Fix:**
```typescript
// user-service.ts — User service operations

export class UserService {
  async getUser(id: string) {
    return fetch(`/users/${id}`).then(res => res.json());
  }
}
```

---

### V12 — Go unexported public function (CONV-export-naming)

**Violation:** Function that should be public API is unexported (lowercase first letter)
**Expected violation:** `CONV-export-naming` on unexported `getUserByID`
**Language:** Go

```go
// user_service.go — User service public API

package service

// getUserByID fetches a user by their ID
func getUserByID(id string) (string, error) {
    return id, nil
}
```

**Why Stricture catches this:** Public API functions must be exported (PascalCase). If `getUserByID` is meant to be used by external packages, it should be `GetUserByID`.

**Fix:**
```go
// user_service.go — User service public API

package service

// GetUserByID fetches a user by their ID
func GetUserByID(id string) (string, error) {
    return id, nil
}
```

---

### V13 — Test file in src directory (CONV-test-file-location)

**Violation:** Test file located in `src/` instead of mirrored `tests/` directory
**Expected violation:** `CONV-test-file-location` on file `src/services/user-service.test.ts`
**Language:** TypeScript

```
project/
├── src/
│   ├── services/
│   │   ├── user-service.ts
│   │   └── user-service.test.ts    ← VIOLATION: Test in src/
└── tests/
    └── (empty)
```

**Why Stricture catches this:** Tests must mirror the source structure in a separate `tests/` directory. This keeps source clean and makes test discovery predictable.

**Fix:**
```
project/
├── src/
│   ├── services/
│   │   └── user-service.ts
└── tests/
    └── services/
        └── user-service.test.ts    ← Correct location
```

---

### V14 — Go test at wrong directory level (CONV-test-file-location)

**Violation:** Go test file at incorrect directory level, breaking package structure
**Expected violation:** `CONV-test-file-location` on file `cmd/main_test.go`
**Language:** Go

```
project/
├── cmd/
│   ├── main.go
│   └── main_test.go          ← VIOLATION: Should be in cmd/ or internal/_test/
└── internal/
    └── service/
        └── user_service.go
```

**Why Stricture catches this:** Go tests should either be alongside the code they test (same package) or in a mirrored `_test/` directory for integration tests. Placing `main_test.go` in `cmd/` breaks the package boundary.

**Fix:**
```
project/
├── cmd/
│   └── main.go
└── internal/
    ├── service/
    │   ├── user_service.go
    │   └── user_service_test.go    ← Alongside for unit tests
    └── _test/
        └── integration_test.go     ← Separate for integration tests
```

---

### V15 — Test file not mirroring source structure (CONV-test-file-location)

**Violation:** Test file location doesn't mirror the source directory hierarchy
**Expected violation:** `CONV-test-file-location` on file `tests/all-tests.test.ts`
**Language:** TypeScript

```
project/
├── src/
│   ├── services/
│   │   └── user-service.ts
│   └── utils/
│       └── format-date.ts
└── tests/
    └── all-tests.test.ts         ← VIOLATION: Doesn't mirror src/ structure
```

**Why Stricture catches this:** Test files must mirror the source directory structure. Each source file should have a corresponding test file in the mirrored location.

**Fix:**
```
project/
├── src/
│   ├── services/
│   │   └── user-service.ts
│   └── utils/
│       └── format-date.ts
└── tests/
    ├── services/
    │   └── user-service.test.ts  ← Mirrors src/services/
    └── utils/
        └── format-date.test.ts   ← Mirrors src/utils/
```

---

### V16 — Module missing index.ts barrel export (CONV-required-exports)

**Violation:** Module directory missing required `index.ts` barrel export file
**Expected violation:** `CONV-required-exports` on module `src/services/`
**Language:** TypeScript

```
project/
├── src/
│   ├── index.ts
│   └── services/
│       ├── user-service.ts
│       └── auth-service.ts
│       # Missing index.ts!
```

**Why Stricture catches this:** Every module directory must have an `index.ts` that re-exports its public API. This enables clean imports like `import { createUserService } from './services'` instead of `import { createUserService } from './services/user-service'`.

**Fix:**
```
project/
├── src/
│   ├── index.ts
│   └── services/
│       ├── index.ts              ← Add barrel export
│       ├── user-service.ts
│       └── auth-service.ts
```

With `src/services/index.ts`:
```typescript
// index.ts — Services module exports

export { createUserService } from './user-service';
export { createAuthService } from './auth-service';
```

---

### V17 — Go package missing expected exports (CONV-required-exports)

**Violation:** Go package missing documented exported functions that should be public API
**Expected violation:** `CONV-required-exports` on package `pkg/formatter`
**Language:** Go

```go
// formatter.go — String formatting utilities (MISSING exports)

package formatter

// formatUser is unexported, but should be part of public API
func formatUser(name string) string {
    return "User: " + name
}

// parseUserInput is unexported, but should be part of public API
func parseUserInput(input string) string {
    return input
}
```

**Why Stricture catches this:** If a package is in `pkg/` (indicating public API), its main functions should be exported. Godoc comments without exported functions indicate a convention violation.

**Fix:**
```go
// formatter.go — String formatting utilities

package formatter

// FormatUser formats a user name with prefix
func FormatUser(name string) string {
    return "User: " + name
}

// ParseUserInput sanitizes and parses user input
func ParseUserInput(input string) string {
    return input
}
```

---

### V18 — Python module missing __all__ declarations (CONV-required-exports)

**Violation:** Python package `__init__.py` missing `__all__` to define public exports
**Expected violation:** `CONV-required-exports` on file `src/services/__init__.py`
**Language:** Python

```python
"""services/__init__.py — Services module"""

from .user_service import UserService
from .auth_service import AuthService

# Missing __all__ declaration!
```

**Why Stricture catches this:** Python packages should explicitly declare their public API using `__all__`. This controls what gets exported with `from services import *` and signals the intended public interface.

**Fix:**
```python
"""services/__init__.py — Services module"""

from .user_service import UserService
from .auth_service import AuthService

__all__ = [
    "UserService",
    "AuthService",
]
```

