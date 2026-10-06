# User Story: Automated Market Data Ingestion Pipeline

> **Target**: Market Data Pipeline (`services/portfolio-api/internal/marketdata/`, `cmd/market-ingest/`)  
> **Status**: Ready for Implementation  
> **Implementation Plan**: [market-data-ingestion-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/market-data-ingestion-implementation-plan.md)  
> **Roadmap Reference**: [FEATURES.md (Feature #14)](file:///Users/oscargarcia/workspace/graphfolio/FEATURES.md#L24)

---

## 1. User Story

> **As a** system administrator and automated operations workflow,  
> **I want** daily EOD stock closing prices and central bank FX rates automatically ingested and triangulated from resilient external market feeds,  
> **So that** holding valuations, performance charts, and currency conversions remain fresh and accurate without manual intervention.

---

## 2. Acceptance Criteria

- [ ] **Multi-Provider Architecture**: Ingest daily closing prices through Twelve Data with automated fallback to Yahoo Finance on provider errors or rate limits.
- [ ] **ECB FX Ingestion**: Ingest official European Central Bank reference fixing XML feeds with exact 10-decimal triangulation across active currencies.
- [ ] **Rate Limiting & Retries**: Govern outbound network requests using a token-bucket rate limiter with exponential backoff and jitter on HTTP 429 and 5xx responses.
- [ ] **Automatic Trade Ingestion Backfill**: Detect unpriced date ranges when new transactions are entered and automatically trigger background backfill prior to projection rebuilds.
- [ ] **Scheduled Worker Execution**: Provide a CLI worker (`cmd/market-ingest/main.go`) runnable via `make ingest-market-data` for cron scheduling.

---

## 3. Related Documentation

- **Architecture & Implementation Plan**: [market-data-ingestion-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/market-data-ingestion-implementation-plan.md)
- **Feature Matrix & Roadmap**: [FEATURES.md](file:///Users/oscargarcia/workspace/graphfolio/FEATURES.md)
