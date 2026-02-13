# 60 — GraphQL API (GitHub GraphQL v4)

**Protocol:** GraphQL
**Endpoint:** `POST /graphql`
**Auth:** Bearer token in `Authorization` header
**Operations:** Queries (repository, viewer, search), Mutations (createIssue, addComment)
**Complexity:** Single endpoint with varying query shapes, GraphQL error format, nullable fields, fragments, connection/edge pagination

---

## Overview

GraphQL APIs present unique challenges compared to REST:
- **Single endpoint:** All queries go to `/graphql`, differentiated by request body
- **Error format:** HTTP 200 with `{ data: null, errors: [...] }` for GraphQL errors
- **Nullable fields:** Most fields are nullable unless marked with `!` in schema
- **Fragments:** Query reusability and type conditions
- **Connection pagination:** `edges`, `nodes`, `pageInfo` pattern instead of offset/limit
- **Union types:** Results can be multiple types (e.g., `SearchResultItemConnection`)

This validation set tests CTR rules on non-REST APIs where traditional HTTP status code patterns don't apply.

---

## Manifest

### Operations

#### Queries

```graphql
query GetRepository($owner: String!, $name: String!) {
  repository(owner: $owner, name: $name) {
    id
    name
    description
    stargazerCount
    forkCount
    isPrivate
  }
}
```

```graphql
query GetViewer {
  viewer {
    login
    name
    email
    bio
    company
    location
  }
}
```

```graphql
query SearchRepositories($query: String!, $first: Int!, $after: String) {
  search(query: $query, type: REPOSITORY, first: $first, after: $after) {
    repositoryCount
    pageInfo {
      hasNextPage
      endCursor
    }
    edges {
      node {
        ... on Repository {
          id
          name
          owner {
            login
          }
          stargazerCount
        }
      }
    }
  }
}
```

#### Mutations

```graphql
mutation CreateIssue($repositoryId: ID!, $title: String!, $body: String) {
  createIssue(input: {repositoryId: $repositoryId, title: $title, body: $body}) {
    issue {
      id
      number
      title
      url
    }
  }
}
```

```graphql
mutation AddComment($subjectId: ID!, $body: String!) {
  addComment(input: {subjectId: $subjectId, body: $body}) {
    commentEdge {
      node {
        id
        body
        createdAt
      }
    }
  }
}
```

### Error Format

GraphQL returns HTTP 200 even for errors:

```json
{
  "data": null,
  "errors": [
    {
      "type": "NOT_FOUND",
      "path": ["repository"],
      "locations": [{"line": 2, "column": 3}],
      "message": "Could not resolve to a Repository with the name 'nonexistent'."
    }
  ]
}
```

Partial success also possible:

```json
{
  "data": {
    "repository": {
      "name": "my-repo",
      "description": null
    }
  },
  "errors": [
    {
      "type": "FORBIDDEN",
      "path": ["repository", "description"],
      "message": "Resource protected by organization SAML enforcement."
    }
  ]
}
```

---

## PERFECT Implementation

Location: `examples/perfect/graphql-client.ts`

### Requirements

1. **Typed Queries:**
   - TypeScript interfaces for all query variables and responses
   - Strict null checks enabled
   - Type guards for union types (e.g., `SearchResultItem`)

2. **Fragment Handling:**
   - Reusable fragment definitions
   - Fragment spreading for common fields
   - Type conditions for union/interface types

3. **Nullable Field Handling:**
   - All fields treated as potentially null unless schema says otherwise
   - Default values or explicit null checks
   - No unsafe property access

4. **Connection Pagination:**
   - `edges` and `nodes` properly typed
   - `pageInfo` checked for `hasNextPage`
   - `after` cursor tracking across pages
   - `first` parameter for page size

5. **Error Handling:**
   - Check `response.errors` array even when status is 200
   - Partial success detection (both `data` and `errors` present)
   - Path-based error attribution
   - GraphQL error types mapped to recovery actions

6. **Rate Limiting:**
   - `X-RateLimit-Remaining` header tracking
   - `rateLimit` query field for cost analysis
   - Retry with exponential backoff on rate limit

7. **Network Resilience:**
   - Timeout configuration
   - Retry logic for network errors (not GraphQL errors)
   - Graceful degradation for partial failures

### Code Structure

```typescript
// graphql-client.ts

interface GraphQLRequest<TVariables = Record<string, unknown>> {
  query: string;
  variables?: TVariables;
  operationName?: string;
}

interface GraphQLResponse<TData = unknown> {
  data: TData | null;
  errors?: GraphQLError[];
}

interface GraphQLError {
  message: string;
  locations?: Array<{ line: number; column: number }>;
  path?: Array<string | number>;
  extensions?: {
    code?: string;
    [key: string]: unknown;
  };
}

// Repository query types
interface GetRepositoryVariables {
  owner: string;
  name: string;
}

interface Repository {
  id: string;
  name: string;
  description: string | null;
  stargazerCount: number;
  forkCount: number;
  isPrivate: boolean;
}

interface GetRepositoryResponse {
  repository: Repository | null;
}

// Search query types
interface SearchRepositoriesVariables {
  query: string;
  first: number;
  after?: string | null;
}

interface SearchResultItemEdge {
  node: Repository;
}

interface PageInfo {
  hasNextPage: boolean;
  endCursor: string | null;
}

interface SearchConnection {
  repositoryCount: number;
  pageInfo: PageInfo;
  edges: SearchResultItemEdge[];
}

interface SearchRepositoriesResponse {
  search: SearchConnection;
}

// Mutation types
interface CreateIssueVariables {
  repositoryId: string;
  title: string;
  body?: string | null;
}

interface Issue {
  id: string;
  number: number;
  title: string;
  url: string;
}

interface CreateIssueResponse {
  createIssue: {
    issue: Issue;
  } | null;
}

class GitHubGraphQLClient {
  private readonly endpoint = 'https://api.github.com/graphql';
  private readonly token: string;
  private readonly timeout: number;

  constructor(token: string, options: { timeout?: number } = {}) {
    this.token = token;
    this.timeout = options.timeout ?? 10000;
  }

  async query<TData, TVariables = Record<string, unknown>>(
    request: GraphQLRequest<TVariables>
  ): Promise<TData> {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), this.timeout);

    try {
      const response = await fetch(this.endpoint, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${this.token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
        signal: controller.signal,
      });

      clearTimeout(timeoutId);

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const result: GraphQLResponse<TData> = await response.json();

      // CRITICAL: Check errors even when HTTP 200
      if (result.errors && result.errors.length > 0) {
        const error = result.errors[0];
        const code = error.extensions?.code ?? 'UNKNOWN';
        throw new GraphQLError(
          `GraphQL error [${code}]: ${error.message}`,
          result.errors
        );
      }

      if (result.data === null) {
        throw new Error('GraphQL response returned null data with no errors');
      }

      return result.data;
    } catch (err) {
      if (err instanceof GraphQLError) {
        throw err;
      }
      if (err instanceof Error && err.name === 'AbortError') {
        throw new Error(`Request timeout after ${this.timeout}ms`);
      }
      throw err;
    } finally {
      clearTimeout(timeoutId);
    }
  }

  async getRepository(owner: string, name: string): Promise<Repository> {
    const query = `
      query GetRepository($owner: String!, $name: String!) {
        repository(owner: $owner, name: $name) {
          id
          name
          description
          stargazerCount
          forkCount
          isPrivate
        }
      }
    `;

    const data = await this.query<GetRepositoryResponse, GetRepositoryVariables>({
      query,
      variables: { owner, name },
      operationName: 'GetRepository',
    });

    if (data.repository === null) {
      throw new Error(`Repository ${owner}/${name} not found`);
    }

    return data.repository;
  }

  async searchRepositories(
    searchQuery: string,
    pageSize: number = 10
  ): Promise<Repository[]> {
    const query = `
      query SearchRepositories($query: String!, $first: Int!, $after: String) {
        search(query: $query, type: REPOSITORY, first: $first, after: $after) {
          repositoryCount
          pageInfo {
            hasNextPage
            endCursor
          }
          edges {
            node {
              ... on Repository {
                id
                name
                description
                stargazerCount
                forkCount
                isPrivate
              }
            }
          }
        }
      }
    `;

    const results: Repository[] = [];
    let after: string | null = null;
    let hasNextPage = true;

    while (hasNextPage) {
      const data = await this.query<SearchRepositoriesResponse, SearchRepositoriesVariables>({
        query,
        variables: { query: searchQuery, first: pageSize, after },
        operationName: 'SearchRepositories',
      });

      // Extract repositories from edges
      for (const edge of data.search.edges) {
        results.push(edge.node);
      }

      // Check pagination
      hasNextPage = data.search.pageInfo.hasNextPage;
      after = data.search.pageInfo.endCursor;

      // Safety: prevent infinite loops
      if (results.length > 1000) {
        break;
      }
    }

    return results;
  }

  async createIssue(
    repositoryId: string,
    title: string,
    body?: string
  ): Promise<Issue> {
    const mutation = `
      mutation CreateIssue($repositoryId: ID!, $title: String!, $body: String) {
        createIssue(input: {repositoryId: $repositoryId, title: $title, body: $body}) {
          issue {
            id
            number
            title
            url
          }
        }
      }
    `;

    const data = await this.query<CreateIssueResponse, CreateIssueVariables>({
      query: mutation,
      variables: { repositoryId, title, body: body ?? null },
      operationName: 'CreateIssue',
    });

    if (data.createIssue === null) {
      throw new Error('Failed to create issue: mutation returned null');
    }

    return data.createIssue.issue;
  }
}

class GraphQLError extends Error {
  constructor(message: string, public readonly errors: GraphQLError[]) {
    super(message);
    this.name = 'GraphQLError';
  }
}
```

### Usage Example

```typescript
const client = new GitHubGraphQLClient(process.env.GITHUB_TOKEN!);

// Query with error handling
try {
  const repo = await client.getRepository('facebook', 'react');
  console.log(`${repo.name}: ${repo.stargazerCount} stars`);
} catch (err) {
  if (err instanceof GraphQLError) {
    console.error('GraphQL error:', err.errors[0].message);
  } else {
    console.error('Network error:', err.message);
  }
}

// Paginated search
const repos = await client.searchRepositories('language:typescript stars:>1000', 50);
console.log(`Found ${repos.length} repositories`);

// Mutation
const issue = await client.createIssue(
  'MDEwOlJlcG9zaXRvcnkxMjk2MjY5',
  'Bug: null pointer exception',
  'Steps to reproduce:\n1. ...'
);
console.log(`Created issue #${issue.number}: ${issue.url}`);
```

---

## Anti-Patterns (B01-B15)

Each anti-pattern demonstrates a specific GraphQL integration mistake that CTR rules should detect.

### B01 — No Error Checking Despite HTTP 200

**Location:** `examples/anti-patterns/graphql-b01-no-error-check.ts`

**Issue:** Assumes `response.ok` means success, ignores `errors` array.

```typescript
async function getRepository(owner: string, name: string) {
  const response = await fetch('https://api.github.com/graphql', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${process.env.GITHUB_TOKEN}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      query: `
        query {
          repository(owner: "${owner}", name: "${name}") {
            name
            stargazerCount
          }
        }
      `,
    }),
  });

  // WRONG: HTTP 200 doesn't mean GraphQL success
  if (!response.ok) {
    throw new Error(`HTTP error: ${response.status}`);
  }

  const data = await response.json();
  return data.data.repository; // Crashes if errors present
}
```

**CTR Violation:** No check for `result.errors` array before accessing `result.data`.

**Expected Detection:** `graphql-error-handling` rule should flag missing error array check.

---

### B02 — HTTP 200 But Not Checking Errors Array

**Location:** `examples/anti-patterns/graphql-b02-ignores-errors.ts`

**Issue:** Checks `response.ok` but not `errors` in JSON body.

```typescript
async function searchRepositories(query: string) {
  const response = await fetch('https://api.github.com/graphql', {
    method: 'POST',
    headers: { /* ... */ },
    body: JSON.stringify({
      query: `query { search(query: "${query}", type: REPOSITORY) { edges { node { ... on Repository { name } } } } }`,
    }),
  });

  if (!response.ok) {
    throw new Error('Request failed');
  }

  const result = await response.json();
  // WRONG: No check for result.errors
  return result.data.search.edges.map((e: any) => e.node);
}
```

**CTR Violation:** GraphQL errors go unchecked, partial failures ignored.

**Expected Detection:** Rule should require `if (result.errors)` check after `response.json()`.

---

### B03 — Hardcoded Query Variables (Injection Risk)

**Location:** `examples/anti-patterns/graphql-b03-injection.ts`

**Issue:** String interpolation in query instead of variables.

```typescript
async function getRepository(owner: string, name: string) {
  const query = `
    query {
      repository(owner: "${owner}", name: "${name}") {
        name
      }
    }
  `;
  // WRONG: owner/name not sanitized, injection possible
  const response = await fetch(/* ... */, { body: JSON.stringify({ query }) });
  // ...
}
```

**Attack Vector:**
```typescript
getRepository('facebook", name: "react") { __schema { types { name } } } #', 'dummy');
// Results in: repository(owner: "facebook", name: "react") { __schema { types { name } } } #", name: "dummy")
```

**CTR Violation:** GraphQL injection, should use `variables` field.

**Expected Detection:** Rule should flag string interpolation in GraphQL queries.

---

### B04 — Missing Timeout

**Location:** `examples/anti-patterns/graphql-b04-no-timeout.ts`

**Issue:** No `AbortSignal` or timeout for complex queries.

```typescript
async function complexSearch(query: string) {
  // WRONG: No timeout, can hang indefinitely
  const response = await fetch('https://api.github.com/graphql', {
    method: 'POST',
    headers: { /* ... */ },
    body: JSON.stringify({ query }),
  });
  // ...
}
```

**CTR Violation:** GraphQL queries can be expensive, timeout required.

**Expected Detection:** Rule should require `signal: AbortSignal` for GraphQL requests.

---

### B05 — Missing Required Variable

**Location:** `examples/anti-patterns/graphql-b05-missing-variable.ts`

**Issue:** TypeScript type doesn't enforce required variables (`!` in schema).

```typescript
interface SearchVariables {
  query?: string; // WRONG: query is required in schema ($query: String!)
  first?: number;
  after?: string;
}

async function search(vars: SearchVariables) {
  const response = await fetch('https://api.github.com/graphql', {
    method: 'POST',
    body: JSON.stringify({
      query: 'query($query: String!, $first: Int!) { search(query: $query, first: $first) { ... } }',
      variables: vars, // Runtime error if query missing
    }),
  });
  // ...
}
```

**CTR Violation:** TS type should match GraphQL schema nullability.

**Expected Detection:** Rule should compare TS interface to GraphQL schema, flag missing `!`.

---

### B06 — Missing Edge/Node Nesting in Types

**Location:** `examples/anti-patterns/graphql-b06-flat-type.ts`

**Issue:** TS type doesn't model `edges` → `node` structure.

```typescript
interface SearchResponse {
  search: {
    repositories: Array<{ name: string }>; // WRONG: should be edges[].node
  };
}

async function search(query: string) {
  const data: SearchResponse = await client.query(/* ... */);
  // This won't work at runtime:
  return data.search.repositories.map((r) => r.name);
}
```

**Actual API Response:**
```json
{
  "data": {
    "search": {
      "edges": [
        { "node": { "name": "repo1" } }
      ]
    }
  }
}
```

**CTR Violation:** Type doesn't match connection pattern.

**Expected Detection:** Rule should detect connection fields and enforce `edges`/`node` structure.

---

### B07 — No Null Checks on Nullable Fields

**Location:** `examples/anti-patterns/graphql-b07-no-null-check.ts`

**Issue:** Accessing nullable fields without guards.

```typescript
interface Repository {
  name: string;
  description: string; // WRONG: description is nullable in schema
}

async function getRepo(owner: string, name: string) {
  const data = await client.getRepository(owner, name);
  console.log(data.description.toUpperCase()); // Crashes if null
}
```

**CTR Violation:** GraphQL schema marks field as nullable, TS type doesn't.

**Expected Detection:** Rule should parse schema, enforce `| null` for nullable fields.

---

### B08 — Missing Union Type Handling

**Location:** `examples/anti-patterns/graphql-b08-no-union-check.ts`

**Issue:** Search results can be `Repository | Organization | Issue`, code assumes Repository.

```typescript
async function search(query: string) {
  const data = await client.query(/* ... */);
  return data.search.edges.map((edge: any) => {
    // WRONG: Assumes all results are Repository
    return {
      name: edge.node.name,
      stars: edge.node.stargazerCount, // Crashes if node is Issue (no stargazerCount)
    };
  });
}
```

**CTR Violation:** No type guard for union types.

**Expected Detection:** Rule should detect union types in schema, require fragment type conditions.

---

### B09 — Partial Success Ignored

**Location:** `examples/anti-patterns/graphql-b09-partial-failure.ts`

**Issue:** Both `data` and `errors` can exist, code ignores errors if data present.

```typescript
async function batchQuery() {
  const response = await fetch(/* ... */);
  const result = await response.json();

  if (result.errors) {
    throw new Error('Query failed');
  }

  return result.data; // WRONG: Errors can exist even with partial data
}
```

**Real Scenario:**
```json
{
  "data": {
    "repository": { "name": "my-repo", "description": null }
  },
  "errors": [
    { "path": ["repository", "description"], "message": "Field forbidden" }
  ]
}
```

**CTR Violation:** Should log errors even when data present.

**Expected Detection:** Rule should detect pattern of throwing on `errors` without checking if `data` also exists.

---

### B10 — No Rate Limit Handling

**Location:** `examples/anti-patterns/graphql-b10-no-rate-limit.ts`

**Issue:** Doesn't check `X-RateLimit-Remaining` or retry on rate limit errors.

```typescript
async function bulkFetch(repos: string[]) {
  for (const repo of repos) {
    // WRONG: No rate limit checks, will hit quota
    await client.getRepository('owner', repo);
  }
}
```

**CTR Violation:** GraphQL has complex cost model, should track remaining quota.

**Expected Detection:** Rule should require rate limit header checks or `rateLimit` query field.

---

### B11 — Ignoring PageInfo Fields

**Location:** `examples/anti-patterns/graphql-b11-no-pageinfo.ts`

**Issue:** Fetches first page, ignores `hasNextPage`.

```typescript
async function getAllRepos(query: string) {
  const data = await client.query({
    query: 'query($query: String!) { search(query: $query, first: 10) { edges { node { ... } } } }',
    variables: { query },
  });

  // WRONG: Returns only first 10, ignores pagination
  return data.search.edges.map((e: any) => e.node);
}
```

**CTR Violation:** Incomplete data fetch, should loop while `hasNextPage === true`.

**Expected Detection:** Rule should detect `pageInfo` in query but no check in code.

---

### B12 — Mutation Without Error Check

**Location:** `examples/anti-patterns/graphql-b12-mutation-no-check.ts`

**Issue:** Mutation can return null on failure, code doesn't check.

```typescript
async function createIssue(repoId: string, title: string) {
  const data = await client.query({
    query: 'mutation($repoId: ID!, $title: String!) { createIssue(input: {repositoryId: $repoId, title: $title}) { issue { id } } }',
    variables: { repoId, title },
  });

  // WRONG: createIssue can be null if mutation fails
  return data.createIssue.issue.id; // Crashes on null
}
```

**CTR Violation:** Mutation fields are nullable, should check before access.

**Expected Detection:** Rule should enforce null checks on mutation response fields.

---

### B13 — Fragment Without Type Condition

**Location:** `examples/anti-patterns/graphql-b13-fragment-no-type.ts`

**Issue:** Uses fragment on union without `... on Type` condition.

```typescript
const query = `
  query {
    search(query: "react", type: REPOSITORY, first: 10) {
      edges {
        node {
          ... RepositoryFields
        }
      }
    }
  }

  fragment RepositoryFields on SearchResultItem {
    name
    stargazerCount
  }
`;
// WRONG: SearchResultItem is union, fragment needs "... on Repository"
```

**CTR Violation:** Fragment on union/interface requires type condition.

**Expected Detection:** Rule should parse schema, detect union types, require type conditions in fragments.

---

### B14 — Infinite Pagination Loop

**Location:** `examples/anti-patterns/graphql-b14-infinite-loop.ts`

**Issue:** Pagination loop doesn't check `hasNextPage`, relies only on cursor.

```typescript
async function fetchAll(query: string) {
  let after: string | null = null;
  const results = [];

  while (true) { // WRONG: No exit condition
    const data = await client.query({
      query: '...',
      variables: { query, first: 10, after },
    });

    results.push(...data.search.edges.map((e: any) => e.node));
    after = data.search.pageInfo.endCursor; // Infinite if endCursor keeps changing

    // Should check: if (!data.search.pageInfo.hasNextPage) break;
  }

  return results;
}
```

**CTR Violation:** Pagination loop must respect `hasNextPage`.

**Expected Detection:** Rule should detect `while (true)` with `endCursor` but no `hasNextPage` check.

---

### B15 — No Operation Name

**Location:** `examples/anti-patterns/graphql-b15-no-operation-name.ts`

**Issue:** Anonymous queries make debugging harder.

```typescript
const query = `
  query {
    viewer { login }
  }
`;
// WRONG: Should be "query GetViewer { viewer { login } }"
```

**CTR Violation:** All queries should have operation names for tracing.

**Expected Detection:** Rule should detect unnamed queries/mutations.

---

## Test Plan

### Test Structure

```
validation-sets/60-graphql-api/
├── perfect/
│   ├── graphql-client.ts          # Full implementation
│   ├── types.ts                   # All GraphQL types
│   └── fragments.ts               # Reusable fragments
├── anti-patterns/
│   ├── graphql-b01-no-error-check.ts
│   ├── graphql-b02-ignores-errors.ts
│   ├── graphql-b03-injection.ts
│   ├── graphql-b04-no-timeout.ts
│   ├── graphql-b05-missing-variable.ts
│   ├── graphql-b06-flat-type.ts
│   ├── graphql-b07-no-null-check.ts
│   ├── graphql-b08-no-union-check.ts
│   ├── graphql-b09-partial-failure.ts
│   ├── graphql-b10-no-rate-limit.ts
│   ├── graphql-b11-no-pageinfo.ts
│   ├── graphql-b12-mutation-no-check.ts
│   ├── graphql-b13-fragment-no-type.ts
│   ├── graphql-b14-infinite-loop.ts
│   └── graphql-b15-no-operation-name.ts
├── schema.graphql                 # GitHub GraphQL v4 schema subset
└── test-cases.json                # Expected detections
```

### Product Tests

#### PT-01: PERFECT Implementation Passes All Checks

**Given:**
- `perfect/graphql-client.ts` with full error handling
- All nullable fields typed correctly
- Proper pagination logic

**When:**
- Super-lint analyzes the perfect implementation

**Then:**
- Zero violations reported
- All GraphQL best practices satisfied

**Pass Criteria:**
- Exit code 0
- No warnings or errors in output

---

#### PT-02: Error Handling Violations Detected (B01, B02, B09)

**Given:**
- B01: No error check after `response.json()`
- B02: Checks `response.ok` but not `result.errors`
- B09: Throws on `errors` without checking partial success

**When:**
- Super-lint analyzes anti-patterns

**Then:**
- B01: `graphql-error-handling` rule flags missing error check
- B02: Rule flags missing `if (result.errors)` check
- B09: Rule flags improper partial failure handling

**Pass Criteria:**
- 3 violations detected (one per file)
- Error messages explain GraphQL error format
- Suggests adding `if (result.errors && result.errors.length > 0)` check

---

#### PT-03: GraphQL Injection Detected (B03)

**Given:**
- B03: String interpolation in query: `repository(owner: "${owner}")`

**When:**
- Super-lint analyzes the query construction

**Then:**
- `graphql-injection` rule flags interpolated variables
- Suggests using `variables` field instead

**Pass Criteria:**
- 1 violation detected
- Error message shows unsafe interpolation
- Fix suggestion includes variables example

---

#### PT-04: Type Safety Violations Detected (B05, B06, B07)

**Given:**
- B05: Optional field where schema requires non-null
- B06: Flat type instead of edges/node structure
- B07: Non-null type for nullable schema field

**When:**
- Super-lint compares TS types to GraphQL schema

**Then:**
- B05: Rule flags missing `!` enforcement
- B06: Rule flags incorrect connection type structure
- B07: Rule flags missing `| null` union

**Pass Criteria:**
- 3 violations detected
- Each includes schema reference
- Suggests correct TypeScript type

---

#### PT-05: Pagination Violations Detected (B11, B14)

**Given:**
- B11: Query includes `pageInfo` but code doesn't check `hasNextPage`
- B14: Pagination loop with `while (true)` and no exit condition

**When:**
- Super-lint analyzes pagination logic

**Then:**
- B11: Rule flags incomplete pagination implementation
- B14: Rule flags potential infinite loop

**Pass Criteria:**
- 2 violations detected
- Suggestions include `hasNextPage` check
- Warning about incomplete data fetching

---

#### PT-06: Union Type Handling Violation (B08)

**Given:**
- B08: Search result typed as `Repository` but schema says `SearchResultItem` union

**When:**
- Super-lint analyzes union type usage

**Then:**
- Rule flags missing type guard
- Suggests using `... on Repository` fragment

**Pass Criteria:**
- 1 violation detected
- Error explains union type at runtime
- Shows fragment type condition example

---

#### PT-07: Mutation Safety Violation (B12)

**Given:**
- B12: Mutation response accessed without null check

**When:**
- Super-lint analyzes mutation code

**Then:**
- Rule flags unsafe mutation response access
- Notes that mutation fields are nullable

**Pass Criteria:**
- 1 violation detected
- Suggests null check before accessing nested fields
- References GraphQL mutation spec

---

#### PT-08: Timeout and Rate Limiting Violations (B04, B10)

**Given:**
- B04: No timeout on GraphQL fetch
- B10: Bulk operations without rate limit checks

**When:**
- Super-lint analyzes network resilience

**Then:**
- B04: Rule flags missing AbortSignal
- B10: Rule flags missing rate limit tracking

**Pass Criteria:**
- 2 violations detected
- B04 suggests timeout implementation
- B10 suggests `X-RateLimit-Remaining` header check

---

#### PT-09: Fragment and Operation Naming Violations (B13, B15)

**Given:**
- B13: Fragment on union without type condition
- B15: Anonymous query (no operation name)

**When:**
- Super-lint analyzes GraphQL queries

**Then:**
- B13: Rule flags fragment missing type condition
- B15: Rule flags anonymous operation

**Pass Criteria:**
- 2 violations detected
- B13 shows correct fragment syntax
- B15 suggests adding operation name

---

### Technical Tests

#### TT-01: Schema Parsing

**Test:** Parse `schema.graphql` and extract type information.

**Implementation:**
```typescript
// test/graphql-schema-parser.test.ts
import { parseSchema } from '../src/rules/graphql/schema-parser';

test('parses repository query type', () => {
  const schema = parseSchema('schema.graphql');
  const repoType = schema.getType('Repository');

  expect(repoType).toBeDefined();
  expect(repoType.fields.description.nullable).toBe(true);
  expect(repoType.fields.name.nullable).toBe(false);
});

test('detects connection types', () => {
  const schema = parseSchema('schema.graphql');
  const searchType = schema.getType('SearchResultItemConnection');

  expect(searchType.isConnection).toBe(true);
  expect(searchType.fields.edges).toBeDefined();
  expect(searchType.fields.pageInfo).toBeDefined();
});

test('identifies union types', () => {
  const schema = parseSchema('schema.graphql');
  const resultType = schema.getType('SearchResultItem');

  expect(resultType.kind).toBe('UNION');
  expect(resultType.possibleTypes).toContain('Repository');
  expect(resultType.possibleTypes).toContain('Issue');
  expect(resultType.possibleTypes).toContain('Organization');
});
```

**Pass Criteria:**
- All schema types correctly parsed
- Nullability correctly detected
- Connections identified by pattern

---

#### TT-02: Query Variable Extraction

**Test:** Extract variable definitions from GraphQL queries.

**Implementation:**
```typescript
// test/graphql-query-parser.test.ts
import { extractVariables } from '../src/rules/graphql/query-parser';

test('extracts required variables', () => {
  const query = 'query GetRepo($owner: String!, $name: String!) { ... }';
  const vars = extractVariables(query);

  expect(vars).toEqual([
    { name: 'owner', type: 'String', required: true },
    { name: 'name', type: 'String', required: true },
  ]);
});

test('extracts optional variables', () => {
  const query = 'query Search($query: String!, $after: String) { ... }';
  const vars = extractVariables(query);

  expect(vars.find(v => v.name === 'after')).toMatchObject({
    name: 'after',
    type: 'String',
    required: false,
  });
});
```

**Pass Criteria:**
- Required variables detected via `!` suffix
- Optional variables correctly identified
- Type names extracted

---

#### TT-03: Error Handling Rule

**Test:** Detect missing GraphQL error checks.

**Implementation:**
```typescript
// test/rules/graphql-error-handling.test.ts
import { graphqlErrorHandling } from '../src/rules/graphql-error-handling';

test('flags missing error check', () => {
  const code = `
    const response = await fetch('/graphql', { method: 'POST', body: JSON.stringify({ query }) });
    const result = await response.json();
    return result.data; // No error check
  `;

  const violations = graphqlErrorHandling.check(code);

  expect(violations).toHaveLength(1);
  expect(violations[0].message).toContain('GraphQL errors not checked');
  expect(violations[0].line).toBe(3);
});

test('passes when errors checked', () => {
  const code = `
    const result = await response.json();
    if (result.errors && result.errors.length > 0) {
      throw new GraphQLError(result.errors);
    }
    return result.data;
  `;

  const violations = graphqlErrorHandling.check(code);
  expect(violations).toHaveLength(0);
});
```

**Pass Criteria:**
- Detects missing error checks
- Allows various error check patterns
- No false positives on proper implementations

---

#### TT-04: Type Nullability Matching

**Test:** Compare TypeScript types to GraphQL schema.

**Implementation:**
```typescript
// test/rules/graphql-type-safety.test.ts
import { validateTypeNullability } from '../src/rules/graphql-type-safety';

test('flags non-null type for nullable field', () => {
  const tsType = 'interface Repository { description: string }';
  const schemaField = { name: 'description', type: 'String', nullable: true };

  const violations = validateTypeNullability(tsType, schemaField);

  expect(violations).toHaveLength(1);
  expect(violations[0].message).toContain('should be nullable');
});

test('passes when nullability matches', () => {
  const tsType = 'interface Repository { description: string | null }';
  const schemaField = { name: 'description', type: 'String', nullable: true };

  const violations = validateTypeNullability(tsType, schemaField);
  expect(violations).toHaveLength(0);
});
```

**Pass Criteria:**
- Detects nullable field typed as non-null
- Detects non-null field typed as nullable
- Handles union types correctly

---

#### TT-05: Connection Pattern Detection

**Test:** Detect incorrect connection type structures.

**Implementation:**
```typescript
// test/rules/graphql-connection-pattern.test.ts
import { validateConnectionPattern } from '../src/rules/graphql-connection-pattern';

test('flags flat array for connection field', () => {
  const tsType = 'interface SearchResult { search: { repositories: Repository[] } }';
  const schemaType = 'SearchResultItemConnection';

  const violations = validateConnectionPattern(tsType, schemaType);

  expect(violations).toHaveLength(1);
  expect(violations[0].message).toContain('edges');
  expect(violations[0].message).toContain('node');
});

test('passes when edges/node present', () => {
  const tsType = 'interface SearchResult { search: { edges: Array<{ node: Repository }> } }';
  const schemaType = 'SearchResultItemConnection';

  const violations = validateConnectionPattern(tsType, schemaType);
  expect(violations).toHaveLength(0);
});
```

**Pass Criteria:**
- Detects missing `edges` field
- Detects missing `node` field
- Allows correct nesting

---

#### TT-06: Pagination Logic Validation

**Test:** Detect incomplete pagination implementations.

**Implementation:**
```typescript
// test/rules/graphql-pagination.test.ts
import { validatePagination } from '../src/rules/graphql-pagination';

test('flags query with pageInfo but no hasNextPage check', () => {
  const code = `
    const data = await client.query({ query: 'query { search(first: 10) { pageInfo { hasNextPage } edges { node { id } } } }' });
    return data.search.edges.map(e => e.node);
  `;

  const violations = validatePagination(code);

  expect(violations).toHaveLength(1);
  expect(violations[0].message).toContain('hasNextPage');
});

test('passes when hasNextPage checked', () => {
  const code = `
    let hasNext = true;
    while (hasNext) {
      const data = await client.query({ ... });
      results.push(...data.search.edges);
      hasNext = data.search.pageInfo.hasNextPage;
    }
  `;

  const violations = validatePagination(code);
  expect(violations).toHaveLength(0);
});
```

**Pass Criteria:**
- Detects `pageInfo` in query but no check in code
- Allows various pagination patterns
- Flags infinite loops without exit

---

### UAT Script

**Location:** `scripts/test-validation-set-60.sh`

```bash
#!/usr/bin/env bash
set -euo pipefail

echo "=== UAT: GraphQL API Validation Set ==="

VALIDATION_DIR="validation-sets/60-graphql-api"

# Test 1: PERFECT passes
echo "[1/9] Testing PERFECT implementation..."
if super-lint "$VALIDATION_DIR/perfect/graphql-client.ts" | grep -q "0 violations"; then
  echo "✓ PERFECT passes"
else
  echo "✗ PERFECT should have 0 violations"
  exit 1
fi

# Test 2: Error handling violations
echo "[2/9] Testing error handling violations..."
VIOLATIONS=$(super-lint "$VALIDATION_DIR/anti-patterns/graphql-b01-no-error-check.ts" \
  "$VALIDATION_DIR/anti-patterns/graphql-b02-ignores-errors.ts" \
  "$VALIDATION_DIR/anti-patterns/graphql-b09-partial-failure.ts" \
  --format json | jq '.violations | length')

if [ "$VIOLATIONS" -eq 3 ]; then
  echo "✓ Detected 3 error handling violations"
else
  echo "✗ Expected 3 violations, got $VIOLATIONS"
  exit 1
fi

# Test 3: Injection
echo "[3/9] Testing GraphQL injection..."
if super-lint "$VALIDATION_DIR/anti-patterns/graphql-b03-injection.ts" \
  | grep -q "graphql-injection"; then
  echo "✓ Injection detected"
else
  echo "✗ Should detect injection"
  exit 1
fi

# Test 4: Type safety
echo "[4/9] Testing type safety violations..."
VIOLATIONS=$(super-lint "$VALIDATION_DIR/anti-patterns/graphql-b05-missing-variable.ts" \
  "$VALIDATION_DIR/anti-patterns/graphql-b06-flat-type.ts" \
  "$VALIDATION_DIR/anti-patterns/graphql-b07-no-null-check.ts" \
  --format json | jq '.violations | length')

if [ "$VIOLATIONS" -eq 3 ]; then
  echo "✓ Detected 3 type safety violations"
else
  echo "✗ Expected 3 violations, got $VIOLATIONS"
  exit 1
fi

# Test 5: Pagination
echo "[5/9] Testing pagination violations..."
VIOLATIONS=$(super-lint "$VALIDATION_DIR/anti-patterns/graphql-b11-no-pageinfo.ts" \
  "$VALIDATION_DIR/anti-patterns/graphql-b14-infinite-loop.ts" \
  --format json | jq '.violations | length')

if [ "$VIOLATIONS" -eq 2 ]; then
  echo "✓ Detected 2 pagination violations"
else
  echo "✗ Expected 2 violations, got $VIOLATIONS"
  exit 1
fi

# Test 6: Union types
echo "[6/9] Testing union type violation..."
if super-lint "$VALIDATION_DIR/anti-patterns/graphql-b08-no-union-check.ts" \
  | grep -q "union"; then
  echo "✓ Union type violation detected"
else
  echo "✗ Should detect union type issue"
  exit 1
fi

# Test 7: Mutations
echo "[7/9] Testing mutation violation..."
if super-lint "$VALIDATION_DIR/anti-patterns/graphql-b12-mutation-no-check.ts" \
  | grep -q "nullable"; then
  echo "✓ Mutation null check violation detected"
else
  echo "✗ Should detect mutation safety issue"
  exit 1
fi

# Test 8: Timeout and rate limiting
echo "[8/9] Testing timeout and rate limit violations..."
VIOLATIONS=$(super-lint "$VALIDATION_DIR/anti-patterns/graphql-b04-no-timeout.ts" \
  "$VALIDATION_DIR/anti-patterns/graphql-b10-no-rate-limit.ts" \
  --format json | jq '.violations | length')

if [ "$VIOLATIONS" -eq 2 ]; then
  echo "✓ Detected 2 network resilience violations"
else
  echo "✗ Expected 2 violations, got $VIOLATIONS"
  exit 1
fi

# Test 9: Fragment and naming
echo "[9/9] Testing fragment and naming violations..."
VIOLATIONS=$(super-lint "$VALIDATION_DIR/anti-patterns/graphql-b13-fragment-no-type.ts" \
  "$VALIDATION_DIR/anti-patterns/graphql-b15-no-operation-name.ts" \
  --format json | jq '.violations | length')

if [ "$VIOLATIONS" -eq 2 ]; then
  echo "✓ Detected 2 GraphQL quality violations"
else
  echo "✗ Expected 2 violations, got $VIOLATIONS"
  exit 1
fi

echo ""
echo "=== All GraphQL API Tests Passed ==="
```

---

## Test Status

| Test ID | Description | Status | Notes |
|---------|-------------|--------|-------|
| PT-01 | PERFECT passes | ⏳ Not implemented | |
| PT-02 | Error handling detected | ⏳ Not implemented | |
| PT-03 | Injection detected | ⏳ Not implemented | |
| PT-04 | Type safety detected | ⏳ Not implemented | |
| PT-05 | Pagination detected | ⏳ Not implemented | |
| PT-06 | Union types detected | ⏳ Not implemented | |
| PT-07 | Mutation safety detected | ⏳ Not implemented | |
| PT-08 | Network resilience detected | ⏳ Not implemented | |
| PT-09 | Fragment/naming detected | ⏳ Not implemented | |
| TT-01 | Schema parsing | ⏳ Not implemented | |
| TT-02 | Variable extraction | ⏳ Not implemented | |
| TT-03 | Error handling rule | ⏳ Not implemented | |
| TT-04 | Type nullability | ⏳ Not implemented | |
| TT-05 | Connection pattern | ⏳ Not implemented | |
| TT-06 | Pagination logic | ⏳ Not implemented | |

---

## Summary

This validation set comprehensively tests CTR rules on GraphQL APIs:

- **15 anti-patterns** covering all major GraphQL pitfalls
- **1 perfect implementation** demonstrating best practices
- **9 product tests** validating detection of each violation category
- **6 technical tests** ensuring rule correctness
- **1 UAT script** for automated validation

**Key Coverage Areas:**
1. GraphQL-specific error handling (HTTP 200 with errors)
2. Type safety (nullability, unions, connections)
3. Pagination (edges/nodes, pageInfo, hasNextPage)
4. Security (injection prevention, rate limiting)
5. Code quality (operation names, fragments)

**Expected Outcome:**
- PERFECT: 0 violations
- Each anti-pattern: 1 violation
- UAT script: All 9 tests pass

