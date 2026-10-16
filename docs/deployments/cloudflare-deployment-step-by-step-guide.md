# Cloudflare Deployment: Step-by-Step Implementation Guide

> **Target**: Comprehensive operational manual and step-by-step implementation guide for deploying GraphFolio onto the **Cloudflare Container Platform** (Workers with Containers, Durable Objects, Cloudflare Pages, and Cloudflare Hyperdrive).  
> **Target Domain**: `https://graphfolio.app/`  
> **Companion Plan**: [cloudflare-containers-deployment-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/deployments/cloudflare-containers-deployment-plan.md)  
> **Key Principle**: Clear, copy-pasteable CLI commands, concrete verification checks, zero-guesswork secret management, and robust production hardening.

---

## Table of Contents
1. [Prerequisites & Tooling](#1-prerequisites--tooling)
2. [Step 1: Local Docker Container Build & Verification](#2-step-1-local-docker-container-build--verification)
3. [Step 2: Database Provisioning & Schema Bootstrap](#3-step-2-database-provisioning--schema-bootstrap)
4. [Step 3: Cloudflare Hyperdrive Setup](#4-step-3-cloudflare-hyperdrive-setup)
5. [Step 4: Local Cloudflare Worker & Container Testing](#5-step-4-local-cloudflare-worker--container-testing)
6. [Step 5: Cloudflare Secrets Provisioning](#6-step-5-cloudflare-secrets-provisioning)
7. [Step 6: Deploying the Backend Container](#7-step-6-deploying-the-backend-container)
8. [Step 7: Deploying Frontend SPAs to Cloudflare Pages](#8-step-7-deploying-frontend-spas-to-cloudflare-pages)
9. [Step 8: Domain Mapping & DNS Setup (graphfolio.app)](#9-step-8-domain-mapping--dns-setup-graphfolioapp)
10. [Step 9: Lock Down Admin Portal with Cloudflare Zero Trust](#10-step-9-lock-down-admin-portal-with-cloudflare-zero-trust)
11. [Step 10: Automated CI/CD Setup with GitHub Actions](#11-step-10-automated-cicd-setup-with-github-actions)
12. [Troubleshooting & Operational Runbook](#12-troubleshooting--operational-runbook)

---

## 1. Prerequisites & Tooling

Ensure the following tools are installed on your workstation:
- **Docker / Docker Desktop** (running and operational)
- **Node.js 22+** and **npm** (`node -v && npm -v`)
- **Go 1.26+** (`go version`)
- **Cloudflare Wrangler CLI**:
  ```bash
  npm install -g wrangler
  # Verify authentication
  wrangler login
  ```
- **PostgreSQL Client Tools** (`psql` and `pg_isready`)
- **Active Cloudflare Account** with an active DNS zone for `graphfolio.app`.

---

## 2. Step 1: Local Docker Container Build & Verification

Before deploying to Cloudflare, verify that the unified multi-process container builds and runs properly on your local Docker engine.

### 1. Build the Docker Image
From the repository root:
```bash
make docker-build
```
*Expected Output*: Docker will multi-stage compile `bff`, `portfolio-api`, `user-api`, and `portfolio-worker` using Go on Alpine Linux, producing the tagged image `graphfolio-backend:local` (~65MB).

### 2. Run the Container Locally
Ensure your local PostgreSQL instance is running or use your development environment variables:
```bash
make docker-run
```
*Expected Output*:
```
[graphfolio-init] Starting GraphFolio Unified Backend Services...
[graphfolio-init] Launching User API on port 50052...
[graphfolio-init] Launching Portfolio API on port 50051...
[graphfolio-init] Awaiting gRPC services loopback readiness...
[graphfolio-init] Internal gRPC microservices are healthy and accepting connections.
[graphfolio-init] Launching GraphQL BFF on port 8080...
```

### 3. Verify Health and Endpoints in Another Terminal
```bash
# Test Container Health Probe
curl -i http://localhost:8080/healthz
# Expected: HTTP/1.1 200 OK {"status":"healthy"}

# Test GraphQL Introspection Query
curl -i -X POST http://localhost:8080/query \
  -H "Content-Type: application/json" \
  -d '{"query":"{ __typename }"}'
# Expected: HTTP/1.1 200 OK {"data":{"__typename":"Query"}}
```

Stop the container with `Ctrl+C` (the supervisor trap will cleanly terminate all child processes).

---

## 3. Step 2: Database Provisioning & Schema Bootstrap

Cloudflare Containers are stateless compute instances. GraphFolio requires a PostgreSQL 15+ database (e.g. **Neon Serverless Postgres**, **Supabase**, or **AWS RDS**).

### 1. Create a Managed PostgreSQL Database
Sign up for [Neon](https://neon.tech/) or [Supabase](https://supabase.com/) and create a PostgreSQL 15+ database named `graphfolio`. Ensure connection pooling and SSL are enabled.

### 2. Run the Database Bootstrap Script
Execute the repository bootstrap script to create the required roles (`portfolio_svc`, `user_svc`) and schemas (`portfolio`, `user`):
```bash
# Substitute your admin connection string
export PG_ADMIN_URL="postgresql://<ADMIN_USER>:<PASSWORD>@<HOST>/postgres?sslmode=require"

psql "$PG_ADMIN_URL" -v db=graphfolio -f scripts/db/bootstrap.sql
```

### 3. Apply Initial Schema Migrations
```bash
# Run migrations using golang-migrate CLI
export PORTFOLIO_DB_URL="postgresql://portfolio_svc:<PORTFOLIO_PASS>@<HOST>/graphfolio?sslmode=require"
export USER_DB_URL="postgresql://user_svc:<USER_PASS>@<HOST>/graphfolio?sslmode=require"

migrate -path services/portfolio-api/migrations -database "${PORTFOLIO_DB_URL}&search_path=portfolio&x-migrations-table=schema_migrations" up
migrate -path services/user-api/migrations      -database "${USER_DB_URL}&search_path=users&x-migrations-table=schema_migrations" up
```

---

## 4. Step 3: Cloudflare Hyperdrive Setup

Cloudflare Hyperdrive accelerates database queries and reuses pooled connections across Cloudflare's global edge network, eliminating connection exhaustion during container scale-outs.

### 1. Create the Portfolio Database Pool
```bash
npx wrangler hyperdrive create portfolio-hyperdrive \
  --connection-string="postgresql://portfolio_svc:<PASSWORD>@<HOST>:5432/graphfolio?sslmode=require"
```
*Note the returned Hyperdrive ID* (e.g., `a1b2c3d4e5f6...`).

### 2. Create the User Database Pool
```bash
npx wrangler hyperdrive create user-hyperdrive \
  --connection-string="postgresql://user_svc:<PASSWORD>@<HOST>:5432/graphfolio?sslmode=require"
```
*Note the returned Hyperdrive ID*.

### 3. Update `deploy/cloudflare/wrangler.jsonc` (Optional)
If using Hyperdrive bindings, update the `id` fields in [deploy/cloudflare/wrangler.jsonc](file:///Users/oscargarcia/workspace/graphfolio/deploy/cloudflare/wrangler.jsonc):
```jsonc
"hyperdrive": [
  {
    "binding": "PORTFOLIO_HYPERDRIVE",
    "id": "a1b2c3d4e5f6..."
  },
  {
    "binding": "USER_HYPERDRIVE",
    "id": "f6e5d4c3b2a1..."
  }
]
```

---

## 5. Step 4: Local Cloudflare Worker & Container Testing

You can run both the Cloudflare Worker gateway and the Docker container locally using Wrangler.

### 1. Configure Local Secrets (`.dev.vars`)
Copy the example file to `.dev.vars` inside `deploy/cloudflare/`:
```bash
cp deploy/cloudflare/.dev.vars.example deploy/cloudflare/.dev.vars
```
Edit `deploy/cloudflare/.dev.vars` with your actual development database credentials and API keys.

### 2. Start Wrangler Dev with Container
```bash
make cf-dev
```
*Expected Output*: Wrangler starts the local edge worker on `http://localhost:8787` and spins up the local container instance.

### 3. Verify Edge Gateway
In another terminal:
```bash
# Edge Health Check
curl http://localhost:8787/healthz

# Edge Proxied GraphQL Request
curl -X POST http://localhost:8787/query \
  -H "Content-Type: application/json" \
  -H "Origin: https://graphfolio.app" \
  -d '{"query":"{ __typename }"}'
```

---

## 6. Step 5: Cloudflare Secrets Provisioning

Store production secrets in Cloudflare's encrypted Secrets Store. These values are never stored in plaintext and are injected dynamically into the container process environment at runtime.

Navigate to `deploy/cloudflare` and execute:

```bash
cd deploy/cloudflare

# 1. JWT Signing Secret (Generate with: openssl rand -hex 32)
npx wrangler secret put JWT_SECRET

# 2. Production Database URLs (Pointing to your managed PostgreSQL instance)
npx wrangler secret put PORTFOLIO_DB_URL
npx wrangler secret put USER_DB_URL

# 3. Twelve Data Market Ingestion API Key
npx wrangler secret put TWELVE_DATA_API_KEY

# 4. Internal Cron Secret (Random 64-char token for scheduled EOD valuation)
npx wrangler secret put INTERNAL_CRON_SECRET

cd ../..
```

---

## 7. Step 6: Deploying the Backend Container

Deploy the Cloudflare Worker Gateway and the container image:

```bash
make cf-deploy
```

*What happens during deployment*:
1. Wrangler packages the Worker gateway (`deploy/cloudflare/src/index.ts`).
2. Wrangler builds or packages the Docker image from `deploy/docker/Dockerfile`.
3. Pushes the container image to the Cloudflare-managed container registry.
4. Binds the `GraphFolioContainer` Durable Object.
5. Sets up the Cron Trigger (`0 22 * * 1-5`) for market-close valuation.
6. Maps the custom domain route `api.graphfolio.app/*`.

### Verification:
```bash
curl -f https://api.graphfolio.app/healthz
# Expected: {"status":"healthy"}
```

---

## 8. Step 7: Deploying Frontend SPAs to Cloudflare Pages

Deploy the Investor Portal (`main-app`) and Admin Portal (`admin-app`) to Cloudflare Pages.

### 1. Create the Cloudflare Pages Projects
```bash
npx wrangler pages project create graphfolio-main-app --production-branch=main
npx wrangler pages project create graphfolio-admin-app --production-branch=main
```

### 2. Build and Deploy Main Investor Portal (`main-app`)
```bash
cd web

# 1. Install dependencies
npm ci

# 2. Build main-app targeting the production API endpoint
VITE_GRAPHQL_ENDPOINT=https://api.graphfolio.app/query npm run build:main

# 3. Publish to Cloudflare Pages
npx wrangler pages deploy apps/main-app/dist --project-name=graphfolio-main-app

cd ..
```

### 3. Build and Deploy Admin Portal (`admin-app`)
```bash
cd web

# 1. Build admin-app targeting the production API endpoint
VITE_GRAPHQL_ENDPOINT=https://api.graphfolio.app/query npm run build:admin

# 2. Publish to Cloudflare Pages
npx wrangler pages deploy apps/admin-app/dist --project-name=graphfolio-admin-app

cd ..
```

*Note*: The `_routes.json` files already placed in `web/apps/main-app/public/_routes.json` and `web/apps/admin-app/public/_routes.json` ensure client-side Single-Page Application (SPA) routing functions properly.

---

## 9. Step 8: Domain Mapping & DNS Setup (`graphfolio.app`)

Now link your domain `graphfolio.app` to all three services in your Cloudflare dashboard:

### 1. Map Investor Portal to Apex & Subdomain
1. In Cloudflare Dashboard, go to **Workers & Pages** $\rightarrow$ select **`graphfolio-main-app`**.
2. Click **Custom domains** $\rightarrow$ **Set up a custom domain**.
3. Add **`graphfolio.app`** (Apex) $\rightarrow$ Cloudflare configures CNAME flattening automatically.
4. Click **Set up a custom domain** again $\rightarrow$ Add **`app.graphfolio.app`**.

*(CLI shortcut)*:
```bash
npx wrangler pages domain add graphfolio-main-app graphfolio.app
npx wrangler pages domain add graphfolio-main-app app.graphfolio.app
```

### 2. Map Admin Portal
1. Go to **Workers & Pages** $\rightarrow$ select **`graphfolio-admin-app`**.
2. Click **Custom domains** $\rightarrow$ **Set up a custom domain**.
3. Add **`admin.graphfolio.app`**.

*(CLI shortcut)*:
```bash
npx wrangler pages domain add graphfolio-admin-app admin.graphfolio.app
```

### 3. Map API Gateway Subdomain (`api.graphfolio.app`)
Because `wrangler.jsonc` includes:
```jsonc
"routes": [
  { "pattern": "api.graphfolio.app/*", "custom_domain": true }
]
```
Running `make cf-deploy` automatically binds `api.graphfolio.app` to your backend container worker.

### 4. Verify Cloudflare DNS Zone Records
In your Cloudflare Dashboard under **DNS** $\rightarrow$ **Records**, confirm the table:

| Type | Name | Content | Proxy status |
| :--- | :--- | :--- | :---: |
| **CNAME** | `@` (`graphfolio.app`) | `graphfolio-main-app.pages.dev` | **Proxied (Orange Cloud)** |
| **CNAME** | `app` | `graphfolio-main-app.pages.dev` | **Proxied (Orange Cloud)** |
| **CNAME** | `admin` | `graphfolio-admin-app.pages.dev` | **Proxied (Orange Cloud)** |
| **DNS Record** | `api` | Managed by Worker Custom Domain | **Proxied (Orange Cloud)** |

### 5. Enforce Full Strict SSL/TLS
1. In the Cloudflare Dashboard for `graphfolio.app`, go to **SSL/TLS** $\rightarrow$ **Overview**.
2. Select **Full (strict)**.
3. Under **Edge Certificates**:
   - Turn **Always Use HTTPS** on.
   - Turn **Automatic HTTPS Rewrites** on.
   - Set **Minimum TLS Version** to **TLS 1.2**.

---

## 10. Step 9: Lock Down Admin Portal with Cloudflare Zero Trust

The Admin Portal (`admin.graphfolio.app`) has access to asset closing price overrides, master instrument catalogs, and market backfills. Protect it using Cloudflare Zero Trust:

1. In the Cloudflare Dashboard, click **Zero Trust** in the left sidebar.
2. Navigate to **Access** $\rightarrow$ **Applications** $\rightarrow$ click **Add an application**.
3. Select **Self-hosted**.
4. Configure Application Details:
   - **Application name**: `GraphFolio Internal Admin`
   - **Session Duration**: `24 hours`
   - **Application domain**: `admin.graphfolio.app`
5. Configure Policy:
   - **Policy name**: `Authorized Finance & Operations Team`
   - **Action**: `Allow`
   - **Include Rule**: Add `Emails` (e.g. `your-email@example.com`) or `Email domain` (`@graphfolio.app`).
6. Configure Identity Providers: Enable Google Workspace, GitHub, or One-Time PIN.
7. Click **Save Application**.

*Verification*: Open `https://admin.graphfolio.app` in an incognito window. You must see the Cloudflare Zero Trust login prompt before any application code loads.

---

## 11. Step 10: Automated CI/CD Setup with GitHub Actions

The repository includes a ready-to-use GitHub Actions workflow in [.github/workflows/deploy-cloudflare.yml](file:///Users/oscargarcia/workspace/graphfolio/.github/workflows/deploy-cloudflare.yml).

### 1. Add Repository Secrets in GitHub
Navigate to your GitHub repository $\rightarrow$ **Settings** $\rightarrow$ **Secrets and variables** $\rightarrow$ **Actions** $\rightarrow$ **New repository secret**:

| Secret Name | Value Description |
| :--- | :--- |
| `CLOUDFLARE_API_TOKEN` | API Token created in Cloudflare Dashboard with `Edit Cloudflare Workers` and `Edit Cloudflare Pages` permissions |
| `CLOUDFLARE_ACCOUNT_ID` | Your Cloudflare Account ID (found on the right sidebar of the Workers & Pages dashboard) |
| `PROD_JWT_SECRET` | 256-bit signing key for JWT tokens |
| `PROD_PORTFOLIO_DB_URL` | Production PostgreSQL connection string for the `portfolio` schema |
| `PROD_USER_DB_URL` | Production PostgreSQL connection string for the `user` schema |
| `PROD_TWELVE_DATA_API_KEY` | Twelve Data API Key |
| `PROD_INTERNAL_CRON_SECRET`| Shared secret for scheduled EOD valuation execution |

### 2. Trigger Automated Deployment
Commit and push your changes to `main`:
```bash
git add .
git commit -m "feat: configure cloudflare deployment artifacts"
git push origin main
```
GitHub Actions will automatically run backend unit tests, frontend typechecks, build the container, deploy the container to Cloudflare Workers, and deploy both web applications to Cloudflare Pages.

---

## 12. Troubleshooting & Operational Runbook

### 1. Real-Time Log Streaming
Stream live container and Worker logs directly to your terminal:
```bash
cd deploy/cloudflare
npx wrangler tail
```

### 2. Manually Triggering the Nightly Valuation Job
Test the valuation worker without waiting for the 22:00 UTC cron trigger:
```bash
# In deploy/cloudflare
npx wrangler triggers fire --cron "0 22 * * 1-5"
```
Or send an authenticated HTTP request directly:
```bash
curl -X POST https://api.graphfolio.app/internal/jobs/valuation-eod \
  -H "Content-Type: application/json" \
  -H "X-Internal-Secret: <YOUR_INTERNAL_CRON_SECRET>" \
  -d '{"asOf":"2026-10-16"}'
```

### 3. Common Issues & Solutions

| Symptom | Cause | Solution |
| :--- | :--- | :--- |
| **503 Service Unavailable** on API requests | The internal gRPC microservices haven't completed startup before BFF probe. | Check container logs via `wrangler tail`. Ensure `PORTFOLIO_DB_URL` and `USER_DB_URL` are accessible from the container with SSL enabled. |
| **CORS Error in Browser Console** | Incoming request `Origin` header does not match `CORS_ORIGINS`. | Check `CORS_ORIGINS` in `wrangler.jsonc`. Ensure it includes `https://graphfolio.app` and `https://app.graphfolio.app`. |
| **Invalid Credentials on Database** | Database SSL certificate validation failure. | Append `?sslmode=require` to `PORTFOLIO_DB_URL` and `USER_DB_URL`. |
| **404 on Page Refresh in React** | Missing SPA rewrite rule on Cloudflare Pages. | Verify that `_routes.json` is present in `web/apps/main-app/public/` and `web/apps/admin-app/public/`. |
