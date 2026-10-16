# Cloudflare Containers Deployment Plan for GraphFolio

> **Target**: Comprehensive architectural specification and deployment plan for hosting GraphFolio on the Cloudflare edge compute platform utilizing **Cloudflare Containers**, **Cloudflare Workers**, **Cloudflare Pages / Workers Static Assets**, and **Cloudflare Hyperdrive**.  
> **Scope**: Monorepo Go microservices (`services/portfolio-api`, `services/user-api`, `bff`), Valuation background worker (`cmd/worker`), React 19 SPAs (`web/apps/main-app`, `web/apps/admin-app`), and managed PostgreSQL database connectivity.  
> **Key Principle**: Maximize global edge performance, achieve sub-millisecond static asset delivery, preserve zero floating-point arithmetic integrity, and maintain strict gRPC isolation while optimizing container lifecycle, cold starts, and cost efficiency.

---

## 1. Executive Summary & Target Topology

GraphFolio is a multi-tier financial portfolio management platform comprising:
- **Two React 19 SPAs**: `main-app` (investor portal) and `admin-app` (operational directory and price ledger).
- **Go GraphQL BFF**: Orchestrator terminating client GraphQL queries and managing JWT/cookie sessions.
- **Two Go gRPC Microservices**: `portfolio-api` (financial engine, 3-pillar return attribution, tax lots) and `user-api` (Argon2id authentication, preferences).
- **Background Valuation Engine**: Daily EOD market valuation and historical backfill worker.
- **Relational Storage**: PostgreSQL 15+ database with dual schemas (`portfolio`, `user`) and `shopspring/decimal` support.

### Cloudflare Architecture Diagram

```mermaid
flowchart TD
    subgraph Clients["Global Clients"]
        Investor["Investor Browser<br>(graphfolio.app)"]
        Admin["Operations Browser<br>(admin.graphfolio.app)"]
    end

    subgraph Edge["Cloudflare Global Network"]
        WAF["Cloudflare WAF / Zero Trust Access"]
        CFPages["Cloudflare Pages / Static Assets<br>(main-app & admin-app)"]
        WorkerGateway["Cloudflare Worker API Gateway<br>(Router & CORS)"]
        CronTrigger["Cloudflare Cron Trigger<br>(Daily 22:00 UTC)"]
        Hyperdrive["Cloudflare Hyperdrive<br>(Pool & Acceleration)"]
    end

    subgraph Containers["Cloudflare Containers Runtime"]
        DO["GraphFolio Durable Object"]
        Supervisor["Process Supervisor"]
        BFF["GraphQL BFF (:8080)"]
        PortSvc["Portfolio API (:50051 gRPC)"]
        UserSvc["User API (:50052 gRPC)"]
    end

    subgraph External["External Infrastructure"]
        Postgres[("Managed PostgreSQL 15+<br>Schemas: portfolio, user")]
        TwelveData["Twelve Data Market API"]
    end

    Investor -->|"HTTPS / Assets"| CFPages
    Admin -->|"HTTPS / Zero Trust"| CFPages
    Investor -->|"POST /graphql"| WAF
    Admin -->|"POST /graphql"| WAF
    WAF --> WorkerGateway
    CronTrigger -->|"Scheduled Cron"| WorkerGateway
    WorkerGateway -->|"Container DO Fetch"| DO
    DO --> BFF
    Supervisor --> BFF
    Supervisor --> PortSvc
    Supervisor --> UserSvc
    BFF -->|"gRPC 127.0.0.1:50051"| PortSvc
    BFF -->|"gRPC 127.0.0.1:50052"| UserSvc
    PortSvc --> Hyperdrive
    UserSvc --> Hyperdrive
    PortSvc -->|"Market Sync"| TwelveData
    Hyperdrive -->|"TLS Connection Pool"| Postgres
```

---

## 2. Architectural Decisions & Strategy Trade-offs

### Decision 1: Deployment Topology — Unified Backend Container vs Split Distributed Containers

| Strategy | Pros | Cons | Recommendation |
| :--- | :--- | :--- | :--- |
| **Option A: Distributed Containers**<br/>(3 separate Cloudflare Containers: `bff`, `portfolio-api`, `user-api`) | Strict isolation; independent scaling of compute. | Complex cross-container gRPC over Cloudflare DO bridges; multiple cold starts per user request; tripled container billing minimums; gRPC multiplexing overhead. | Not recommended for initial phase. |
| **Option B: Unified Container Pod (Recommended)**<br/>(Single OCI container running `bff`, `portfolio-api`, and `user-api` managed by lightweight supervisor) | Zero-latency inter-service gRPC over loopback (`127.0.0.1`); single atomic cold start; shared memory efficiency (~60MB total footprint in Go); streamlined deployment and environment parity. | Shared container lifecycle (all services restart together). | **Selected (Option B)**. |

**Rationale**: In GraphFolio, `bff` strictly calls `portfolio-api` and `user-api` via gRPC. Colocating these compiled Go binaries inside a single lightweight Alpine container yields sub-millisecond RPC latency (`<0.2ms` on loopback), eliminates complex inter-container TLS handshakes, and keeps the entire backend memory under 128MB.

### Decision 2: Frontend Hosting — Cloudflare Pages vs Containerized Nginx

- **Decision**: Host `web/apps/main-app` and `web/apps/admin-app` as static Single-Page Applications (SPAs) on **Cloudflare Pages** (or Cloudflare Workers Static Assets).
- **Rationale**:
  1. Instant global edge distribution across 330+ Cloudflare data centers with zero container cold start.
  2. Free/low-cost asset bandwidth without consuming container CPU or memory time slices.
  3. Continuous preview builds per Git branch.

### Decision 3: Database & State Persistence Strategy

- **Decision**: Managed PostgreSQL (e.g., Neon Serverless Postgres, Supabase, or AWS RDS) fronted by **Cloudflare Hyperdrive**.
- **Rationale**:
  1. Cloudflare Containers are ephemeral and scale down to zero when idle. They cannot host stateful persistent disk volumes for relational databases.
  2. Cloudflare D1 is SQLite-based and does not support GraphFolio's multi-schema architecture (`portfolio` and `user`), `shopspring/decimal` fixed-point registration, PostgreSQL advisory locks, or complex CTEs (`combined_holdings`).
  3. **Cloudflare Hyperdrive** accelerates connection establishment, reuses connection pools across global edge nodes, and avoids PostgreSQL connection starvation during container scale-outs.

### Decision 4: Daily Valuation Worker Execution

- **Decision**: Expose an authenticated internal HTTP webhook on the BFF (`POST /internal/jobs/valuation-eod`), triggered by a **Cloudflare Cron Trigger** (`0 22 * * *`).
- **Rationale**: Cloudflare Workers feature native Cron Triggers. When the cron fires, the Worker forwards an authenticated request to the container, which runs the valuation engine without requiring a 24/7 dedicated compute server.

---

## 3. Container Specification & Multi-Stage Dockerfile

### Process Management Architecture

Inside the unified backend container:
1. `entrypoint.sh` initializes environment variables and validates database connectivity.
2. If `MIGRATE_ON_START=true`, database migrations run once sequentially.
3. Microservices start in order:
   - `user-api` starts on `127.0.0.1:50052`.
   - `portfolio-api` starts on `127.0.0.1:50051`.
   - Healthcheck ensures both gRPC endpoints respond.
   - `bff` starts on `0.0.0.0:8080` (exposed to Cloudflare Worker).

### Multi-Stage Container Dockerfile (`deploy/docker/Dockerfile`)

```dockerfile
# ==============================================================================
# Stage 1: Build Go Binaries
# ==============================================================================
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /workspace

# Copy workspace and module definitions for dependency caching
COPY go.work go.work.sum ./
COPY pkg/go.mod pkg/go.sum* ./pkg/
COPY proto/go.mod proto/go.sum* ./proto/
COPY bff/go.mod bff/go.sum* ./bff/
COPY services/portfolio-api/go.mod services/portfolio-api/go.sum* ./services/portfolio-api/
COPY services/user-api/go.mod services/user-api/go.sum* ./services/user-api/

RUN go work sync

# Copy source code
COPY pkg/ ./pkg/
COPY proto/ ./proto/
COPY bff/ ./bff/
COPY services/portfolio-api/ ./services/portfolio-api/
COPY services/user-api/ ./services/user-api/

# Build static binaries with optimizations (stripped symbols)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /bin/bff ./bff/cmd/server
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /bin/portfolio-api ./services/portfolio-api/cmd/server
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /bin/portfolio-worker ./services/portfolio-api/cmd/worker
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /bin/user-api ./services/user-api/cmd/server

# ==============================================================================
# Stage 2: Minimal Runtime Environment
# ==============================================================================
FROM alpine:3.21 AS runtime

RUN apk add --no-cache ca-certificates tzdata curl bash

# Create non-privileged service user
RUN addgroup -S graphfolio && adduser -S -G graphfolio -u 1001 graphfolio

WORKDIR /app

# Copy compiled binaries from builder stage
COPY --from=builder /bin/bff /app/bff
COPY --from=builder /bin/portfolio-api /app/portfolio-api
COPY --from=builder /bin/portfolio-worker /app/portfolio-worker
COPY --from=builder /bin/user-api /app/user-api

# Copy migration files for dynamic startup execution
COPY --from=builder /workspace/services/portfolio-api/migrations /app/migrations/portfolio
COPY --from=builder /workspace/services/user-api/migrations /app/migrations/user

# Copy container supervisor entrypoint
COPY deploy/docker/entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh /app/bff /app/portfolio-api /app/portfolio-worker /app/user-api

USER graphfolio:graphfolio

EXPOSE 8080

ENTRYPOINT ["/app/entrypoint.sh"]
```

### Supervisor Entrypoint (`deploy/docker/entrypoint.sh`)

```bash
#!/usr/bin/env bash
set -eo pipefail

echo "[container-init] Starting GraphFolio Unified Backend..."

# Trap termination signals for graceful shutdown
cleanup() {
    echo "[container-init] Received termination signal. Terminating child processes..."
    kill -TERM "$BFF_PID" "$PORTFOLIO_PID" "$USER_PID" 2>/dev/null || true
    wait "$BFF_PID" "$PORTFOLIO_PID" "$USER_PID" 2>/dev/null || true
    echo "[container-init] All services terminated gracefully."
    exit 0
}
trap cleanup SIGINT SIGTERM

# Configure internal service URLs
export PORTFOLIO_API_URL="127.0.0.1:50051"
export USER_API_URL="127.0.0.1:50052"
export PORT="8080"

# 1. Start User API (gRPC :50052)
echo "[container-init] Launching User API on :50052..."
PORT=50052 /app/user-api &
USER_PID=$!

# 2. Start Portfolio API (gRPC :50051)
echo "[container-init] Launching Portfolio API on :50051..."
PORT=50051 /app/portfolio-api &
PORTFOLIO_PID=$!

# 3. Quick loopback health probe
echo "[container-init] Waiting for gRPC listeners to accept connections..."
for i in {1..30}; do
    if nc -z 127.0.0.1 50051 2>/dev/null && nc -z 127.0.0.1 50052 2>/dev/null; then
        echo "[container-init] Both internal gRPC microservices are healthy."
        break
    fi
    sleep 0.1
done

# 4. Start GraphQL BFF (HTTP :8080)
echo "[container-init] Launching GraphQL BFF on :8080..."
/app/bff &
BFF_PID=$!

# Wait for BFF or any process failure
wait -n "$BFF_PID" "$PORTFOLIO_PID" "$USER_PID"
echo "[container-init] One of the backend processes exited. Initiating container shutdown..."
cleanup
```

---

## 4. Cloudflare Worker Gateway & Container Configuration

Cloudflare Containers are bound to Workers via **Durable Objects**. The Worker handles TLS termination, request sanitization, CORS negotiation, and cron schedules, and delegates payload streaming directly to the container via `getTcpPort(8080).fetch()`.

### Wrangler Configuration (`wrangler.jsonc`)

```jsonc
{
  "$schema": "node_modules/wrangler/config-schema.json",
  "name": "graphfolio-backend",
  "main": "src/index.ts",
  "compatibility_date": "2026-10-15",
  "compatibility_flags": ["nodejs_compat"],

  // 1. Cloudflare Container Configuration
  "containers": [
    {
      "class_name": "GraphFolioContainer",
      "image": "./deploy/docker" // or registry URL: "registry.cloudflare.com/my-account/graphfolio-backend:latest"
    }
  ],

  // 2. Durable Objects Binding to Container Class
  "durable_objects": {
    "bindings": [
      {
        "name": "GRAPHFOLIO_CONTAINER",
        "class_name": "GraphFolioContainer"
      }
    ]
  },

  // 3. Durable Object Migration Tag
  "migrations": [
    {
      "tag": "v1",
      "new_classes": ["GraphFolioContainer"]
    }
  ],

  // 4. Cloudflare Hyperdrive Database Acceleration
  "hyperdrive": [
    {
      "binding": "PORTFOLIO_HYPERDRIVE",
      "id": "<portfolio-hyperdrive-id>"
    },
    {
      "binding": "USER_HYPERDRIVE",
      "id": "<user-hyperdrive-id>"
    }
  ],

  // 5. Cron Triggers for Scheduled EOD Valuation
  "triggers": {
    "crons": ["0 22 * * 1-5"] // Mon-Fri at 22:00 UTC (Market Close)
  },

  // 6. Custom Domain Route for API Gateway & Container Pod
  "routes": [
    {
      "pattern": "api.graphfolio.app/*",
      "custom_domain": true
    }
  ],

  // 7. Environment Variables
  "vars": {
    "ENVIRONMENT": "production",
    "CORS_ORIGINS": "https://graphfolio.app,https://app.graphfolio.app,https://admin.graphfolio.app",
    "MIGRATE_ON_START": "false"
  }
}
```

### Worker Entrypoint & Durable Object Handler (`src/index.ts`)

```typescript
import { DurableObject } from "cloudflare:workers";

interface Env {
  GRAPHFOLIO_CONTAINER: DurableObjectNamespace<GraphFolioContainer>;
  PORTFOLIO_HYPERDRIVE: Hyperdrive;
  USER_HYPERDRIVE: Hyperdrive;
  CORS_ORIGINS: string;
  INTERNAL_CRON_SECRET: string;
}

// -----------------------------------------------------------------------------
// Durable Object housing the Cloudflare Container instance
// -----------------------------------------------------------------------------
export class GraphFolioContainer extends DurableObject {
  async fetch(request: Request): Promise<Response> {
    const url = new URL(request.url);

    // Forward the request to the container's BFF HTTP port (:8080)
    url.protocol = "http:";
    url.host = "127.0.0.1:8080";

    const forwardedRequest = new Request(url.toString(), {
      method: request.method,
      headers: request.headers,
      body: request.body,
      redirect: "manual",
    });

    // Strip incoming host header to avoid host matching mismatches
    forwardedRequest.headers.delete("host");

    try {
      const tcpPort = this.ctx.container.getTcpPort(8080);
      return await tcpPort.fetch(forwardedRequest);
    } catch (err: any) {
      return new Response(JSON.stringify({ error: "Container upstream unavailable", details: err?.message }), {
        status: 503,
        headers: { "Content-Type": "application/json" },
      });
    }
  }
}

// -----------------------------------------------------------------------------
// Worker Gateway Router
// -----------------------------------------------------------------------------
export default {
  async fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
    const url = new URL(request.url);
    const origin = request.headers.get("Origin") || "";
    const allowedOrigins = env.CORS_ORIGINS.split(",");
    const isAllowedOrigin = allowedOrigins.includes(origin);

    // 1. Handle CORS Preflight Requests at Edge
    if (request.method === "OPTIONS") {
      if (isAllowedOrigin) {
        return new Response(null, {
          status: 204,
          headers: {
            "Access-Control-Allow-Origin": origin,
            "Access-Control-Allow-Methods": "GET, POST, OPTIONS",
            "Access-Control-Allow-Headers": "Content-Type, Authorization, X-Requested-With",
            "Access-Control-Allow-Credentials": "true",
            "Access-Control-Max-Age": "86400",
          },
        });
      }
      return new Response(null, { status: 204 });
    }

    // 2. Health check probe for edge diagnostics
    if (url.pathname === "/healthz") {
      return new Response(JSON.stringify({ status: "healthy", timestamp: new Date().toISOString() }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      });
    }

    // 3. Route to GraphFolio Container Instance
    // Route to single stable singleton DO instance (or user-sharded if partitioned)
    const id = env.GRAPHFOLIO_CONTAINER.idFromName("graphfolio-production-instance");
    const containerStub = env.GRAPHFOLIO_CONTAINER.get(id);

    const response = await containerStub.fetch(request);

    // 4. Inject CORS headers on the outgoing response
    const modifiedResponse = new Response(response.body, response);
    if (isAllowedOrigin) {
      modifiedResponse.headers.set("Access-Control-Allow-Origin", origin);
      modifiedResponse.headers.set("Access-Control-Allow-Credentials", "true");
    }

    return modifiedResponse;
  },

  // ---------------------------------------------------------------------------
  // Scheduled Cron Handler for EOD Valuation Job
  // ---------------------------------------------------------------------------
  async scheduled(event: ScheduledEvent, env: Env, ctx: ExecutionContext): Promise<void> {
    console.log(`[cron] Triggering Daily EOD Valuation Job at ${new Date(event.scheduledTime).toISOString()}`);
    
    const id = env.GRAPHFOLIO_CONTAINER.idFromName("graphfolio-production-instance");
    const containerStub = env.GRAPHFOLIO_CONTAINER.get(id);

    const internalJobReq = new Request("http://127.0.0.1:8080/internal/jobs/valuation-eod", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-Internal-Secret": env.INTERNAL_CRON_SECRET,
      },
      body: JSON.stringify({
        asOf: new Date(event.scheduledTime).toISOString().slice(0, 10),
      }),
    });

    const res = await containerStub.fetch(internalJobReq);
    if (!res.ok) {
      console.error(`[cron] Valuation job failed with status ${res.status}: ${await res.text()}`);
    } else {
      console.log(`[cron] Valuation job completed successfully: ${await res.text()}`);
    }
  },
};
```

---

## 5. Database Setup & Cloudflare Hyperdrive Integration

### Why Hyperdrive is Critical
Cloudflare Containers can spin up new instances across multiple regions dynamically. Connecting directly from serverless containers to PostgreSQL creates high connection overhead and TLS handshake latency on every cold start.

**Hyperdrive solves this by**:
1. Maintaining warm, pooled connections to PostgreSQL close to the database server.
2. Handling transparent SSL termination and keep-alives.
3. Automatically accelerating read queries with smart connection pipelining.

### Configuring Hyperdrive via Wrangler CLI

```bash
# 1. Create Hyperdrive pool for Portfolio database schema
npx wrangler hyperdrive create portfolio-hyperdrive \
  --connection-string="postgresql://portfolio_svc:<PASSWORD>@<DB_HOST>:5432/graphfolio?sslmode=require"

# 2. Create Hyperdrive pool for User database schema
npx wrangler hyperdrive create user-hyperdrive \
  --connection-string="postgresql://user_svc:<PASSWORD>@<DB_HOST>:5432/graphfolio?sslmode=require"
```

Inside the container, `PORTFOLIO_DB_URL` and `USER_DB_URL` are provided as environment variables pointing to the Hyperdrive connection strings, which are fully compatible with `pgx/v5`.

---

## 6. Frontend Deployment on Cloudflare Pages

### Build Configuration

GraphFolio frontend is structured as an npm workspaces monorepo under `web/`:
- Investor Portal: `web/apps/main-app` (dist output: `web/apps/main-app/dist`)
- Admin Portal: `web/apps/admin-app` (dist output: `web/apps/admin-app/dist`)

### 1. Investor Portal (`graphfolio.app` / `app.graphfolio.app`)

| Setting | Value |
| :--- | :--- |
| **Framework preset** | Vite |
| **Root directory** | `web` |
| **Build command** | `npm run build:main` |
| **Build output directory** | `apps/main-app/dist` |
| **Environment Variable** | `VITE_GRAPHQL_ENDPOINT=https://api.graphfolio.app/graphql` |

#### SPA Fallback Routing (`web/apps/main-app/public/_routes.json`)
```json
{
  "version": 1,
  "include": ["/*"],
  "exclude": ["/assets/*", "/favicon.ico", "/robots.txt"]
}
```

### 2. Admin Portal (`admin.graphfolio.app`)

| Setting | Value |
| :--- | :--- |
| **Framework preset** | Vite |
| **Root directory** | `web` |
| **Build command** | `npm run build:admin` |
| **Build output directory** | `apps/admin-app/dist` |
| **Environment Variable** | `VITE_GRAPHQL_ENDPOINT=https://api.graphfolio.app/graphql` |
| **Security Layer** | **Cloudflare Zero Trust Access** (enforcing Google/GitHub SSO or corporate IdP before reaching the admin SPA) |

---

## 7. Domain Name Mapping & DNS Configuration (`graphfolio.app`)

This section details how to map your domain `graphfolio.app` across all GraphFolio services using Cloudflare Authoritative DNS, Cloudflare Pages, Cloudflare Workers Custom Domains, and Cloudflare Zero Trust.

### Domain & Subdomain Allocation Matrix

| Hostname | Target Service | Provisioning Mechanism | SSL/TLS Mode | Security Layer |
| :--- | :--- | :--- | :--- | :--- |
| **`graphfolio.app`** (Apex) | Investor Portal | Cloudflare Pages (`graphfolio-main-app`) | Full (strict) Edge Certificate | Public Access with WAF |
| **`app.graphfolio.app`** | Investor Portal (Alt) | Cloudflare Pages (`graphfolio-main-app`) | Full (strict) Edge Certificate | Public Access with WAF |
| **`admin.graphfolio.app`** | Operations Portal | Cloudflare Pages (`graphfolio-admin-app`)| Full (strict) Edge Certificate | **Cloudflare Zero Trust Access** |
| **`api.graphfolio.app`** | GraphQL API / Backend | Cloudflare Worker (`graphfolio-backend`)| Full (strict) Edge Certificate | CORS + WAF Rate Limiting |

### Step-by-Step Domain Mapping Procedures

#### Step 1: Map Frontend SPAs in Cloudflare Pages

##### A. Map the Investor Portal (`graphfolio.app` & `app.graphfolio.app`):
1. In the **Cloudflare Dashboard**, navigate to **Workers & Pages** $\rightarrow$ select your Pages project **`graphfolio-main-app`**.
2. Click the **Custom domains** tab $\rightarrow$ click **Set up a custom domain**.
3. Enter `graphfolio.app` and click **Continue**.
   - Cloudflare automatically manages the apex DNS record via **CNAME flattening** (resolving directly to Cloudflare edge Anycast IPs).
4. Repeat the process to add `app.graphfolio.app`:
   - Enter `app.graphfolio.app` $\rightarrow$ click **Continue** $\rightarrow$ click **Activate domain**.
   - Cloudflare creates a CNAME pointing `app` to `graphfolio-main-app.pages.dev` with proxy enabled (orange cloud).

##### B. Map the Admin Portal (`admin.graphfolio.app`):
1. Navigate to **Workers & Pages** $\rightarrow$ select your Pages project **`graphfolio-admin-app`**.
2. Click the **Custom domains** tab $\rightarrow$ click **Set up a custom domain**.
3. Enter `admin.graphfolio.app` $\rightarrow$ click **Continue** $\rightarrow$ click **Activate domain**.
   - Cloudflare creates a CNAME pointing `admin` to `graphfolio-admin-app.pages.dev` with proxy enabled.

*(Alternative via CLI)*:
```bash
npx wrangler pages domain add graphfolio-main-app graphfolio.app
npx wrangler pages domain add graphfolio-main-app app.graphfolio.app
npx wrangler pages domain add graphfolio-admin-app admin.graphfolio.app
```

---

#### Step 2: Map the GraphQL API Gateway (`api.graphfolio.app`)

Because the Backend-for-Frontend is powered by a Cloudflare Worker bound to the container pod, map `api.graphfolio.app` directly as a **Worker Custom Domain**:

##### Method A: Declarative via `wrangler.jsonc` (Recommended)
Add the route definition to your `wrangler.jsonc`:
```jsonc
"routes": [
  {
    "pattern": "api.graphfolio.app/*",
    "custom_domain": true
  }
]
```
When running `npx wrangler deploy`, Wrangler automatically registers `api.graphfolio.app` as a Custom Domain on the Worker and creates the DNS record in your Cloudflare zone.

##### Method B: Via Cloudflare Dashboard
1. Go to **Workers & Pages** $\rightarrow$ select `graphfolio-backend`.
2. Navigate to **Settings** $\rightarrow$ **Triggers** $\rightarrow$ **Custom Domains**.
3. Click **Add Custom Domain** $\rightarrow$ enter `api.graphfolio.app` $\rightarrow$ click **Add Custom Domain**.
4. Cloudflare provisions a dedicated edge certificate for `api.graphfolio.app` and routes all incoming HTTP traffic directly to your Worker gateway.

---

#### Step 3: Cloudflare DNS Zone Records Verification

Ensure your Cloudflare DNS table reflects the following entries (all proxied through Cloudflare with the orange cloud active):

| Type | Name | Content / Target | Proxy Status | TTL | Description |
| :--- | :--- | :--- | :---: | :---: | :--- |
| **CNAME** | `@` (`graphfolio.app`) | `graphfolio-main-app.pages.dev` | **Proxied** | Auto | Apex Investor Portal (Flattened) |
| **CNAME** | `app` | `graphfolio-main-app.pages.dev` | **Proxied** | Auto | Subdomain Investor Portal |
| **CNAME** | `admin` | `graphfolio-admin-app.pages.dev` | **Proxied** | Auto | Admin Portal (Zero Trust protected) |
| **DNS Record** | `api` | Managed by Worker Custom Domain | **Proxied** | Auto | GraphQL API & Container Gateway |

---

#### Step 4: Configure SSL/TLS & Edge Security for `graphfolio.app`

In the Cloudflare Dashboard for `graphfolio.app`:
1. **SSL/TLS Encryption Mode**:
   - Go to **SSL/TLS** $\rightarrow$ **Overview** $\rightarrow$ select **Full (strict)**.
   - Ensures end-to-end encryption between browser, Cloudflare edge, and origin services.
2. **Edge Certificates**:
   - Go to **SSL/TLS** $\rightarrow$ **Edge Certificates**.
   - Enable **Always Use HTTPS** (automatically redirects all `http://` traffic to `https://`).
   - Enable **Automatic HTTPS Rewrites**.
   - Set **Minimum TLS Version** to **TLS 1.2** (or TLS 1.3).
3. **HTTP Strict Transport Security (HSTS)**:
   - Max-Age: `12 months` (`31536000` seconds).
   - Include subdomains: **Enabled**.
   - Preload: **Enabled**.

---

#### Step 5: Lock Down `admin.graphfolio.app` with Cloudflare Zero Trust

Protect the administrative operations portal so only authorized personnel can access it:
1. In Cloudflare Dashboard, go to **Zero Trust** $\rightarrow$ **Access** $\rightarrow$ **Applications**.
2. Click **Add an application** $\rightarrow$ select **Self-hosted**.
3. Fill in Application Configuration:
   - **Application name**: `GraphFolio Admin Portal`
   - **Session Duration**: `24 hours`
   - **Application domain**: `admin.graphfolio.app`
4. Add an Access Policy:
   - **Policy name**: `Authorized Operations Team`
   - **Action**: `Allow`
   - **Include Rule**: `Emails ending in: @graphfolio.app` (or specific administrator emails).
5. Identity Provider:
   - Enable Google Workspace, GitHub, or One-Time PIN (OTP).
6. Save Application. Unauthenticated visitors to `https://admin.graphfolio.app` will be redirected to the Cloudflare Access login screen before reaching the React SPA.

---

## 8. Security, Networking & Secrets Management

GraphFolio manages sensitive credentials and API keys using **Cloudflare Encrypted Secrets**, **`.dev.vars` for local development**, and **GitHub Secrets for automated CI/CD**. Sensitive keys are never committed to Git, never exposed in build artifacts, and never baked into Docker images.

### Secrets Architecture & Ingestion Flow

```
Local Development:
  .dev.vars (Gitignored) ─────────► npx wrangler dev ─────────► Local Container Env

Production Deployment:
  GitHub Secrets (CI/CD)
           │
           ▼
  wrangler secret put ────────────► Cloudflare Secrets Store (AES-256 Encrypted)
                                              │
                                              ▼
                                   Cloudflare Worker (env.*)
                                              │
                                              ▼
                              Container Process (os.Getenv)
                              ├── PORTFOLIO_DB_URL
                              ├── USER_DB_URL
                              ├── JWT_SECRET
                              └── TWELVE_DATA_API_KEY
```

### 1. Required Secrets & API Keys Inventory

| Secret Name | Purpose | Target Component | Sensitivity |
| :--- | :--- | :--- | :--- |
| `JWT_SECRET` | 256-bit HMAC-SHA256 signing key for investor session access tokens | `user-api` & `bff` | **Critical** (Protects user authentication) |
| `PORTFOLIO_DB_URL` | PostgreSQL connection string or Hyperdrive pool URI with SSL enforced | `portfolio-api` & `cmd/worker` | **Critical** (Direct access to `portfolio` schema) |
| `USER_DB_URL` | PostgreSQL connection string or Hyperdrive pool URI with SSL enforced | `user-api` | **Critical** (Direct access to `user` schema) |
| `TWELVE_DATA_API_KEY` | API key for Twelve Data market price and FX ingestion pipelines | `portfolio-api` | High (Third-party metered provider) |
| `INTERNAL_CRON_SECRET` | Shared 64-character bearer token authenticating nightly valuation cron calls | Worker $\rightarrow$ Container `bff` | High (Restricts batch job invocation) |

### 2. Setting Production Secrets

#### Method A: Via Wrangler CLI (Interactive / Direct)
Execute the following commands to encrypt and store each secret in Cloudflare's secure key management service:

```bash
# 1. JWT Signing Secret (minimum 256 bits / 32 characters)
npx wrangler secret put JWT_SECRET

# 2. Database Connection Strings (with SSL required)
npx wrangler secret put PORTFOLIO_DB_URL
npx wrangler secret put USER_DB_URL

# 3. Market Data Ingestion API Key (Twelve Data)
npx wrangler secret put TWELVE_DATA_API_KEY

# 4. Internal Cron Secret (Authenticates worker-to-container scheduled invocations)
npx wrangler secret put INTERNAL_CRON_SECRET
```

#### Method B: Via Cloudflare Web Dashboard
1. Navigate to **Cloudflare Dashboard** $\rightarrow$ **Workers & Pages**.
2. Select your Worker project (`graphfolio-backend`).
3. Go to **Settings** $\rightarrow$ **Variables and Secrets**.
4. Under **Secrets**, click **Add** for each key and ensure encryption is checked.

### 3. How Secrets Reach the Go Container Process

Cloudflare Containers run as Durable Objects integrated with the Worker runtime. Secrets stored in Cloudflare's Secret Store are bound directly to the container's OS environment at container launch:

1. **Wrangler Declaration (`wrangler.jsonc`)**:
   Non-sensitive settings are listed in the `vars` block. Secrets configured via `wrangler secret put` are injected automatically into the container runtime without appearing in `wrangler.jsonc`.
2. **Go Standard Library Access (`os.Getenv`)**:
   The Go microservices read secrets natively via standard environment variables:
   ```go
   // Read database configuration
   dbCfg, err := database.ConfigFromEnv("PORTFOLIO_DB_URL")
   
   // Read market data API key
   twelveKey := os.Getenv("TWELVE_DATA_API_KEY")
   
   // Read JWT signing secret
   jwtSecret := os.Getenv("JWT_SECRET")
   ```

### 4. Local Development Secret Management (`.dev.vars`)

During local testing with `npx wrangler dev`, Wrangler does not contact the remote Cloudflare API. Instead, it reads a local `.dev.vars` file located in the workspace root:

```ini
# .dev.vars (REQUIRED FOR LOCAL TESTING - DO NOT COMMIT TO GIT)
JWT_SECRET=dev_jwt_secret_must_be_at_least_32_characters_long_123456
PORTFOLIO_DB_URL=postgresql://portfolio_svc:portfolio_dev_pass@localhost:5432/graphfolio?sslmode=disable
USER_DB_URL=postgresql://user_svc:user_dev_pass@localhost:5432/graphfolio?sslmode=disable
TWELVE_DATA_API_KEY=demo
INTERNAL_CRON_SECRET=local_dev_cron_secret_64chars_abcdef1234567890
```

Add `.dev.vars` to `.gitignore` to prevent accidental credential leakage:
```gitignore
# Security & Credentials
.env
.dev.vars
*.pem
*.key
```

### 5. Cookie & Authentication Domain Security
Because the frontend runs on `https://graphfolio.app` (and `https://app.graphfolio.app`) and the API runs on `https://api.graphfolio.app`:
- Set `SameSite=None` and `Secure=true` for the `refresh_token` HTTP-only cookie.
- Set `Domain=.graphfolio.app` so the refresh token is accessible across the root domain and subdomains.
- Ensure `Credentials: 'include'` is set on all GraphQL client fetches in `@graphfolio/api-client`.

---

## 9. Continuous Integration & Deployment (CI/CD)

### GitHub Actions Pipeline (`.github/workflows/deploy-cloudflare.yml`)

```yaml
name: Deploy GraphFolio to Cloudflare

on:
  push:
    branches: [main]
  workflow_dispatch:

jobs:
  test:
    name: Backend & Frontend Tests
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go 1.24
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'
          cache: true

      - name: Run Go Unit Tests
        run: |
          go test ./pkg/...
          go test ./services/portfolio-api/...
          go test ./services/user-api/...
          go test ./bff/...

      - name: Set up Node.js 22
        uses: actions/setup-node@v4
        with:
          node-version: 22
          cache: 'npm'
          cache-dependency-path: web/package-lock.json

      - name: Run Web Tests & Typecheck
        run: |
          cd web
          npm ci
          npm run test
          npm run lint

  deploy-container:
    name: Deploy Container to Cloudflare
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Deploy to Cloudflare Workers with Containers
        uses: cloudflare/wrangler-action@v3
        with:
          apiToken: ${{ secrets.CLOUDFLARE_API_TOKEN }}
          accountId: ${{ secrets.CLOUDFLARE_ACCOUNT_ID }}
          command: deploy
          secrets: |
            JWT_SECRET
            PORTFOLIO_DB_URL
            USER_DB_URL
            TWELVE_DATA_API_KEY
            INTERNAL_CRON_SECRET
        env:
          JWT_SECRET: ${{ secrets.PROD_JWT_SECRET }}
          PORTFOLIO_DB_URL: ${{ secrets.PROD_PORTFOLIO_DB_URL }}
          USER_DB_URL: ${{ secrets.PROD_USER_DB_URL }}
          TWELVE_DATA_API_KEY: ${{ secrets.PROD_TWELVE_DATA_API_KEY }}
          INTERNAL_CRON_SECRET: ${{ secrets.PROD_INTERNAL_CRON_SECRET }}

  deploy-frontend:
    name: Deploy Frontend Pages
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Node.js 22
        uses: actions/setup-node@v4
        with:
          node-version: 22

      - name: Build Main App
        run: |
          cd web
          npm ci
          npm run build:main

      - name: Deploy Main App to Cloudflare Pages
        uses: cloudflare/pages-action@v1
        with:
          apiToken: ${{ secrets.CLOUDFLARE_API_TOKEN }}
          accountId: ${{ secrets.CLOUDFLARE_ACCOUNT_ID }}
          projectName: graphfolio-main-app
          directory: web/apps/main-app/dist
          gitHubToken: ${{ secrets.GITHUB_TOKEN }}
```

---

## 10. Step-by-Step Implementation Roadmap

```mermaid
gantt
    title Cloudflare Deployment Implementation Phases
    dateFormat  YYYY-MM-DD
    section Phase 1: Containerization
    Create Multi-stage Dockerfile          :p1_1, 2026-10-16, 2d
    Create Entrypoint & Process Supervisor :p1_2, after p1_1, 1d
    Local Docker & Compose Verification    :p1_3, after p1_2, 2d
    section Phase 2: Cloudflare Infra
    Provision Managed PostgreSQL (Neon/RDS):p2_1, 2026-10-17, 1d
    Configure Cloudflare Hyperdrive Pools  :p2_2, after p2_1, 1d
    Set up Cloudflare Secrets & Policies   :p2_3, after p2_2, 1d
    section Phase 3: Worker & Gateway
    Build wrangler.jsonc with DO Bindings  :p3_1, after p1_3, 2d
    Implement src/index.ts Worker Gateway  :p3_2, after p3_1, 2d
    Internal Valuation EOD Webhook Handler :p3_3, after p3_2, 1d
    section Phase 4: Frontend Pages
    Deploy main-app to Cloudflare Pages    :p4_1, after p3_2, 1d
    Deploy admin-app with Zero Trust Access:p4_2, after p4_1, 1d
    section Phase 5: Verification
    End-to-End Smoke Tests & Audit         :p5_1, after p4_2, 2d
    DNS Cutover to Production Custom Domain:p5_2, after p5_1, 1d
```

### Detailed Phase Breakdown

#### Phase 1: Local Docker Containerization & Multi-Service Orchestration
- [ ] Create `deploy/docker/Dockerfile` with multi-stage Go build and minimal Alpine runtime.
- [ ] Create `deploy/docker/entrypoint.sh` implementing graceful signal forwarding and process supervision for `portfolio-api`, `user-api`, and `bff`.
- [ ] Add `/healthz` HTTP endpoint to `bff/cmd/server/main.go` returning readiness of both gRPC connections.
- [ ] Add internal webhook endpoint `/internal/jobs/valuation-eod` in BFF calling `portfolio-api` valuation service.
- [ ] Test locally using `docker build -f deploy/docker/Dockerfile -t graphfolio-backend:local .` and verify all 3 services communicate over loopback.

#### Phase 2: Cloudflare Infrastructure & Hyperdrive Provisioning
- [ ] Provision managed PostgreSQL 15+ instance with SSL enforced.
- [ ] Run database bootstrap script (`scripts/db/bootstrap.sql`) to create schemas and users (`portfolio_svc`, `user_svc`).
- [ ] Apply initial schema migrations using `golang-migrate` CLI.
- [ ] Create Cloudflare Hyperdrive bindings:
  - `portfolio-hyperdrive`
  - `user-hyperdrive`
- [ ] Validate database connectivity and transaction compatibility with `shopspring/decimal`.

#### Phase 3: Cloudflare Worker Gateway & Container DO Integration
- [ ] Initialize Worker project directory with `wrangler.jsonc`.
- [ ] Configure `containers` and `durable_objects` definitions pointing to `./deploy/docker`.
- [ ] Write `src/index.ts` implementing `DurableObject` with `this.ctx.container.getTcpPort(8080).fetch(req)`.
- [ ] Implement CORS header handling and HTTP preflight negotiation at the edge.
- [ ] Add `scheduled` cron trigger handler invoking EOD valuation.
- [ ] Test container deployment using `npx wrangler dev` with local Docker engine.

#### Phase 4: Frontend Deployment on Cloudflare Pages
- [ ] Configure `graphfolio-main-app` on Cloudflare Pages connected to Git repository.
- [ ] Set build command `npm run build:main` and output directory `web/apps/main-app/dist`.
- [ ] Configure `_routes.json` for client-side React routing.
- [ ] Configure `graphfolio-admin-app` on Cloudflare Pages.
- [ ] Configure Cloudflare Zero Trust Access rule for `admin.graphfolio.app` restricting access to authorized team emails.

#### Phase 5: Production Verification, Telemetry & DNS Cutover
- [ ] Run full smoke testing suite against `https://api.graphfolio.app`.
- [ ] Verify authentication flow (Register -> Login -> Set-Cookie -> Refresh Token -> Logout).
- [ ] Verify portfolio calculations, ledger projections, and cash flow reports.
- [ ] Trigger manual EOD valuation via cron test event.
- [ ] Set up Cloudflare Logpush and Tail Workers for real-time observability.
- [ ] Update DNS records in Cloudflare Dashboard for custom domains (`graphfolio.app`, `app.graphfolio.app`, `admin.graphfolio.app`, `api.graphfolio.app`).

---

## 11. Smoke Testing & Verification Checklist

| Test Item | Verification Method | Expected Result |
| :--- | :--- | :--- |
| **Container Boot & Health** | `curl -f https://api.graphfolio.app/healthz` | HTTP 200 `{"status":"healthy"}` |
| **GraphQL Schema Introspection** | `curl -X POST https://api.graphfolio.app/graphql -H "Content-Type: application/json" -d '{"query":"{ __typename }"}'` | HTTP 200 `{"data":{"__typename":"Query"}}` |
| **Argon2id Authentication** | Submit `mutation { login(email, password) }` | HTTP 200 with JWT Bearer token and `httpOnly` secure `refresh_token` cookie |
| **gRPC Loopback Communication** | Query `me { portfolio { summary { totalValue { amount } } } }` | BFF successfully calls `portfolio-api` and returns exact decimal calculations |
| **Hyperdrive Pool Latency** | Measure query roundtrip via Cloudflare Trace / Workers Analytics | P95 latency `< 25ms` across warm edge invocations |
| **CORS Validation** | Send `OPTIONS /graphql` with `Origin: https://graphfolio.app` | HTTP 204 with `Access-Control-Allow-Origin: https://graphfolio.app` and credentials flag |
| **Scheduled Valuation Cron** | Trigger test cron event via `wrangler triggers fire --cron "0 22 * * 1-5"` | Worker invokes container and logs successful valuation snapshot creation |
| **Admin Zero Trust Protection** | Access `https://admin.graphfolio.app` from unauthenticated browser | Redirected to Cloudflare Access login screen |

---

## 12. Cost & Resource Optimization Guidelines

1. **Scale-to-Zero Efficiency**:
   - Cloudflare Containers scale down to zero instances when traffic ceases. Setting DO container wake timeouts ensures compute is only billed during active user sessions and daily EOD valuation processing.
2. **Binary Stripping**:
   - Compiling Go binaries with `-ldflags="-s -w"` and using Alpine Linux produces an image under **65MB**, minimizing cold-start pull times to `< 1.2s`.
3. **Hyperdrive Connection Multiplexing**:
   - Hyperdrive keeps serverless database connection counts low (avoiding expensive database tier upgrades strictly for connection limits).
4. **Static Asset Caching**:
   - Cloudflare Pages caches Vite bundles with immutable 1-year cache headers (`/assets/*`), completely offloading frontend static traffic from container compute.
