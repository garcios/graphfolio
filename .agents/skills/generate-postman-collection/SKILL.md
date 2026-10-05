---
name: generate-postman-collection
description: >-
  Guide and step-by-step procedures for generating, updating, and validating Postman Collection v2.1 JSON files for the GraphFolio BFF GraphQL API.
  Use when the user asks to "generate postman collection", "create postman collection", "update postman collection",
  "export postman", "generate-postman", "postman-collection", or maintain GraphQL Postman collections for GraphFolio.
---

# Generating Postman Collections for the GraphFolio BFF GraphQL API

This skill provides comprehensive instructions, architectural guidelines, GraphQL operation catalogs, automated test assertions, and verification procedures for generating and maintaining a complete, valid **Postman Collection v2.1.0** JSON file for the GraphFolio Backend-For-Frontend (BFF) service.

---

## 1. Overview & Reference

- **Target Output File**: [`docs/graphfolio-bff.postman_collection.json`](file:///Users/oscargarcia/workspace/graphfolio/docs/graphfolio-bff.postman_collection.json)
- **Accompanying Documentation**: [`docs/postman-collection.md`](file:///Users/oscargarcia/workspace/graphfolio/docs/postman-collection.md)
- **BFF GraphQL Schema**: [`bff/graph/schema.graphqls`](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.graphqls)
- **BFF Entry Point**: [`bff/cmd/server/main.go`](file:///Users/oscargarcia/workspace/graphfolio/bff/cmd/server/main.go)
- **BFF Resolvers**: [`bff/graph/schema.resolvers.go`](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.resolvers.go)

### Core Goals
1. **100% GraphQL Schema Parity**: Cover all queries (`portfolio`, `instruments`, `portfolioHistory`, `transactions`) and mutations (`addTransaction`, `deleteTransaction`) defined in `bff/graph/schema.graphqls`.
2. **Native GraphQL Editor Integration**: Use Postman's native `"mode": "graphql"` body format with decoupled `query` and `variables` strings so Postman provides full schema autocomplete, syntax highlighting, and variable interpolation.
3. **Exact Arbitrary-Precision Decimal Payloads**: Adhere to the Zero Floating-Point Policy. Quantities, prices, fees, and money amounts use stringified arbitrary-precision decimals (e.g. `"175.50"`) matching the GraphQL `Decimal` scalar.
4. **Environment & Variable Portability**: Use collection variables (`{{baseUrl}}`, `{{timeframe}}`, `{{lastTransactionId}}`) to ensure painless switching between local development, staging, or containerized environments.
5. **Dynamic Workflow Chaining**: Use Postman test scripts (`pm.test` and `pm.collectionVariables.set`) to extract identifiers from mutations (such as newly created transaction IDs) to feed subsequent operations (like `deleteTransaction`).
6. **In-Place Update Semantics**: When updating an existing collection, **preserve** stable collection metadata (`info._postman_id`), existing custom variables, and custom test scripts while merging new schema fields or query variations.

---

## 2. Postman Collection v2.1.0 GraphQL Specification

Every collection generated must strictly comply with the Postman Collection Format v2.1.0 specification with native GraphQL bodies:

```json
{
  "info": {
    "_postman_id": "c6a2e457-4576-4e58-bb12-9c3f0b3ad291",
    "name": "GraphFolio BFF GraphQL API",
    "description": "Comprehensive Postman collection for the GraphFolio Backend-For-Frontend (BFF) GraphQL service.",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "1. Portfolio & Holdings",
      "description": "Consolidated portfolio metrics, net worth, and holding positions.",
      "item": [
        {
          "name": "Get Portfolio Summary",
          "event": [
            {
              "listen": "test",
              "script": {
                "type": "text/javascript",
                "exec": [
                  "pm.test(\"Status code is 200\", function () {",
                  "    pm.response.to.have.status(200);",
                  "});",
                  "pm.test(\"GraphQL response has no errors\", function () {",
                  "    const jsonData = pm.response.json();",
                  "    pm.expect(jsonData.errors).to.be.undefined;",
                  "    pm.expect(jsonData.data.portfolio).to.exist;",
                  "});"
                ]
              }
            }
          ],
          "request": {
            "method": "POST",
            "header": [
              { "key": "Content-Type", "value": "application/json" },
              { "key": "Accept", "value": "application/json" }
            ],
            "body": {
              "mode": "graphql",
              "graphql": {
                "query": "query GetPortfolio {\n  portfolio {\n    totalValue {\n      amount\n      currencyCode\n    }\n    cashBalance {\n      amount\n      currencyCode\n    }\n    investments {\n      id\n      ticker\n      quantity\n    }\n  }\n}",
                "variables": "{}"
              }
            },
            "url": {
              "raw": "{{baseUrl}}/query",
              "host": ["{{baseUrl}}"],
              "path": ["query"]
            },
            "description": "Fetches the full consolidated portfolio state."
          },
          "response": []
        }
      ]
    }
  ],
  "variable": [
    { "key": "baseUrl", "value": "http://localhost:8080", "type": "string" },
    { "key": "timeframe", "value": "TIMEFRAME_1M", "type": "string" },
    { "key": "lastTransactionId", "value": "", "type": "string" }
  ]
}
```

### Key Schema Rules
- `info.schema` MUST be `"https://schema.getpostman.com/json/collection/v2.1.0/collection.json"`.
- `info._postman_id` MUST be a valid UUID v4 string.
- GraphQL POST requests MUST target `{{baseUrl}}/query`.
- GraphQL requests MUST declare `"mode": "graphql"` with `"graphql": { "query": "...", "variables": "..." }`.
- `variables` within the `graphql` object MUST be a JSON-serialized string (e.g. `"{\n  \"timeframe\": \"TIMEFRAME_1M\"\n}"`).

---

## 3. BFF Codebase Inspection Guide

When updating or generating the collection, inspect the following files to ensure parity with the current backend state:

| File | What to Inspect |
|---|---|
| [`bff/graph/schema.graphqls`](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.graphqls) | The source of truth for GraphQL Query, Mutation, Input, Object, and Enum types. |
| [`bff/graph/schema.resolvers.go`](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.resolvers.go) | Resolver implementations, default parameters, and mapping to gRPC requests. |
| [`bff/graph/helpers.go`](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/helpers.go) | Proto-to-GraphQL conversion rules, timeframe mappings, and pagination helpers. |
| [`bff/cmd/server/main.go`](file:///Users/oscargarcia/workspace/graphfolio/bff/cmd/server/main.go) | HTTP route handlers (`/query` for GraphQL, `/` for Playground) and port configuration (`defaultPort = "8080"`). |
| [`proto/portfolio/v1/portfolio.proto`](file:///Users/oscargarcia/workspace/graphfolio/proto/portfolio/v1/portfolio.proto) | gRPC service contract definitions, error codes, and field validation boundaries. |

---

## 4. GraphFolio BFF Operation Catalog

Organize collection items into 6 logical folders matching the system architecture:

### Folder 1: Portfolio & Holdings
Consolidated portfolio metrics, net worth, cash balance, and asset breakdown.

- **POST `{{baseUrl}}/query` - `Get Portfolio Summary`**
  - **Query**:
    ```graphql
    query GetPortfolio {
      portfolio {
        totalValue { amount currencyCode }
        todayReturnAmount { amount currencyCode }
        todayReturnPercent
        annualizedReturnPercent
        cashBalance { amount currencyCode }
        investments {
          id
          ticker
          name
          price { amount currencyCode }
          quantity
          totalValue { amount currencyCode }
          todayReturnAmount { amount currencyCode }
          todayReturnPercent
          totalReturnAmount { amount currencyCode }
          totalReturnPercent
        }
      }
    }
    ```
  - **Variables**: `{}`
  - **Assertions**: Verify 200 OK, `jsonData.errors` undefined, and `portfolio.investments` is an array.

---

### Folder 2: Market Data & Instruments
Tradable instruments, asset classes, and security master definitions.

- **POST `{{baseUrl}}/query` - `List All Instruments`**
  - **Query**:
    ```graphql
    query ListInstruments {
      instruments {
        id
        symbol
        name
        currencyCode
        assetClass
      }
    }
    ```
  - **Variables**: `{}`
  - **Assertions**: Verify 200 OK, returns list of instruments (`AAPL`, `MSFT`, `GOOGL`, `NVDA`, `SPY`, `QQQ`).

---

### Folder 3: Performance History
Time-series valuation curves, Time-Weighted Return (TWR) indexes, and period return calculations across timeframes.

- **POST `{{baseUrl}}/query` - `Get Portfolio History (Variable Timeframe)`**
  - **Variables**: `{"timeframe": "{{timeframe}}"}`
  - **Query**:
    ```graphql
    query GetPortfolioHistory($timeframe: HistoryTimeframe!) {
      portfolioHistory(timeframe: $timeframe) {
        startValue { amount currencyCode }
        endValue { amount currencyCode }
        returnAmount { amount currencyCode }
        returnPercent
        points {
          date
          totalValue { amount currencyCode }
          marketValue { amount currencyCode }
          cashValue { amount currencyCode }
          twrIndex
          dailyReturn
        }
      }
    }
    ```
- **Predefined Timeframe Variants**:
  - `Get Portfolio History (1 Day - 1D)`: `TIMEFRAME_1D`
  - `Get Portfolio History (1 Week - 1W)`: `TIMEFRAME_1W`
  - `Get Portfolio History (1 Month - 1M)`: `TIMEFRAME_1M`
  - `Get Portfolio History (1 Year - 1Y)`: `TIMEFRAME_1Y`
  - `Get Portfolio History (All Time - ALL)`: `TIMEFRAME_ALL`

---

### Folder 4: Transaction Ledger
Paginated transaction queries, multi-criteria filtering, and execution details.

- **POST `{{baseUrl}}/query` - `List Transactions (Default Pagination)`**
  - **Variables**: `{"page": 1, "pageSize": 20}`
  - **Query**:
    ```graphql
    query ListTransactions($type: TransactionType, $symbol: String, $page: Int, $pageSize: Int) {
      transactions(type: $type, symbol: $symbol, page: $page, pageSize: $pageSize) {
        items {
          id
          type
          symbol
          instrumentName
          tradeDate
          quantity
          price { amount currencyCode }
          amount { amount currencyCode }
          fee { amount currencyCode }
          notes
          createdAt
        }
        totalCount
        page
        pageSize
      }
    }
    ```
  - **Dynamic Variable Script**: Extracts `items[0].id` and sets `pm.collectionVariables.set("lastTransactionId", items[0].id)` if transactions exist.
- **POST `{{baseUrl}}/query` - `List Transactions Filtered by Type (BUY)`**
  - **Variables**: `{"type": "BUY", "page": 1, "pageSize": 10}`
- **POST `{{baseUrl}}/query` - `List Transactions Filtered by Symbol (AAPL)`**
  - **Variables**: `{"symbol": "AAPL", "page": 1, "pageSize": 10}`

---

### Folder 5: Transaction Mutations
Recording trades, cash movements, dividend distributions, and deleting ledger records with automatic replay projection rebuilds.

- **POST `{{baseUrl}}/query` - `Add Transaction - BUY Stock`**
  - **Variables**:
    ```json
    {
      "input": {
        "type": "BUY",
        "symbol": "AAPL",
        "tradeDate": "2026-03-15",
        "quantity": "10",
        "price": "175.50",
        "amount": "1755.00",
        "fee": "1.50",
        "currencyCode": "USD",
        "notes": "Postman BUY: 10 shares of AAPL"
      }
    }
    ```
  - **Test Script**: Automatically caches `transactionId` into `{{lastTransactionId}}`:
    ```javascript
    pm.collectionVariables.set("lastTransactionId", jsonData.data.addTransaction.transactionId);
    ```

- **POST `{{baseUrl}}/query` - `Add Transaction - SELL Stock`**
  - **Variables**:
    ```json
    {
      "input": {
        "type": "SELL",
        "symbol": "AAPL",
        "tradeDate": "2026-03-20",
        "quantity": "5",
        "price": "180.00",
        "amount": "900.00",
        "fee": "1.50",
        "currencyCode": "USD",
        "notes": "Postman SELL: 5 shares of AAPL"
      }
    }
    ```

- **POST `{{baseUrl}}/query` - `Add Transaction - DEPOSIT Cash`**
  - **Variables**:
    ```json
    {
      "input": {
        "type": "DEPOSIT",
        "tradeDate": "2026-03-01",
        "amount": "5000.00",
        "currencyCode": "USD",
        "notes": "Postman Cash Deposit"
      }
    }
    ```

- **POST `{{baseUrl}}/query` - `Add Transaction - WITHDRAWAL Cash`**
  - **Variables**:
    ```json
    {
      "input": {
        "type": "WITHDRAWAL",
        "tradeDate": "2026-03-25",
        "amount": "500.00",
        "currencyCode": "USD",
        "notes": "Postman Cash Withdrawal"
      }
    }
    ```

- **POST `{{baseUrl}}/query` - `Add Transaction - DIVIDEND Income`**
  - **Variables**:
    ```json
    {
      "input": {
        "type": "DIVIDEND",
        "symbol": "MSFT",
        "tradeDate": "2026-03-18",
        "amount": "45.00",
        "currencyCode": "USD",
        "notes": "Postman Cash Dividend MSFT"
      }
    }
    ```

- **POST `{{baseUrl}}/query` - `Delete Transaction`**
  - **Variables**:
    ```json
    {
      "id": "{{lastTransactionId}}"
    }
    ```
  - **Query**:
    ```graphql
    mutation DeleteTransaction($id: ID!) {
      deleteTransaction(id: $id) {
        success
        portfolio {
          totalValue { amount currencyCode }
          cashBalance { amount currencyCode }
        }
      }
    }
    ```
  - **Assertions**: Verify `deleteTransaction.success === true`.

---

### Folder 6: Diagnostics & Introspection
GraphQL schema discovery and service availability probes.

- **POST `{{baseUrl}}/query` - `GraphQL Schema Introspection`**
  - Discovers all types, query roots, and mutations using `__schema`.
- **GET `{{baseUrl}}/` - `GraphQL Playground (Web UI Probe)`**
  - Verifies that the interactive GraphQL playground responds with HTTP 200 and loads HTML.

---

## 5. Generation & In-Place Update Workflow

When generating or updating the Postman collection, follow this procedure:

### Step 1: Check If Target File Exists
Inspect [`docs/graphfolio-bff.postman_collection.json`](file:///Users/oscargarcia/workspace/graphfolio/docs/graphfolio-bff.postman_collection.json):
1. **If already present**:
   - Read the existing file.
   - **Preserve `info._postman_id`**: Reusing the existing UUID prevents Postman from creating duplicate collections upon re-import.
   - **Preserve existing variables**: Retain user-customized values in `variable` (e.g. custom `baseUrl` or credentials).
   - Perform an incremental merge of new operations or updated GraphQL fields.
2. **If missing**:
   - Generate a new UUID v4 for `info._postman_id`.
   - Initialize standard variables (`baseUrl`, `timeframe`, `lastTransactionId`).

### Step 2: Validate Against Schema
Verify that all field selections in query strings match [`bff/graph/schema.graphqls`](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.graphqls). Confirm that input objects specify exact decimal values as strings.

### Step 3: Validate JSON Syntax
Always run JSON validation before saving changes:

```bash
python3 -m json.tool docs/graphfolio-bff.postman_collection.json > /dev/null && echo "Valid JSON!"
```

Verify Postman v2.1.0 schema declaration:
```bash
python3 -c "import json; data=json.load(open('docs/graphfolio-bff.postman_collection.json')); assert data['info']['schema'] == 'https://schema.getpostman.com/json/collection/v2.1.0/collection.json'; assert '_postman_id' in data['info']; print('Postman v2.1.0 schema & ID verified!')"
```

---

## 6. Testing & Execution

### Running with Newman (CLI)
You can run automated end-to-end regression tests against a locally running GraphFolio stack:

```bash
# Run collection against local BFF on :8080
npx newman run docs/graphfolio-bff.postman_collection.json \
  --env-var "baseUrl=http://localhost:8080"
```

---

## 7. Maintenance Checklist

Whenever modifying GraphQL schema or BFF resolvers:
- [ ] Updated schema definitions in [`bff/graph/schema.graphqls`](file:///Users/oscargarcia/workspace/graphfolio/bff/graph/schema.graphqls).
- [ ] Ran `make generate` to synchronize Go models and frontend GenQL client.
- [ ] Updated queries, mutations, or variables in [`docs/graphfolio-bff.postman_collection.json`](file:///Users/oscargarcia/workspace/graphfolio/docs/graphfolio-bff.postman_collection.json) in place.
- [ ] Updated documentation in [`docs/postman-collection.md`](file:///Users/oscargarcia/workspace/graphfolio/docs/postman-collection.md) if new operations or parameters were introduced.
- [ ] Confirmed JSON validity with `python3 -m json.tool`.
- [ ] Ran backend unit tests: `make test`.
