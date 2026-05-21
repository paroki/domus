# ADR-001: API Response Structure Contract

**Status:** Accepted
**Created:** 2026-05-21
**Revised:** 2026-05-21
**Deciders:** Anthonius Munthi
**Context:** API (`api/`)

---

## Revision History

| Version | Date | Changes |
|---|---|---|
| 1.0 | 2026-05-21 | Initial acceptance |
| 1.1 | 2026-05-21 | Renamed from "Standard" to "Contract" to emphasize binding nature |

---

## 1. Context

The Domus API is a GoLang-based RESTful service built on **GoFiber v3**, consumed by:

- A TypeScript React PWA (primary dashboard)
- Potential future mobile clients
- Potential third-party parish integrations

Without a formalised response contract, each endpoint risks returning inconsistent shapes, making frontend parsing fragile and error diagnosis difficult — particularly in a multi-parish deployment context where support staff must diagnose issues remotely.

---

## 2. Decision

All API endpoints **MUST** return a unified JSON envelope with the following top-level structure:

```json
{
  "success": true,
  "data": {},
  "meta": {},
  "error": null
}
```

### 2.1 Field Definitions

| Field     | Type              | Nullable | Description                                                                   |
|-----------|-------------------|----------|-------------------------------------------------------------------------------|
| `success` | `boolean`         | No       | `true` for 2xx responses, `false` for all error responses                     |
| `data`    | `object \| array` | Yes      | Primary response payload. `null` on error responses                           |
| `meta`    | `object`          | Yes      | Auxiliary metadata (pagination, request tracing). `null` when not applicable  |
| `error`   | `object`          | Yes      | Structured error details. `null` on success responses                         |

---

## 3. Response Schemas

### 3.1 Success Response

```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "St. Michael Parish",
    "diocese": "Archdiocese of Jakarta"
  },
  "meta": {
    "request_id": "req_01HZ9K3M2P",
    "timestamp": "2026-05-21T08:00:00Z"
  },
  "error": null
}
```

### 3.2 Paginated Collection Response

```json
{
  "success": true,
  "data": [
    { "id": "...", "name": "..." },
    { "id": "...", "name": "..." }
  ],
  "meta": {
    "request_id": "req_01HZ9K3M2P",
    "timestamp": "2026-05-21T08:00:00Z",
    "pagination": {
      "page":        1,
      "per_page":    20,
      "total_pages": 5,
      "total_items": 98
    }
  },
  "error": null
}
```

#### Pagination Strategy

Domus uses **offset-based pagination**. Cursor-based pagination is not required at this scale and would add unnecessary complexity for parish-sized datasets.

**Query Parameters**

| Parameter  | Type    | Default | Max   | Description              |
|------------|---------|---------|-------|--------------------------|
| `page`     | integer | `1`     | —     | 1-indexed page number    |
| `per_page` | integer | `20`    | `100` | Number of items per page |

**Rules**

- `page` below `1` MUST be rejected with `400 Bad Request`.
- `per_page` above `100` MUST be capped at `100` silently — no error returned.
- `per_page` of `0` MUST be rejected with `400 Bad Request`.
- All collection endpoints MUST support pagination query parameters, even if the total dataset is currently small.
- `meta.pagination` MUST be present on every collection response, including single-page results.

**GoLang Helper**

```go
// PaginatedOK writes a 200 paginated collection response.
func PaginatedOK[T any](c fiber.Ctx, data []T, p Pagination) error {
    m := newMeta(c)
    m.Pagination = &p
    return c.Status(fiber.StatusOK).JSON(Envelope[[]T]{
        Success: true,
        Data:    data,
        Meta:    m,
        Error:   nil,
    })
}
```

**TypeScript Helper**

```typescript
export function isPaginated<T>(meta: Meta | null): meta is Meta & { pagination: Pagination } {
  return meta?.pagination != null;
}
```

### 3.3 Error Response

```json
{
  "success": false,
  "data": null,
  "meta": {
    "request_id": "req_01HZ9K3M2P",
    "timestamp": "2026-05-21T08:00:00Z"
  },
  "error": {
    "code":    "DOMUS-AUTH-001",
    "message": "Authentication token has expired.",
    "details": null
  }
}
```

### 3.4 Validation Error Response

```json
{
  "success": false,
  "data": null,
  "meta": {
    "request_id": "req_01HZ9K3M2P",
    "timestamp": "2026-05-21T08:00:00Z"
  },
  "error": {
    "code":    "DOMUS-VAL-001",
    "message": "Request validation failed.",
    "details": [
      { "field": "email",      "issue": "Must be a valid email address." },
      { "field": "birth_date", "issue": "Cannot be a future date."       }
    ]
  }
}
```

---

## 4. Error Code Convention

Domain error codes follow the pattern: `DOMUS-{DOMAIN}-{SEQ}`

| Segment  | Description                              | Example             |
|----------|------------------------------------------|---------------------|
| `DOMUS`  | System prefix (constant)                 | `DOMUS`             |
| `DOMAIN` | Functional domain (3–5 chars, uppercase) | `AUTH`, `VAL`, `PAR`, `MBR` |
| `SEQ`    | Zero-padded sequence (3 digits)          | `001`, `002`        |

### 4.1 Reserved Domain Prefixes

| Prefix | Domain                          |
|--------|---------------------------------|
| `AUTH` | Authentication & Authorisation  |
| `VAL`  | Input Validation                |
| `PAR`  | Parish Management               |
| `MBR`  | Member / Parishioner            |
| `FIN`  | Finance & Contributions         |
| `EVT`  | Events & Liturgical Calendar    |
| `DOC`  | Documents & Sacramental Records |
| `SYS`  | System / Infrastructure         |

---

## 5. HTTP Status Code Conventions

| HTTP Status                  | Scenario                                        |
|------------------------------|-------------------------------------------------|
| `200 OK`                     | Successful GET, PUT, PATCH                      |
| `201 Created`                | Successful POST resulting in resource creation  |
| `204 No Content`             | Successful DELETE (no response body)            |
| `400 Bad Request`            | Validation failure (`DOMUS-VAL-*`)              |
| `401 Unauthorized`           | Missing or invalid auth token (`DOMUS-AUTH-*`)  |
| `403 Forbidden`              | Authenticated but insufficient permission       |
| `404 Not Found`              | Resource does not exist                         |
| `409 Conflict`               | Duplicate resource or state conflict            |
| `422 Unprocessable Entity`   | Semantically invalid request                    |
| `500 Internal Server Error`  | Unhandled server-side failure                   |

> **Rule:** HTTP status codes MUST accurately reflect the response semantic. Using `200` for error responses is strictly prohibited.

---

## 6. GoLang Reference Implementation

### 6.1 File Location

```
internal/delivery/gofiber/response.go
```

### 6.2 Request ID Strategy

Domus uses GoFiber's built-in `requestid` middleware. The middleware injects a unique ID per request, retrievable via `requestid.FromContext(c)`. This is the canonical method — manual `c.Locals()` assignment is non-compliant.

**Middleware registration (main.go or app bootstrap):**

```go
import "github.com/gofiber/fiber/v3/middleware/requestid"

app.Use(requestid.New())
```

### 6.3 Response Structs

```go
package response

import "time"

// Envelope is the top-level wrapper for all API responses.
type Envelope[T any] struct {
    Success bool       `json:"success"`
    Data    T          `json:"data"`
    Meta    *Meta      `json:"meta"`
    Error   *ErrorBody `json:"error"`
}

// Meta carries auxiliary metadata included in every response.
type Meta struct {
    RequestID  string      `json:"request_id"`
    Timestamp  time.Time   `json:"timestamp"`
    Pagination *Pagination `json:"pagination,omitempty"`
}

// Pagination carries paging information for collection responses.
type Pagination struct {
    Page       int `json:"page"`
    PerPage    int `json:"per_page"`
    TotalPages int `json:"total_pages"`
    TotalItems int `json:"total_items"`
}

// ErrorBody carries structured error information.
type ErrorBody struct {
    Code    string       `json:"code"`
    Message string       `json:"message"`
    Details []FieldError `json:"details"`
}

// FieldError represents a single field-level validation failure.
type FieldError struct {
    Field string `json:"field"`
    Issue string `json:"issue"`
}
```

### 6.4 Helper Constructors

```go
package response

import (
    "time"

    "github.com/gofiber/fiber/v3"
    "github.com/gofiber/fiber/v3/middleware/requestid"
)

func newMeta(c fiber.Ctx) *Meta {
    return &Meta{
        RequestID: requestid.FromContext(c),
        Timestamp: time.Now().UTC(),
    }
}

// OK writes a 200 success response.
func OK[T any](c fiber.Ctx, data T) error {
    return c.Status(fiber.StatusOK).JSON(Envelope[T]{
        Success: true,
        Data:    data,
        Meta:    newMeta(c),
        Error:   nil,
    })
}

// Created writes a 201 success response.
func Created[T any](c fiber.Ctx, data T) error {
    return c.Status(fiber.StatusCreated).JSON(Envelope[T]{
        Success: true,
        Data:    data,
        Meta:    newMeta(c),
        Error:   nil,
    })
}

// NoContent writes a 204 response with no body.
func NoContent(c fiber.Ctx) error {
    return c.SendStatus(fiber.StatusNoContent)
}

// PaginatedOK writes a 200 paginated collection response.
func PaginatedOK[T any](c fiber.Ctx, data []T, p Pagination) error {
    m := newMeta(c)
    m.Pagination = &p
    return c.Status(fiber.StatusOK).JSON(Envelope[[]T]{
        Success: true,
        Data:    data,
        Meta:    m,
        Error:   nil,
    })
}

// Fail writes a structured error response.
func Fail(c fiber.Ctx, status int, code, message string, details []FieldError) error {
    return c.Status(status).JSON(Envelope[any]{
        Success: false,
        Data:    nil,
        Meta:    newMeta(c),
        Error: &ErrorBody{
            Code:    code,
            Message: message,
            Details: details,
        },
    })
}
```

### 6.5 Usage Example

```go
import (
    "net/http"

    "github.com/gofiber/fiber/v3"
    "github.com/paroki/domus/api/internal/delivery/gofiber/response"
)

func GetParish(c fiber.Ctx) error {
    parish, err := parishService.FindByID(c.Params("id"))
    if err != nil {
        return response.Fail(c, fiber.StatusNotFound, "DOMUS-PAR-001", "Parish not found.", nil)
    }
    return response.OK(c, parish)
}

func DeleteParish(c fiber.Ctx) error {
    if err := parishService.Delete(c.Params("id")); err != nil {
        return response.Fail(c, fiber.StatusInternalServerError, "DOMUS-SYS-001", "Failed to delete parish.", nil)
    }
    return response.NoContent(c)
}
```

---

## 7. TypeScript Reference Implementation

### 7.1 Generic Response Types

```typescript
// types/api.ts

export interface Meta {
  request_id:  string;
  timestamp:   string;
  pagination?: Pagination;
}

export interface Pagination {
  page:        number;
  per_page:    number;
  total_pages: number;
  total_items: number;
}

export interface FieldError {
  field: string;
  issue: string;
}

export interface ApiError {
  code:     string;
  message:  string;
  details:  FieldError[] | null;
}

export interface ApiResponse<T> {
  success: boolean;
  data:    T | null;
  meta:    Meta | null;
  error:   ApiError | null;
}
```

### 7.2 Fetch Wrapper

```typescript
// lib/apiClient.ts

import type { ApiResponse, FieldError } from '@/types/api';

export class ApiException extends Error {
  constructor(
    public readonly code: string,
    message: string,
    public readonly details: FieldError[] | null = null,
  ) {
    super(message);
    this.name = 'ApiException';
  }
}

export async function apiFetch<T>(
  url: string,
  options?: RequestInit,
): Promise<ApiResponse<T>> {
  const res  = await fetch(url, options);
  const body = await res.json() as ApiResponse<T>;

  if (!body.success && body.error) {
    throw new ApiException(body.error.code, body.error.message, body.error.details);
  }

  return body;
}
```

### 7.3 Pagination Guard

```typescript
export function isPaginated(meta: Meta | null): meta is Meta & { pagination: Pagination } {
  return meta?.pagination != null;
}
```

---

## 8. Consequences

### 8.1 Positive

- **Consistency** — All consumers parse a single, predictable shape regardless of endpoint.
- **Debuggability** — Domain error codes (`DOMUS-AUTH-001`) allow rapid triage without log access.
- **Type Safety** — GoLang generics and TypeScript generics enforce correctness at compile time.
- **Extensibility** — The `meta` field accommodates future concerns (rate limiting, versioning) without breaking changes.
- **i18n Ready** — Error codes are locale-agnostic; message translation can be layered on the frontend using the code as a key.

### 8.2 Negative / Trade-offs

- **Boilerplate** — All endpoints must use the helper constructors; direct `c.JSON()` calls are non-compliant and must be caught in code review.
- **`204 No Content` Exception** — DELETE endpoints return no body via `response.NoContent()`; consumers must handle the absence of an envelope for this status code.
- **Fiber Coupling** — The `response` package is tightly coupled to `fiber.Ctx` and must not be imported outside the `internal/delivery/gofiber/` layer.

---

## 9. Compliance

All new API endpoints introduced after the acceptance date of this ADR **MUST** conform to this contract. Existing endpoints **SHOULD** be migrated within the same sprint as their next scheduled change.

Non-compliant responses detected during code review **MUST** be blocked from merging.

Direct calls to `c.JSON()` that bypass the response helpers are considered non-compliant.

---

*Document maintained by the Domus Engineering Team. Review upon any breaking change to the API contract.*