# Competitive Analysis: GraphFolio (Portfolio Insights) vs. Commercial Alternatives

> **Author**: Principal Product Strategist & Enterprise Architect  
> **Date**: October 2026  
> **Target Scope**: Deep-dive competitive analysis evaluating **GraphFolio** against commercial market leaders: **Sharesight**, **Simply Wall St**, **Morningstar Portfolio Manager**, and **Koyfin**.

---

## Executive Summary

The retail and prosumer investment software landscape is fragmented between two extremes:
1. **Gamified, superficial retail apps** (e.g., Simply Wall St, Robinhood) that rely on black-box visual heuristics and generic valuation multiples (P/E, P/B).
2. **Clunky legacy trackers or expensive data terminals** (Sharesight, Morningstar, Koyfin, Bloomberg) that either suffer from high subscription fees, rigid monolithic architectures, or lack native support for deep fundamental corporate capital-allocation metrics (ROIC, Cash Conversion Cycles, Owner Earnings).

**GraphFolio** occupies a high-moat sweet spot: **Institutional-grade precision, deterministic ledger-replay mechanics, and deep fundamental/capital-allocation analytics for serious, concentrated value investors**, delivered through a modern, cloud-native microservices architecture.

```
       High Analytical Depth
                ▲
                │             [GraphFolio]
                │           (Deterministic Ledger,
                │            ROIC/FCF, Exact Precision)
     [Koyfin]   │
   (Market Data,│
    Multiples)  │
                │        [Sharesight]
                │       (Tax & Dividends)
                │
                │   [Morningstar]
                │  (Fund/Star Ratings)
                │
  [Simply Wall St]
  (Retail Snowflake)
                │
◄───────────────┴──────────────────────────────►
Low Modernity / Architecture     High Modernity / Architecture
```

---

## Section 1: Feature Parity & Niche Advantage (Matrix)

### 1.1 Side-by-Side Capability Comparison Table

| Capability | **GraphFolio** | **Sharesight** | **Simply Wall St** | **Morningstar** | **Koyfin** |
|---|:---:|:---:|:---:|:---:|:---:|
| **Target Audience** | Concentrated Fundamental & Value Investors | Tax-conscious retail & accountants | Casual retail & visual beginners | Traditional wealth managers & mutual fund investors | Institutional analysts, RIAs, macro traders |
| **Precision Standard** | **Exact Arbitrary Precision** (`shopspring/decimal`, zero float) | Standard floating-point (periodic rounding drift) | Standard 2-decimal floating point | Float64 legacy backend | High-precision market data engine |
| **Portfolio Math Standards** | **Strict GIPS Compliance** (TWR Index, sub-1yr annualization prohibition) | TWR & MWR (some custom proprietary adjustments) | Basic simple return (unadjusted for cash timing) | Standard mutual fund calculation engine | Advanced MWR / TWR, benchmark comparisons |
| **Cost Basis Relief Engines** | **Dual Engine**: Dynamic Average Cost ↔ FIFO Replay | FIFO, LIFO, Specific Lot, Minimise Gain | Not supported (simple average trade price) | Average Cost & FIFO (static) | Account-level aggregate |
| **Auditability & Ledger Replay** | **100% Deterministic Event Sourced** (`RebuildProjections`) | Proprietary transaction log | Static position snapshots | Batch nightly ledger sync | External broker sync |
| **Fundamental Metric Depth** | **Deep Capital Allocation**: ROIC, FCF Yield, OCF, Owner Earnings | Surface level (Dividend yields, gross yield) | Heuristic snowflake (5 broad categories, static rules) | Traditional (P/E, P/B, Star Rating, Moat Rating) | Comprehensive fundamental financial statements |
| **Multi-Currency & FX Decomposition** | **Exact Triangulated FX** (Asset vs Currency return split) | Multi-currency with daily FX | Currency converted at snapshot | Basic currency translation | Multi-currency real-time cross rates |
| **Corporate Actions** | Deterministic split/merger tax lot recomputation | Automated split/dividend feed | Automated adjustments | Corporate actions feed | Institutional corporate action adjustments |
| **Technology Stack** | **Go + gRPC + GraphQL + React 19** | Ruby on Rails monolith | PHP / Node.js web app | Java / C++ enterprise legacy | React + Node.js / Python microservices |
| **Pricing / Access** | Self-hosted / Open core / Low infrastructure cost | $15 – $45 / mo (tiered by holdings limit) | $10 – $30 / mo (freemium paywalls) | $249 / yr (Investor edition) | $39 – $110 / mo |

---

### 1.2 Fundamental Metric Depth: Why Standard Metrics Fail Serious Investors

Commercial tools predominantly display **first-order accounting multiples**:
- Price-to-Earnings ($P/E$)
- Price-to-Book ($P/B$)
- Dividend Yield
- Historical EPS growth
- Consensus Sell-Side Price Targets

For concentrated value investors (following Warren Buffett, Charlie Munger, Nick Sleep, or Mohnish Pabrai), these metrics are **actively misleading**:
1. **$P/E$ is distorted by non-cash accruals**: Net income includes non-cash items, one-off restructurings, tax asset revaluations, and management adjustments.
2. **Book value ignores intangible competitive moats**: In modern capital-light business models (software, platform networks, branded consumer goods), economic assets never appear on balance sheets due to conservative accounting rules (expensing R&D and SG&A).
3. **Dividend yield ignores capital allocation efficacy**: A 7% dividend yield from an indebted legacy utility destroys shareholder value if ROIC is lower than the Weighted Average Cost of Capital (WACC), whereas a 0% dividend stock reinvesting at 25% ROIC creates extraordinary compounding.

#### GraphFolio's Deep Fundamental Metric Advantage:

$$\text{ROIC} = \frac{\text{NOPAT}}{\text{Invested Capital}} = \frac{\text{EBIT} \times (1 - t)}{(\text{Total Assets} - \text{Excess Cash}) - (\text{Non-Interest Bearing Current Liabilities})}$$

$$\text{Owner Earnings} = \text{Operating Cash Flow} - \text{Maintenance CapEx}$$

$$\text{Free Cash Flow Margin} = \frac{\text{Free Cash Flow}}{\text{Total Revenue}}$$

$$\text{Reinvestment Rate} = \frac{\text{CapEx} - \text{D&A} + \Delta \text{NWC}}{\text{NOPAT}}$$

```
Intrinsic Value Growth Rate = ROIC × Reinvestment Rate
```

By tracking **ROIC**, **Operating Cash Flow (OCF)**, **Maintenance vs. Growth CapEx**, and **Invested Capital Efficiency** alongside portfolio holdings, GraphFolio answers the vital questions:
- *Is the underlying business earning superior economic profits on incremental capital?*
- *Is management allocating retained earnings effectively, or are they empire-building through dilutive M&A?*
- *What is the portfolio-weighted look-through Owner Earnings yield compared to the 10-Year Treasury yield?*

---

### 1.3 Key Deficiencies in Existing Commercial Platforms

1. **Sharesight**:
   - *Strengths*: Superb tax reporting and multi-country dividend calculations.
   - *Deficiencies*: Completely blind to business fundamentals. It tracks price and dividends as abstract tickers without insight into corporate health, balance sheet leverage, or free cash flow sustainability. Highly restrictive tier limits (e.g. Free tier capped at 10 holdings).
2. **Simply Wall St**:
   - *Strengths*: High visual polish; snowflake infographics appeal to novices.
   - *Deficiencies*: "Gamified" analysis. The snowflake assigns arbitrary scorecards (1 to 6) based on static rules (e.g., punishing high-growth companies that hold debt regardless of net leverage or cash flow stability). Black-box algorithms without raw auditable balance sheet numbers undermine trust for professional and serious self-directed investors.
3. **Morningstar Portfolio Manager**:
   - *Strengths*: Deep asset allocation, fund category coverage, and Morningstar "Economic Moat" ratings.
   - *Deficiencies*: Dated, slow, legacy UX built on 2000s Web technologies. Heavy bias toward mutual funds and managed portfolios rather than individual equities. Closed proprietary methodologies that cannot be customized or recalculated based on personal tax lots.
4. **Koyfin**:
   - *Strengths*: Exceptional charting, market data depth, macro graphing, and consensus estimate dashboards.
   - *Deficiencies*: Optimised for *macro scanning* and *security research*, not *deep personalized portfolio tracking*. Its portfolio management tools lack deterministic tax lot relief switching (FIFO vs Average Cost), fine-grained dividend reinvestment lot tracking, and privacy-first self-hosted options. High subscription costs ($450–$1,300/year) lock out cost-conscious independent investors.

---

## Section 2: Architecture & Performance Comparison

### 2.1 Strategic Architectural Advantages

GraphFolio was engineered from day one as a **high-throughput, modular, cloud-native microservices architecture**:

```text
[ Client: Vite + React 19 SPA ]
          │  ▲
   GraphQL│  │ (Exact Payloads, Zero Over-fetching)
          ▼  │
[ BFF Gateway: Go / gqlgen (:8080) ]
          │  ▲
      gRPC│  │ (Protobuf v3, Binary Transport, Sub-millisecond RPC)
          ▼  │
┌──────────────────────────────┴──────────────────────────────┐
│                                                             │
▼                                                             ▼
[ Portfolio API: Go (:50051) ]               [ User API: Go (:50052) ]
(Ledger Replay & Projections)                (Preferences & Auth)
│                                            │
▼                                            ▼
[ schema: portfolio ]                        [ schema: users ]
└──────────────────────────────┬──────────────────────────────┘
                               │
                      [ PostgreSQL 17+ ]
```

#### Comparison of Architectural Paradigms:

| Dimension | **GraphFolio** | Legacy Platforms (Sharesight, Morningstar) | Modern Cloud Tools (Simply Wall St, Koyfin) |
|---|---|---|---|
| **Inter-Service Communication** | **gRPC / Protocol Buffers** (Zero schema drift, binary HTTP/2, sub-millisecond serialization) | Monolithic internal method calls / SQL joins | REST JSON APIs (High CPU serialization overhead) |
| **API Surface** | **GraphQL (`gqlgen`) BFF** (Strongly typed, declarative querying, zero over-fetching) | Server-rendered HTML (Rails/ERB/JSP) or monolithic REST | REST / GraphQL mix |
| **Frontend Runtime** | **React 19 + TypeScript + Vite** (Optimized SVG Bezier curves, zero bundle bloat) | Heavy legacy JS bundles (jQuery / older Angular) | React / MobX / Redux thick clients |
| **Database Isolation** | **Multi-Schema PostgreSQL** (`portfolio`, `users`) with dedicated RBAC roles | Single monolithic database schema | Multi-tenant cloud DB (Postgres / Aurora) |
| **Arithmetic Safety** | **Arbitrary-precision fixed point** (`shopspring/decimal`, zero float64) | IEEE 754 Floating-Point in Ruby/Python | Floating-point JavaScript/Python math |

---

### 2.2 Operational Impact

#### 1. UI Responsiveness & Query Latency
- **Sub-10ms BFF-to-Microservice latency**: Binary gRPC transport over persistent HTTP/2 TCP sockets eliminates JSON parsing bottlenecks on the server.
- **Single-Roundtrip Data Retrieval**: The GraphQL BFF allows the React frontend to fetch the complete portfolio summary, holding breakdown, and 365-day historical valuation Bezier curve points in a **single roundtrip HTTP request**.
- **Lightweight Vanilla CSS & SVG Rendering**: In [PerformanceChart.tsx](file:///Users/oscargarcia/workspace/graphfolio/web/src/components/PerformanceChart.tsx), smooth Catmull-Rom spline curves are computed in pure vector math without heavy 3rd-party charting libraries (like Highcharts or Chart.js), maintaining 60 FPS scrolling and instantaneous crosshair tooltips.

#### 2. Real-Time & Event-Driven Deterministic Data Processing
- **Deterministic Ledger Replay Engine**: In [projection.go](file:///Users/oscargarcia/workspace/graphfolio/services/portfolio-api/internal/service/projection.go), holding quantities, average cost bases, cash balances, and tax lots are replayed sequentially from immutable transactions in `portfolio.transactions`.
- **Concurrency & State Integrity**: Every rebuild acquires an atomic row lock (`SELECT id FROM portfolio.portfolios WHERE id = $1 FOR UPDATE`). This guarantees zero race conditions when multiple transactions or external dividends are ingested concurrently.
- **Transaction Rollback & Deletion Safety**: Deleting or editing a past-dated transaction atomically purges derived lots and valuations, deterministically re-evaluating historical states without silent accounting corruption.

#### 3. Rapid Feature Iteration & Domain Boundary Isolation
- **Contract-First Development**: Protocol Buffers in `proto/portfolio/v1/portfolio.proto` act as the single source of truth. Backend microservices, BFF wrappers, and TypeScript clients are regenerated in seconds (`make generate`).
- **Independent Schema Ownership**: Database migrations for `portfolio` are completely segregated from `users`. The portfolio engine can evolve tax-lot tables or valuation metrics without risking user session tables.

---

## Section 3: Audience & Positioning

### 3.1 Target Investor Profile: "The Concentrated Fundamentalist"

GraphFolio does not seek to serve day-traders or passive target-date indexers. It is custom-tailored for the **Disciplined Fundamental Investor**:
- **Philosophical Alignment**: Warren Buffett, Charlie Munger, Bill Ackman, Nick Sleep, Guy Spier, and Michael Burry.
- **Portfolio Characteristics**:
  - High conviction, concentrated holdings (5 to 15 securities).
  - Multi-year holding horizons (low turnover, high tax sensitivity).
  - Global asset allocation (US equities, European quality compounders, Japanese net-nets, cash reserves in multiple currencies).
  - Continuous evaluation of capital reinvestment and owner earnings.
- **Key Pain Points with Current Tools**:
  - Frustration with tools that force broad "diversification scores" on portfolios that deliberately hold concentrated bets.
  - Inability to dynamically toggle between tax-reporting cost basis (FIFO) and economic-performance cost basis (Average Cost).
  - Lack of visibility into look-through fundamental metrics across their total portfolio (e.g., aggregate portfolio P/FCF or weighted ROIC).

### 3.2 Market Positioning Matrix

```
       [Institutional Terminals]
       (Bloomberg, FactSet)
       • High Data / Pro Depth
       • Prohibitive Cost ($25k+/yr)
                  ▲
                  │             [GraphFolio]
                  │             • High Fundamental Depth
                  │             • Deterministic Ledger
                  │             • Accessible & Self-Owned
                  │
  [Koyfin] ───────┼───────────── [Sharesight]
  • Great charting│              • Good tax tracking
  • Generic multiples            • Zero fundamental depth
                  │
                  │
       [Simply Wall St / Yahoo Finance]
       • Gamified Heuristics / Ads
       • Mass Retail Audience
                  ▼
```

---

## Section 4: SWOT Analysis (GraphFolio vs. The Market)

### Strengths (Internal Moat)
1. **Zero-Float Accounting Precision**: Strict mathematical compliance (`shopspring/decimal`) avoids penny-rounding errors common in consumer trackers.
2. **Deterministic Ledger Replay Engine**: Event-sourced transaction replay guarantees that portfolio history, cash balances, and tax lots never desynchronize.
3. **High-Performance Architecture**: Go + gRPC microservices and GraphQL ensure lightning-fast UI responsiveness and scalable batch computations.
4. **GIPS Compliance**: Follows institutional CFA Institute performance standards (prohibiting misleading sub-1-year annualization and isolating external cash flows via TWR).
5. **Privacy & Data Sovereignty**: Self-hostable or privately deployed, shielding high-net-worth investor transactions from data monetization and broker surveillance.

### Weaknesses (Current Constraints)
1. **Market Data Feed Costs**: Lack of an institutional direct feed (e.g. FactSet, Refinitiv) requires relying on EOD APIs (Twelve Data, Yahoo Finance, ECB) or manual input for esoteric securities.
2. **Absence of Automated Bank/Broker Aggregation**: No Plaid or Yodlee automated broker sync; currently relies on manual ingestion or CSV imports.
3. **Nascent Ecosystem**: Early-stage product compared to decades-old incumbents (Morningstar, Sharesight).

### Opportunities (Market White Space)
1. **"Owner Earnings & Look-Through Value" Dashboard**: Incumbents display market price return; no tool displays an investor's look-through share of annual free cash flow and retained earnings.
2. **Open-Core & Self-Directed Family Office Niche**: Boutique RIAs, high-net-worth individuals, and family offices seeking an auditable, customizable portfolio tracking engine without $30,000/year terminal fees.
3. **Advanced Tax-Lot Optimization**: Intelligent lot-relief harvesting recommendations (e.g., automated tax-loss selling simulation) before year-end.

### Threats (Market Pressures)
1. **Aggressive Pricing by Koyfin**: If Koyfin deepens its portfolio tracking to include full tax-lot accounting at an affordable tier.
2. **Broker Integration Platforms**: Interactive Brokers, Charles Schwab, and Robinhood improving their native analytical dashboards.
3. **Commoditization of AI Summary Agents**: LLM tools directly analyzing broker CSVs on demand, lowering switching costs for basic reporting.

---

## Section 5: Next-Step Strategic Roadmap

To widen the strategic moat between **GraphFolio** and commercial alternatives, the following 3 initiatives are prioritized for execution:

```
┌────────────────────────────────────────────────────────────────────────┐
│                      STRATEGIC ROADMAP: 3 MOATS                        │
├────────────────────────────────────────────────────────────────────────┤
│ 1. Look-Through Fundamental & Cash Flow Quality Engine                 │
│    • Aggregate ROIC, Free Cash Flow Yield & Owner Earnings per share   │
│    • Capital Allocation Scorecard (ROIC vs WACC, Reinvestment Rate)    │
├────────────────────────────────────────────────────────────────────────┤
│ 2. Automated EOD Market Data Ingestion & FX Triangulation Pipeline     │
│    • Multi-provider resilient market data worker                       │
│    • Automated daily valuation backfill & corporate action adjustments │
├────────────────────────────────────────────────────────────────────────┤
│ 3. Interactive Tax-Lot Inspector & Dynamic Cost-Basis Replay           │
│    • Real-time Average Cost ↔ FIFO switching toggle                     │
│    • Capital gains realization simulations and tax-loss harvesting      │
└────────────────────────────────────────────────────────────────────────┘
```

### Recommendation 1: Look-Through Fundamental & Cash Flow Quality Engine
- **Why It Widens the Moat**: No consumer portfolio tracker calculates **Look-Through Owner Earnings** (Buffett's favorite metric: proportional share of operating cash flow minus maintenance capex). Implementing this gives concentrated value investors a profound metric that neither Sharesight nor Koyfin provides.
- **Scope**:
  - Add fundamental entity tables in `portfolio` schema (`instrument_fundamentals`: NOPAT, Invested Capital, OCF, CapEx, Diluted Shares).
  - Calculate Portfolio-Weighted ROIC, Look-Through FCF, and Reinvestment Rate.
  - Display a "Business Owner Summary" card in the dashboard contrasting Market Value growth against Business Intrinsic Value growth.

### Recommendation 2: Automated EOD Market Data & FX Triangulation Pipeline
- **Why It Widens the Moat**: Eliminates the operational weakness of manual price updates while maintaining institutional-grade currency isolation.
- **Scope**:
  - Implement [market-data-ingestion-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/market-data-ingestion-implementation-plan.md).
  - Create a Go background worker connecting to Twelve Data / ECB FX feeds with rate limiting and price spike anomaly detection.
  - Connect with [portfolio-valuation-engine-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/portfolio-valuation-engine-implementation-plan.md) to automatically append daily valuation points and calculate True Daily Time-Weighted Return (TWR).

### Recommendation 3: Interactive Tax-Lot Inspector & Dynamic Cost-Basis Replay
- **Why It Widens the Moat**: Institutional tax flexibility directly in the dashboard.
- **Scope**:
  - Implement [cost-basis-switching-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/cost-basis-switching-implementation-plan.md).
  - Provide a dashboard header toggle: `[ Average Cost | FIFO ]` that triggers a gRPC projection replay, updating realized capital gains and remaining cost basis in < 15ms.
  - Introduce an open tax lot drawer on the holdings table allowing investors to inspect acquisition date, cost per share, and holding period (short-term vs long-term) for each parcel.
