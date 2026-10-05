# GraphFolio BFF Postman Collection

This document describes the **Postman Collection v2.1.0** for the GraphFolio Backend-For-Frontend (BFF) service.

The collection file is located at:
- [`docs/graphfolio-bff.postman_collection.json`](file:///Users/oscargarcia/workspace/graphfolio/docs/graphfolio-bff.postman_collection.json)

---

## 1. Overview

The Backend-For-Frontend (BFF) serves as the unified GraphQL API gateway between client applications (Web React frontend, scripts, external consumers) and the underlying gRPC microservices (`portfolio-api`, `user-api`).

- **Base URL**: `http://localhost:8080`
- **GraphQL Endpoint**: `http://localhost:8080/query`
- **Interactive Playground**: `http://localhost:8080/`
- **Serialization**: Arbitrary-precision decimal strings (`"175.50"`) mapped to the custom `Decimal` scalar and `Money` types (`amount`, `currencyCode`).

---

## 2. Importing the Collection

### In the Postman App
1. Open **Postman**.
2. Click **Import** (top left).
3. Drag & drop or select [`docs/graphfolio-bff.postman_collection.json`](file:///Users/oscargarcia/workspace/graphfolio/docs/graphfolio-bff.postman_collection.json).
4. The collection `GraphFolio BFF GraphQL API` will appear in your workspace with full GraphQL query editor support, pre-configured variables, and tests.

---

## 3. Collection Variables

The collection defines the following built-in collection variables:

| Variable | Default Value | Description |
|---|---|---|
| `baseUrl` | `http://localhost:8080` | Host and port where the BFF is running |
| `timeframe` | `TIMEFRAME_1M` | History timeframe (`TIMEFRAME_1D`, `TIMEFRAME_1W`, `TIMEFRAME_1M`, `TIMEFRAME_1Y`, `TIMEFRAME_ALL`) |
| `lastTransactionId` | `""` | Dynamically set by transaction creation and list queries; consumed by `Delete Transaction` |

---

## 4. Folder Structure & Operations

The collection organizes 19 requests across 6 logical folders:

### 1. Portfolio & Holdings
- **Get Portfolio Summary** (`POST {{baseUrl}}/query`)
  - Retrieves consolidated net worth (`totalValue`), cash balance (`cashBalance`), today's return amount and percentage, annualized returns, and individual asset holdings breakdown with current prices, quantities, and return metrics.

### 2. Market Data & Instruments
- **List All Instruments** (`POST {{baseUrl}}/query`)
  - Retrieves the security master catalog of tradable instruments (e.g. `AAPL`, `MSFT`, `GOOGL`, `NVDA`, `SPY`, `QQQ`) with symbols, names, currencies, and asset classes (`EQUITY`, `ETF`).

### 3. Performance History
- **Get Portfolio History (Variable Timeframe)** (`POST {{baseUrl}}/query` using `{{timeframe}}`)
- **Get Portfolio History (1 Day - 1D)**
- **Get Portfolio History (1 Week - 1W)**
- **Get Portfolio History (1 Month - 1M)**
- **Get Portfolio History (1 Year - 1Y)**
- **Get Portfolio History (All Time - ALL)**
  - Retrieves time-series valuation points (`date`, `totalValue`, `marketValue`, `cashValue`, `twrIndex`, `dailyReturn`) and period summary returns (`startValue`, `endValue`, `returnAmount`, `returnPercent`).

### 4. Transaction Ledger
- **List Transactions (Default Pagination)** (`POST {{baseUrl}}/query`)
  - Paginated transaction records (`page`, `pageSize`, `totalCount`). Automatically extracts `items[0].id` and updates `{{lastTransactionId}}`.
- **List Transactions Filtered by Type (BUY)**
  - Filters ledger items by transaction type (`BUY`, `SELL`, `DIVIDEND`, `DEPOSIT`, `WITHDRAWAL`).
- **List Transactions Filtered by Symbol (AAPL)**
  - Filters ledger items by instrument ticker symbol.

### 5. Transaction Mutations
- **Add Transaction - BUY Stock**
  - Ingests a `BUY` trade order. Automatically saves the returned `transactionId` to `{{lastTransactionId}}`.
- **Add Transaction - SELL Stock**
  - Ingests a `SELL` trade order with automated tax lot relief and cash credit.
- **Add Transaction - DEPOSIT Cash**
  - Ingests a cash deposit into portfolio cash balance.
- **Add Transaction - WITHDRAWAL Cash**
  - Ingests a cash withdrawal.
- **Add Transaction - DIVIDEND Income**
  - Ingests cash dividend income attributed to a stock holding.
- **Delete Transaction**
  - Deletes a transaction by ID (using `{{lastTransactionId}}`) and triggers an automatic replay projection rebuild (`RebuildProjections`) to restore portfolio state and cash balance.

### 6. Diagnostics & Introspection
- **GraphQL Schema Introspection** (`POST {{baseUrl}}/query`)
  - Executes a standard GraphQL `__schema` introspection query.
- **GraphQL Playground (Web UI Probe)** (`GET {{baseUrl}}/`)
  - Verifies HTTP 200 response and HTML rendering for the developer playground UI.

---

## 5. Automated Tests & Variable Chaining

Every request in the collection includes pre-configured JavaScript assertions (`pm.test`):
- Verifies HTTP 200 OK status.
- Confirms the absence of `errors` in the GraphQL envelope (`jsonData.errors === undefined`).
- Verifies structural integrity of the returned payload.
- Dynamically extracts IDs from mutations or list queries to enable automated end-to-end testing flows (e.g. **Add BUY Transaction** ➔ saves `lastTransactionId` ➔ **Delete Transaction** ➔ deletes using saved ID).

---

## 6. Running with Newman

You can execute automated regression tests against the running BFF server via [Newman](https://learning.postman.com/docs/collections/using-newman-cli/command-line-integration-with-newman/):

```bash
npx newman run docs/graphfolio-bff.postman_collection.json \
  --env-var "baseUrl=http://localhost:8080"
```
