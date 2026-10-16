# GraphFolio Cloud Deployment Cost Estimates & Comparison

> **Target**: Comprehensive cost estimation, side-by-side analysis, and Total Cost of Ownership (TCO) evaluation for the four cloud deployment architectures documented in `docs/deployments/`:
> 1. [Cloudflare Containers & Pages](file:///Users/oscargarcia/workspace/graphfolio/docs/deployments/cloudflare-containers-deployment-plan.md)
> 2. [Google Cloud Platform (Cloud Run & Cloud SQL)](file:///Users/oscargarcia/workspace/graphfolio/docs/deployments/gcp-deployment-plan.md)
> 3. [Microsoft Azure (Container Apps & Flexible Server)](file:///Users/oscargarcia/workspace/graphfolio/docs/deployments/azure-deployment-plan.md)
> 4. [Amazon Web Services (ECS Fargate & RDS Multi-AZ)](file:///Users/oscargarcia/workspace/graphfolio/docs/deployments/aws-deployment-plan.md)
> 
> **Evaluation Date**: October 2026  
> **Currency**: USD ($)  
> **Key Principle**: Realistic, transparent line-item pricing modeling idle minimums, network egress, multi-AZ resilience, and serverless scale-to-zero capabilities.

---

## 1. Executive Cost Summary

The four cloud architectures represent distinct operational and billing paradigms:
- **Cloudflare**: Maximizes scale-to-zero edge economics by colocating the compiled Go backend into a single container pod managed by Workers and Durable Objects, offloading frontend assets to Cloudflare Pages (free), and relying on managed external PostgreSQL with Cloudflare Hyperdrive.
- **Google Cloud Platform (GCP)**: A balanced serverless microservice model using Cloud Run (per-request billing, sub-second cold starts, direct VPC egress) paired with an enterprise-grade private Cloud SQL PostgreSQL HA database.
- **Microsoft Azure**: A serverless Kubernetes abstraction using Azure Container Apps (native gRPC/HTTP2 Envoy mesh, scale-to-zero or idle pricing) paired with Azure Database for PostgreSQL Flexible Server with Zone-Redundant High Availability.
- **Amazon Web Services (AWS)**: A traditional multi-tier enterprise VPC deployment using ECS on Fargate, dual Application Load Balancers (Public + Internal gRPC), NAT Gateways, and Amazon RDS PostgreSQL Multi-AZ.

### Monthly Cost Comparison by Workload Tier

| Workload Tier | Monthly Traffic & Workload Profile | Cloudflare (Plan 1) | GCP (Plan 2) | Azure (Plan 3) | AWS (Plan 4) |
| :--- | :--- | :---: | :---: | :---: | :---: |
| **Tier 1: Starter / MVP** | ~1,000 MAU, 50k reqs/day, dev/burstable DB, scale-to-zero compute | **$24 – $35 / mo** | **$45 – $75 / mo** | **$35 – $65 / mo** | **$165 – $220 / mo** |
| **Tier 2: Production Baseline** | ~25,000 MAU, 500k reqs/day, warm instances, Multi-AZ High Availability DB | **$45 – $85 / mo** | **$235 – $295 / mo** | **$320 – $410 / mo** | **$475 – $580 / mo** |
| **Tier 3: Growth / High Traffic** | ~100,000+ MAU, 3M reqs/day, horizontal auto-scaling, dedicated high-IOPS DB | **$110 – $220 / mo** | **$420 – $650 / mo** | **$580 – $850 / mo** | **$780 – $1,150 / mo** |

### Annual TCO (Production Baseline Tier 2)

```
Cloudflare : █████ ($540 – $1,020 / year)
GCP        : ████████████████████ ($2,820 – $3,540 / year)
Azure      : ███████████████████████████ ($3,840 – $4,920 / year)
AWS        : ████████████████████████████████████████ ($5,700 – $6,960 / year)
```

---

## 2. Plan 1: Cloudflare Deployment Cost Deep-Dive

Deployment plan reference: [cloudflare-containers-deployment-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/deployments/cloudflare-containers-deployment-plan.md)

### Architectural Overview
- **Compute**: Single unified container pod (BFF :8080, Portfolio API :50051, User API :50052) running inside a Cloudflare Durable Object.
- **Frontend**: Cloudflare Pages (`graphfolio.app` and `admin.graphfolio.app`).
- **Database**: Managed Serverless PostgreSQL (e.g. Neon Pro or Supabase Pro) accelerated by Cloudflare Hyperdrive.
- **Networking**: Cloudflare Global Anycast Edge with integrated DDoS/WAF.

### Itemized Monthly Cost Table

| Service Component | Tier 1 (Starter) | Tier 2 (Production) | Tier 3 (Growth) | Billing Metric & Pricing Formula |
| :--- | :---: | :---: | :---: | :--- |
| **Cloudflare Workers Paid Base** | $5.00 | $5.00 | $5.00 | Flat subscription includes 10M requests & 30M CPU-ms. |
| **Cloudflare Worker Additional Requests** | $0.00 | $1.50 | $7.50 | $0.30 per 1M requests above included 10M. |
| **Cloudflare Containers Compute**<br>(Durable Objects Active Memory & CPU) | $2.50 | $15.00 | $45.00 | $12.50 / GB-month of active RAM. Single 128MB container pod active on request & market close. |
| **Cloudflare Pages (SPAs)** | $0.00 | $0.00 | $0.00 | Free tier includes unlimited requests, bandwidth, and edge caching. |
| **Cloudflare Hyperdrive** | $0.00 | $0.00 | $0.00 | Included free with Workers Paid plan. Connection pooling & query acceleration. |
| **Managed PostgreSQL**<br>(Neon / Supabase) | $19.00 | $29.00 | $85.00 | Neon Launch ($19/mo) / Supabase Pro ($25/mo). Scales based on compute units and storage. |
| **Cloudflare Cron Triggers** | $0.00 | $0.00 | $0.00 | Free with Workers. Triggers nightly EOD valuation at 22:00 UTC. |
| **Data Egress (Bandwidth)** | $0.00 | $0.00 | $0.00 | Cloudflare provides **$0 egress fees** across all edge cache and API routes. |
| **Cloudflare Zero Trust (Admin Access)** | $0.00 | $0.00 | $0.00 | Free tier covers up to 50 active seats (sufficient for internal ops team). |
| **Total Estimated Monthly Cost** | **$26.50** | **$50.50** | **$142.50** | |

---

## 3. Plan 2: Google Cloud Platform (GCP) Cost Deep-Dive

Deployment plan reference: [gcp-deployment-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/deployments/gcp-deployment-plan.md)

### Architectural Overview
- **Compute**: Independent Cloud Run services (`bff`, `portfolio-api`, `user-api`) with native gRPC HTTP/2.
- **Valuation Job**: Cloud Run Job triggered by Cloud Scheduler.
- **Frontend**: Google Cloud Storage (GCS) buckets behind Cloud CDN.
- **Database**: Cloud SQL for PostgreSQL 15+ in a private VPC with Regional High Availability.
- **Ingress & Security**: Global External Application Load Balancer with Cloud Armor WAF and Identity-Aware Proxy (IAP).

### Itemized Monthly Cost Table

| Service Component | Tier 1 (Starter) | Tier 2 (Production) | Tier 3 (Growth) | Billing Metric & Pricing Formula |
| :--- | :---: | :---: | :---: | :--- |
| **Cloud Run (BFF, Portfolio, User API)** | $5.00 | $45.00 | $120.00 | Tier 1: Scale-to-zero. Tier 2: `min-instances: 1` on BFF/Portfolio ($0.000024/vCPU-s, $0.0000025/GB-s). |
| **Cloud Run Job (Valuation Worker)** | $0.15 | $0.35 | $1.20 | Runs ~5–10 mins nightly: 30 runs * 600s * (1 vCPU, 2GB RAM). Minimal cost. |
| **Cloud SQL for PostgreSQL 15+** | $25.00 | $175.00 | $285.00 | Tier 1: `db-f1-micro` ($15) or `db-g1-small` ($25). Tier 2: `db-custom-2-7680` ($85) + Regional HA ($85) + 100GB SSD ($17). |
| **Global External HTTPS Load Balancer** | $18.25 | $21.50 | $35.00 | $0.025/hour for first 5 forwarding rules (~$18.25/mo) + $0.008/GB processed. |
| **Cloud CDN & Cloud Storage** | $1.50 | $4.50 | $15.00 | GCS storage ($0.020/GB) + Cloud CDN cache fill and egress ($0.08/GB). |
| **Cloud Armor WAF** | $5.00 | $11.00 | $22.00 | $5.00/policy/mo + $1.00/rule/mo + $0.75 per 1M requests. |
| **Secret Manager & Cloud Scheduler** | $0.50 | $1.20 | $2.50 | $0.06 per 10k secret accesses + 3 free scheduler jobs. |
| **Cloud Logging & Trace Monitoring** | $0.00 | $5.00 | $18.00 | First 50 GiB/month free; $0.50/GiB thereafter. |
| **Network Egress (Google Cloud Egress)** | $1.50 | $6.50 | $25.00 | $0.085–$0.12 per GB for internet egress. |
| **Total Estimated Monthly Cost** | **$56.90** | **$270.05** | **$523.70** | |

---

## 4. Plan 3: Microsoft Azure Cost Deep-Dive

Deployment plan reference: [azure-deployment-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/deployments/azure-deployment-plan.md)

### Architectural Overview
- **Compute**: Azure Container Apps (ACA) with internal Envoy gRPC HTTP/2 mesh and scale-to-zero capabilities.
- **Valuation Job**: Azure Container Apps Job with built-in cron scheduler (`0 22 * * 1-5`).
- **Frontend**: Azure Static Web Apps (Standard Tier) with custom domains and SSL.
- **Database**: Azure Database for PostgreSQL - Flexible Server with Zone-Redundant High Availability.
- **Edge & Security**: Azure Front Door (Standard) with WAF rate limiting and Microsoft Entra ID integration.

### Itemized Monthly Cost Table

| Service Component | Tier 1 (Starter) | Tier 2 (Production) | Tier 3 (Growth) | Billing Metric & Pricing Formula |
| :--- | :---: | :---: | :---: | :--- |
| **Azure Container Apps (ACA)**<br>(BFF, Portfolio, User API) | $5.00 | $38.00 | $95.00 | Free grant: 180k vCPU-s, 360k GiB-s, 2M reqs/mo.<br>Tier 2: Idle allocation ($0.0000035/vCPU-s) for min-replicas: 1 + active requests. |
| **Container Apps Job (Valuation Worker)** | $0.15 | $0.40 | $1.20 | Scheduled batch task: ~10 mins nightly * (1 vCPU, 2 GiB). |
| **Azure PostgreSQL Flexible Server** | $28.00 | $255.00 | $380.00 | Tier 1: Burstable `B1ms`/`B2s` ($28). Tier 2: `Standard_D2ds_v5` ($115) + Zone-Redundant HA ($115) + 128GB SSD ($25). |
| **Azure Front Door (Standard) & WAF** | $0.00 | $53.00 | $78.00 | Tier 1: Direct ACA Ingress ($0). Tier 2: Base profile ($35/mo) + WAF Policy ($18/mo) + data routing ($0.01/GB). |
| **Azure Static Web Apps (SWA)** | $0.00 | $9.00 | $18.00 | Tier 1: Free plan ($0). Tier 2: Standard tier ($9/mo per app) for custom domains and SLA. |
| **Azure Container Registry (ACR)** | $5.00 | $5.00 | $20.00 | Basic tier: $0.167/day (~$5.00/mo includes 10GB storage). |
| **Azure Key Vault & Virtual Network** | $1.00 | $3.50 | $8.00 | Key Vault operations ($0.03 per 10k) + Private DNS zones for PostgreSQL VNet integration. |
| **Azure Monitor & Log Analytics** | $0.00 | $6.00 | $22.00 | First 5 GB/month free; $2.30/GB ingested thereafter. |
| **Outbound Data Egress** | $1.00 | $6.50 | $26.00 | First 100 GB/month free; $0.087/GB thereafter. |
| **Total Estimated Monthly Cost** | **$40.15** | **$376.40** | **$648.20** | |

### Key Economic Takeaways
1. **Envoy Mesh Eliminates Load Balancer Overhead**: Azure Container Apps handles internal gRPC routing natively without requiring a separate Application Gateway, saving ~$35–$45/mo compared to AWS.
2. **PostgreSQL Flexible Server HA**: Zone-redundant high availability doubles compute costs ($115 $\rightarrow$ $230/mo) for synchronous multi-zone replication.
3. **Scale-to-Zero Advantage**: ACA's idle pricing mode drops vCPU rates by over 80% when standby replicas are idle, making warm standby instances more cost-effective than AWS Fargate.

---

## 5. Plan 4: Amazon Web Services (AWS) Cost Deep-Dive

Deployment plan reference: [aws-deployment-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/deployments/aws-deployment-plan.md)

### Architectural Overview
- **Compute**: AWS ECS on Fargate (Multi-AZ tasks for BFF, Portfolio API, and User API).
- **Valuation Job**: Scheduled ECS Fargate `RunTask` via Amazon EventBridge.
- **Networking**: 2 Application Load Balancers (Public internet-facing ALB + Internal ALB for gRPC) + 2 Multi-AZ NAT Gateways.
- **Database**: Amazon RDS for PostgreSQL 15+ Multi-AZ.
- **Frontend**: Amazon S3 private buckets + Amazon CloudFront with Origin Access Control (OAC).

### Itemized Monthly Cost Table

| Service Component | Tier 1 (Starter) | Tier 2 (Production) | Tier 3 (Growth) | Billing Metric & Pricing Formula |
| :--- | :---: | :---: | :---: | :--- |
| **ECS Fargate Compute**<br>(BFF, Portfolio, User API) | $45.00 | $144.00 | $288.00 | $0.04048/vCPU-hr, $0.004445/GB-hr.<br>Tier 2: 2 tasks each of BFF (0.5v/1GB), PortSvc (1v/2GB), UserSvc (0.5v/1GB) running 24/7. |
| **EventBridge + Fargate RunTask**<br>(Valuation Worker) | $0.20 | $0.45 | $1.50 | Daily one-shot task: ~15 mins * (1 vCPU, 2GB) = negligible. |
| **Amazon RDS PostgreSQL 15+** | $45.00 | $210.00 | $350.00 | Tier 1: `db.t4g.medium` Single-AZ ($45). Tier 2: `db.m6g.large` Multi-AZ ($185) + 100GB gp3 storage ($25). |
| **Application Load Balancers (ALBs)**<br>(Public ALB + Internal gRPC ALB) | $42.00 | $48.00 | $72.00 | $0.0225/hr per ALB ($16.42/mo base * 2 ALBs = $32.84) + LCU processing metrics. |
| **NAT Gateways (Multi-AZ Egress)** | $35.00 | $70.00 | $85.00 | $0.045/hr per NAT Gateway (~$32.85/mo each). Multi-AZ requires 2 NAT GWs ($65.70 base) + $0.045/GB data processed. |
| **Amazon CloudFront & S3 Storage** | $2.50 | $6.50 | $22.00 | S3 storage ($0.023/GB) + CloudFront data transfer ($0.085/GB) + HTTPS requests. |
| **AWS WAF v2 (CloudFront & ALB)** | $6.00 | $16.00 | $28.00 | $5.00/WebACL/mo + $1.00/rule/mo + $0.60 per 1M requests inspected. |
| **AWS Secrets Manager & KMS** | $2.00 | $4.00 | $8.00 | $0.40/secret/mo (4 secrets = $1.60) + KMS key usage. |
| **Amazon CloudWatch & Container Insights** | $3.00 | $12.00 | $35.00 | Metrics, alarms, and Container Insights log ingestion. |
| **AWS Data Egress to Internet** | $2.00 | $8.50 | $30.00 | First 100GB/mo free; $0.09/GB thereafter. |
| **Total Estimated Monthly Cost** | **$182.70** | **$519.45** | **$919.50** | |

---

## 6. Architectural Cost Drivers Comparison

| Cost Driver | Cloudflare (Plan 1) | GCP (Plan 2) | Azure (Plan 3) | AWS (Plan 4) | Impact Summary |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Idle Infrastructure Base (Minimum to keep alive)** | **~$24/mo**<br>(Workers $5 + Neon $19) | **~$45/mo**<br>(Cloud SQL small + Load Balancer) | **~$40/mo**<br>(Burstable Postgres + ACR) | **~$150/mo**<br>(2 ALBs + 1 NAT GW + Fargate min + RDS) | AWS has a high fixed cost floor; Cloudflare is practically pay-per-use. |
| **Inter-Service gRPC Routing** | **$0.00**<br>(Loopback `127.0.0.1` inside container pod) | **$0.00**<br>(Internal Cloud Run URL with IAM) | **$0.00**<br>(Envoy internal HTTP/2 ingress) | **~$25/mo**<br>(Dedicated Internal ALB for gRPC) | Cloudflare, GCP, and Azure handle inter-service RPC without dedicated load balancer hardware. |
| **Internet Egress / NAT Gateways** | **$0.00**<br>(Free egress on Cloudflare) | **$0.00**<br>(Direct VPC Egress has built-in Google egress without NAT fee) | **$0.00**<br>(ACA VNet outbound internet is included without NAT Gateway) | **$65–$80/mo**<br>(NAT Gateways charged per hour + per GB) | NAT Gateway hourly fees on AWS represent a significant hidden expense. |
| **Database High Availability (HA)** | **Included in SaaS tier**<br>(Neon / Supabase manage replicas) | **~$85/mo add-on**<br>(Cloud SQL standby instance) | **~$115/mo add-on**<br>(Flexible Server zone-redundant standby) | **~$95/mo add-on**<br>(RDS Multi-AZ synchronous replica) | Managed serverless DBs bundle HA replication into their base platform tiers. |
| **DDoS & Web Application Firewall** | **Included**<br>(Standard Cloudflare edge rules) | **~$11/mo**<br>(Cloud Armor policy + rules) | **~$18/mo**<br>(Front Door WAF policy) | **~$16/mo**<br>(AWS WAF WebACL + managed rules) | Cloudflare provides enterprise-grade WAF and DDoS mitigation as a core platform feature. |

---

## 7. One-Year Total Cost of Ownership (TCO) Projections

Assuming a standard trajectory:
- **Months 1–3**: Development & Soft Launch (Tier 1 Starter)
- **Months 4–9**: Production Baseline (Tier 2)
- **Months 10–12**: Growth & Scaling (Tier 2/3 transition)

### 12-Month Cumulative Cost Projection

| Month Range | Stage | Cloudflare (Plan 1) | GCP (Plan 2) | Azure (Plan 3) | AWS (Plan 4) |
| :--- | :--- | :---: | :---: | :---: | :---: |
| **Months 1–3 (3 mos)** | Dev / MVP Launch | $80 | $170 | $120 | $550 |
| **Months 4–9 (6 mos)** | Production Baseline | $300 | $1,620 | $2,250 | $3,120 |
| **Months 10–12 (3 mos)**| Growth Phase | $250 | $1,050 | $1,580 | $2,050 |
| **Total 1-Year Cloud Spend** | **12-Month Total** | **$630** | **$2,840** | **$3,950** | **$5,720** |
| **Cost Ratio vs Cloudflare** | — | **1.0x (Baseline)** | **4.5x** | **6.3x** | **9.1x** |

---

## 8. Final Decision & Recommendation Matrix

| Decision Criteria | Winner | Why |
| :--- | :---: | :--- |
| **Lowest Operational Cost** | **Cloudflare** | At **~$50/month** for full production, Cloudflare is ~80% cheaper than GCP, ~87% cheaper than Azure, and ~90% cheaper than AWS, with zero bandwidth fees. |
| **Fastest Global Performance** | **Cloudflare** | Code and static assets run across 330+ edge locations with native Anycast routing and sub-millisecond cold starts. |
| **Best Serverless Hyperscaler** | **GCP** | Cloud Run provides native gRPC, standard Docker deployment, independent microservice scaling, and Google-managed database isolation without AWS/Azure database pricing premiums. |
| **Best Microsoft Enterprise Integration** | **Azure** | Seamless integration with Microsoft Entra ID (Azure AD), Visual Studio, GitHub OIDC, and Azure Front Door for Microsoft-centric organizations. |
| **Deepest Corporate Ecosystem** | **AWS** | Preferred if the organization already has an AWS Enterprise Discount Program (EDP), strict compliance certifications, or existing VPC infrastructure. |

### Recommended Action Plan
- **Primary Recommendation**: **Deploy Plan 1 (Cloudflare Containers + Pages)** for minimum overhead ($50/mo), zero network complexity, and maximum global edge speed.
- **Enterprise Hyperscaler Recommendation**:
  - If your organization is standardizing on **Google Cloud**: Choose **[GCP](file:///Users/oscargarcia/workspace/graphfolio/docs/deployments/gcp-deployment-plan.md)** ($270/mo) for its superior serverless ergonomics and cost efficiency.
  - If your organization operates within the **Microsoft ecosystem**: Choose **[Azure](file:///Users/oscargarcia/workspace/graphfolio/docs/deployments/azure-deployment-plan.md)** ($376/mo) for native Container Apps gRPC, Entra ID SSO, and Azure Front Door.
