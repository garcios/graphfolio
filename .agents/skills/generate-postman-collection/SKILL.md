---
name: generate-postman-collection
description: >-
  Guide and step-by-step procedures for generating, updating, and validating Postman Collection v2.1 JSON files for the Spend Limit REST API.
  Use when the user asks to "generate postman collection", "create postman collection", "update postman collection",
  "export postman", "generate-postman", "postman-collection", or generate API collections based on docs/prompts/prompt-postman.txt.
---

# Generating Postman Collections for the Spend Limit API

This skill provides comprehensive instructions, architectural guidelines, payload catalogs, and verification procedures for generating and maintaining a complete, valid **Postman Collection v2.1.0** JSON file for the Spend Limit REST API.

---

## 1. Overview & Reference

The reference prompt for this workflow is located at [docs/prompts/prompt-postman.txt](file:///Users/oscargarcia/workspace/spend-limit/docs/prompts/prompt-postman.txt).

The target output file is [docs/postman/spend-limit.postman_collection.json](file:///Users/oscargarcia/workspace/spend-limit/docs/postman/spend-limit.postman_collection.json).

### Core Goals
1. **100% Endpoint Coverage**: Reflect all routes defined in [server/server.go](file:///Users/oscargarcia/workspace/spend-limit/server/server.go).
2. **Accurate Payloads & Types**: Populate request bodies with realistic JSON matching Go request DTOs in [server/handlers.go](file:///Users/oscargarcia/workspace/spend-limit/server/handlers.go) and domain rules in [domain/domain.go](file:///Users/oscargarcia/workspace/spend-limit/domain/domain.go).
3. **Environment & Variable Portability**: Use collection variables (`{{baseUrl}}`, `{{customerId}}`, `{{period}}`) and Postman dynamic macros (e.g. `{{$guid}}`) for seamless import and execution.
4. **Valid Postman v2.1.0 Schema**: Ensure strict compliance with the Postman Collection v2.1.0 schema so it imports without syntax or structure errors.
5. **In-Place Update Semantics**: When the output file already exists, **update it in place** rather than regenerating from scratch or creating duplicate files. Preserve stable metadata (`info._postman_id`), existing custom variables, headers, and test scripts while merging new endpoints or updated DTO payloads.

---

## 2. Postman Collection v2.1.0 Specification

Every generated collection must follow the Postman Collection Format v2.1.0:

```json
{
  "info": {
    "_postman_id": "<uuid-v4>",
    "name": "Spend Limit REST API",
    "description": "Comprehensive Postman collection for the Spend Limit and Responsible Gaming service.",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "Folder Name",
      "description": "Folder description",
      "item": [
        {
          "name": "Request Name",
          "request": {
            "method": "POST",
            "header": [
              { "key": "Content-Type", "value": "application/json" },
              { "key": "Accept", "value": "application/json" },
              { "key": "X-Request-ID", "value": "{{$guid}}", "description": "Optional tracing ID" }
            ],
            "body": {
              "mode": "raw",
              "raw": "{\n  \"period\": \"{{period}}\",\n  \"amount\": \"100.00\"\n}",
              "options": {
                "raw": {
                  "language": "json"
                }
              }
            },
            "url": {
              "raw": "{{baseUrl}}/api/v1/customers/:id/limits",
              "host": ["{{baseUrl}}"],
              "path": ["api", "v1", "customers", ":id", "limits"],
              "variable": [
                {
                  "key": "id",
                  "value": "{{customerId}}",
                  "description": "Customer UUID"
                }
              ]
            },
            "description": "Detailed description of the endpoint."
          },
          "response": []
        }
      ]
    }
  ],
  "variable": [
    { "key": "baseUrl", "value": "http://localhost:8080", "type": "string" },
    { "key": "customerId", "value": "123e4567-e89b-12d3-a456-426614174000", "type": "string" },
    { "key": "period", "value": "HOURS_24", "type": "string" }
  ]
}
```

### Key Schema Requirements
- `info.schema` MUST be `"https://schema.getpostman.com/json/collection/v2.1.0/collection.json"`.
- `info._postman_id` MUST be a valid UUID string (e.g. `8a32d1ef-4576-4e58-bb12-9c3f0b3ad291`).
- Path variables in `url.raw` use `:variableName` syntax and MUST be declared in `url.variable` array.
- Query parameters in `url.raw` (e.g. `?client_id={{customerId}}`) MUST be declared in `url.query` array with `key`, `value`, and `description`.
- Request bodies with JSON payloads MUST specify `"mode": "raw"` and `"options": { "raw": { "language": "json" } }`.

---

## 3. Go Codebase Inspection Guide

When updating or generating the collection, inspect the following files to ensure full parity with code:

| File | What to Inspect |
|---|---|
| [server/server.go](file:///Users/oscargarcia/workspace/spend-limit/server/server.go) | All route registrations on `mux.HandleFunc("<METHOD> <path>", ...)` |
| [server/handlers.go](file:///Users/oscargarcia/workspace/spend-limit/server/handlers.go) | Handler implementations, request struct definitions (`readJSON`), path variables (`r.PathValue`), query parameters (`r.URL.Query().Get`), and status codes |
| [server/response.go](file:///Users/oscargarcia/workspace/spend-limit/server/response.go) | Response envelopes (`ErrorResponse`, `ErrorDetail`, error codes) |
| [domain/domain.go](file:///Users/oscargarcia/workspace/spend-limit/domain/domain.go) | Domain models, valid periods (`HOURS_24`, `DAYS_7`, `DAYS_30`), sources (`CUSTOMER`, `INTERNAL`, `ADMIN`), and transaction types |
| [config/config.go](file:///Users/oscargarcia/workspace/spend-limit/config/config.go) | Default host and port configuration (`SERVER_HOST`, `SERVER_PORT`) |

---

## 4. Spend Limit API Endpoint & Payload Catalog

Organize collection items into 5 logical folders matching the API's architecture:

### Folder 1: Health
Service health probes and diagnostics.

- **GET `/health`**
  - **Name**: Health Check & Database Probe
  - **Headers**: `Accept: application/json`
  - **URL**: `{{baseUrl}}/health`
  - **Description**: Verifies service status and runs MySQL ping check.

---

### Folder 2: Spend Limits
Customer limit lifecycle, creation, immediate decreases, cooling-off increases, removals, and pending queries.

- **POST `/api/v1/customers/:id/limits`**
  - **Name**: Create Spend Limit
  - **Headers**: `Content-Type: application/json`, `Accept: application/json`, `X-Request-ID: {{$guid}}`
  - **URL**: `{{baseUrl}}/api/v1/customers/:id/limits` (`:id` = `{{customerId}}`)
  - **Body (`raw/json`)**:
    ```json
    {
      "period": "{{period}}",
      "amount": "100.00",
      "source": "CUSTOMER",
      "requested_by": "customer"
    }
    ```
  - **Status**: 201 Created (or 409 Conflict if limit already exists).

- **PUT `/api/v1/customers/:id/limits/:period`**
  - **Name**: Update Spend Limit
  - **Headers**: `Content-Type: application/json`, `Accept: application/json`, `X-Request-ID: {{$guid}}`
  - **URL**: `{{baseUrl}}/api/v1/customers/:id/limits/:period` (`:id` = `{{customerId}}`, `:period` = `{{period}}`)
  - **Body (`raw/json`)**:
    ```json
    {
      "amount": "50.00",
      "requested_by": "customer"
    }
    ```
  - **Notes**: Decreases take effect immediately (200 OK). Increases queue with a mandatory 24h cooling-off period (202 Accepted).

- **DELETE `/api/v1/customers/:id/limits/:period`**
  - **Name**: Request Delete Spend Limit
  - **Headers**: `Accept: application/json`, `X-Request-ID: {{$guid}}`
  - **URL**: `{{baseUrl}}/api/v1/customers/:id/limits/:period?requested_by=customer`
  - **Query Params**: `requested_by` (optional, defaults to customer ID)
  - **Status**: 202 Accepted (enters 24h cooling-off period).

- **POST `/api/v1/customers/:id/limits/:period/cancel`**
  - **Name**: Cancel Pending Limit Request
  - **Headers**: `Accept: application/json`, `X-Request-ID: {{$guid}}`
  - **URL**: `{{baseUrl}}/api/v1/customers/:id/limits/:period/cancel`
  - **Status**: 200 OK (cancels pending increase or removal before cooling-off expires).

- **GET `/api/v1/customers/:id/limits`**
  - **Name**: List Customer Spend Limits
  - **Headers**: `Accept: application/json`
  - **URL**: `{{baseUrl}}/api/v1/customers/:id/limits`
  - **Status**: 200 OK (returns array of active limits across all periods).

- **GET `/api/v1/customers/:id/limits/current`**
  - **Name**: Get Current Restricting Spend Limit
  - **Headers**: `Accept: application/json`
  - **URL**: `{{baseUrl}}/api/v1/customers/:id/limits/current`
  - **Status**: 200 OK (returns the most restrictive active limit detail).

- **GET `/api/v1/customers/:id/requests/pending`**
  - **Name**: List Customer Pending Requests
  - **Headers**: `Accept: application/json`
  - **URL**: `{{baseUrl}}/api/v1/customers/:id/requests/pending`
  - **Status**: 200 OK.

- **GET `/api/v1/requests/pending`**
  - **Name**: List All Pending Requests (Admin/Global)
  - **Headers**: `Accept: application/json`
  - **URL**: `{{baseUrl}}/api/v1/requests/pending?client_id={{customerId}}`
  - **Query Params**: `client_id` (optional filter).

---

### Folder 3: Transactions
Real-money betting, generosity/bonus bets, and refunds with limit validation.

- **POST `/api/v1/customers/:id/transactions/bet`**
  - **Name**: Place Cash Bet
  - **Headers**: `Content-Type: application/json`, `Accept: application/json`, `X-Request-ID: {{$guid}}`
  - **URL**: `{{baseUrl}}/api/v1/customers/:id/transactions/bet`
  - **Body (`raw/json`)**:
    ```json
    {
      "amount": "25.00",
      "txn_id": "txn-{{$guid}}"
    }
    ```
  - **Status**: 201 Created on success, 422 Unprocessable Entity (`LIMIT_REACHED`) if limit is breached.

- **POST `/api/v1/customers/:id/transactions/generosity`**
  - **Name**: Place Bonus / Generosity Bet
  - **Headers**: `Content-Type: application/json`, `Accept: application/json`, `X-Request-ID: {{$guid}}`
  - **URL**: `{{baseUrl}}/api/v1/customers/:id/transactions/generosity`
  - **Body (`raw/json`)**:
    ```json
    {
      "amount": "10.00",
      "txn_id": "txn-bonus-{{$guid}}"
    }
    ```
  - **Notes**: Generosity bets do not consume spend limit budget, but are blocked if customer has reached $0 remaining balance.

- **POST `/api/v1/customers/:id/transactions/refund`**
  - **Name**: Refund Cash Bet
  - **Headers**: `Content-Type: application/json`, `Accept: application/json`, `X-Request-ID: {{$guid}}`
  - **URL**: `{{baseUrl}}/api/v1/customers/:id/transactions/refund`
  - **Body (`raw/json`)**:
    ```json
    {
      "amount": "25.00",
      "txn_id": "ref-{{$guid}}"
    }
    ```
  - **Status**: 200 OK (restores spent budget, floored at $0).

---

### Folder 4: Admin & Operations
Batch processing, testing maintenance, counter resets, and Transactional Outbox inspection.

- **POST `/api/v1/admin/apply-pending-changes`**
  - **Name**: Apply Expired Pending Changes
  - **Headers**: `Content-Type: application/json`, `Accept: application/json`
  - **URL**: `{{baseUrl}}/api/v1/admin/apply-pending-changes`
  - **Body (`raw/json`)**:
    ```json
    {
      "client_id": "{{customerId}}"
    }
    ```
  - **Notes**: Promotes expired cooling-off increases and deletions. `client_id` is optional.

- **POST `/api/v1/admin/reset-counter`**
  - **Name**: Reset Spend Counter
  - **Headers**: `Content-Type: application/json`, `Accept: application/json`
  - **URL**: `{{baseUrl}}/api/v1/admin/reset-counter`
  - **Body (`raw/json`)**:
    ```json
    {
      "client_id": "{{customerId}}",
      "period": "{{period}}"
    }
    ```

- **POST `/api/v1/admin/time-travel`**
  - **Name**: Time Travel (Testing Utility)
  - **Headers**: `Content-Type: application/json`, `Accept: application/json`
  - **URL**: `{{baseUrl}}/api/v1/admin/time-travel`
  - **Body (`raw/json`)**:
    ```json
    {
      "client_id": "{{customerId}}",
      "period": "{{period}}",
      "field": "effectiveAt",
      "at": "2026-09-01T00:00:00Z"
    }
    ```
  - **Field Values**: `"effectiveAt"`, `"windowEnd"`, `"windowStart"`.

- **GET `/api/v1/admin/events`**
  - **Name**: List Outbox Domain Events
  - **Headers**: `Accept: application/json`
  - **URL**: `{{baseUrl}}/api/v1/admin/events?client_id={{customerId}}`
  - **Query Params**: `client_id` (optional filter).

---

### Folder 5: Observability & Metrics
LRU memory cache and telemetry monitoring.

- **GET `/api/v1/admin/cache/metrics`**
  - **Name**: Get Memory Cache Metrics
  - **Headers**: `Accept: application/json`
  - **URL**: `{{baseUrl}}/api/v1/admin/cache/metrics`
  - **Alternative Aliases**: `/api/v1/cache/metrics`, `/api/v1/metrics/cache`.
  - **Response Payload**: `{"hits": 120, "misses": 5, "evictions": 0, "len": 4, "capacity": 10000, "hit_rate": 0.96, "active_clients": 2}`.

---

## 5. Step-by-Step Generation & In-Place Update Workflow

Follow this procedure when creating or updating the Postman collection:

### Step 1: Check If Target File Exists & Load Existing Collection
Before generating or updating, inspect the target path [docs/postman/spend-limit.postman_collection.json](file:///Users/oscargarcia/workspace/spend-limit/docs/postman/spend-limit.postman_collection.json):

1. **If the file ALREADY EXISTS**:
   - Read and parse the existing JSON file into memory.
   - **Preserve `info._postman_id`**: Reusing the existing UUID ensures Postman updates the collection on re-import rather than creating a duplicate collection in the user's workspace.
   - **Preserve existing customizations**: Retain any existing collection variables, descriptions, pre-request scripts, or test scripts (`item[].event`).
   - Extract the existing folder hierarchy and endpoints to perform an incremental merge.

2. **If the file DOES NOT EXIST**:
   - Initialize a new collection structure using a freshly generated UUID v4 for `info._postman_id`.
   - Set up the standard collection variables (`baseUrl`, `customerId`, `period`).

### Step 2: Scan Go Routes and Handlers
Review [server/server.go](file:///Users/oscargarcia/workspace/spend-limit/server/server.go) and [server/handlers.go](file:///Users/oscargarcia/workspace/spend-limit/server/handlers.go) for:
1. Any newly added endpoints.
2. Changes in path patterns or query parameters.
3. Updated DTO fields, validation tags, or default values.

### Step 3: Merge & Update Endpoints and Folders
Perform an in-place merge of discovered API endpoints against the collection:
- **Match existing requests**: Match endpoints by HTTP method and relative path.
  - If a route already exists: Update its URL parameters, request body payload (matching updated Go structs), or headers if they changed. Preserve any user-added test scripts or notes.
  - If a route is new: Append it under the appropriate logical folder (`Health`, `Spend Limits`, `Transactions`, `Admin & Operations`, `Observability & Metrics`).
- **Ensure Folder Structure**: If a required folder is missing, create it in logical order.
- **Synchronize Variables**: Ensure default variables (`baseUrl`, `customerId`, `period`) exist in the `variable` array without overwriting existing custom variables.

### Step 4: Validate Updated JSON Output
Always validate the updated JSON file using Python or `jq` before committing:

```bash
# Validate JSON syntax
python3 -c "import json; json.load(open('docs/postman/spend-limit.postman_collection.json')); print('JSON syntax is valid!')"

# Verify top-level Postman schema key and preserved _postman_id
python3 -c "import json; data=json.load(open('docs/postman/spend-limit.postman_collection.json')); assert data['info']['schema'] == 'https://schema.getpostman.com/json/collection/v2.1.0/collection.json'; assert '_postman_id' in data['info']; print('Postman v2.1 schema & ID confirmed!')"
```

### Step 5: Save Updated Collection File In-Place
Write the updated collection directly back to:
- `docs/postman/spend-limit.postman_collection.json`

> [!IMPORTANT]
> Always update the existing file in place (`docs/postman/spend-limit.postman_collection.json`). Do not create new alternate files (like `spend-limit-v2.json` or timestamped files) unless explicitly instructed.

---

## 6. Testing & Execution

### Running with Newman (CLI)
You can execute automated regression runs against a local server using Newman:

```bash
# 1. Start the Spend Limit REST API server
./spendlimit serve &
SERVER_PID=$!

# 2. Run the collection using newman
npx -y newman run docs/postman/spend-limit.postman_collection.json \
  --env-var "baseUrl=http://localhost:8080" \
  --env-var "customerId=123e4567-e89b-12d3-a456-426614174000"

# 3. Stop the test server
kill $SERVER_PID
```

---

## 7. Maintenance Checklist

Whenever modifying or adding API endpoints in Go:
- [ ] Added endpoint to [server/server.go](file:///Users/oscargarcia/workspace/spend-limit/server/server.go) routes.
- [ ] Created handler and request/response DTOs in [server/handlers.go](file:///Users/oscargarcia/workspace/spend-limit/server/handlers.go).
- [ ] Updated [docs/postman/spend-limit.postman_collection.json](file:///Users/oscargarcia/workspace/spend-limit/docs/postman/spend-limit.postman_collection.json) in place (preserving `info._postman_id` and existing variables/scripts).
- [ ] Provided realistic JSON payload matching DTO struct.
- [ ] Validated JSON formatting with `python3 -m json.tool`.
