# User Story: Interactive SVG Performance Chart & Time Range Filtering

> **Story ID**: US-VAL-103  
> **Epic**: Portfolio Valuation & Performance Analytics  
> **Target**: Primary Investor Application (`web/apps/main-app`), BFF GraphQL Orchestrator (`bff/`), Portfolio Microservice (`services/portfolio-api`), Protobuf (`proto/portfolio/v1/portfolio.proto`), Database (`portfolio.portfolio_valuations`)  
> **Status**: **DONE**  
> **Implementation Plan**: [performance-chart-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/performance-chart-implementation-plan.md)  
> **Roadmap Reference**: [FEATURES.md (Feature #7)](file:///Users/oscargarcia/workspace/graphfolio/FEATURES.md#L17)  
> **Standards Reference**: [financial-calculations skill](file:///Users/oscargarcia/workspace/graphfolio/.agents/skills/financial-calculations/SKILL.md), [AGENTS.md](file:///Users/oscargarcia/workspace/graphfolio/AGENTS.md)  

---

## 1. User Story Statement

> **As an** investor tracking my wealth progression,  
> **I want to** view an interactive, responsive SVG performance chart displaying my portfolio's historical value and Time-Weighted Return (TWR) curve across multiple timeframes (`1D`, `1W`, `1M`, `1Y`, `ALL`), with hover crosshairs and floating valuation tooltips,  
> **So that** I can visually inspect my portfolio growth, compare market value against cash reserves, and evaluate period returns over any market window.

---

## 2. Business Context & Visual Requirements

1. **Elimination of Static Placeholders**:
   - The initial prototype relied on a static image file (`portfolio_chart.jpg`) and inactive buttons. Investors require a live, data-driven charting canvas driven by daily snapshots stored in `portfolio.portfolio_valuations`.
2. **Timeframe Slicing & Boundary Returns**:
   - Supports five dynamic time windows: `1D` (past 24h/daily), `1W` (past 7 days), `1M` (past 30 days), `1Y` (past 365 days), and `ALL` (from earliest portfolio inception to present).
   - Period return ($ and %) calculates the exact performance difference between the opening anchor valuation point and the terminal closing valuation point of the chosen timeframe.
3. **Smooth Catmull-Rom Bezier Rendering**:
   - The SVG curve is rendered with smooth Catmull-Rom splines, a subtle vertical gradient area fill, and dynamic hover effects.
4. **Rich Floating Tooltips**:
   - Moving the cursor across the chart activates an interactive crosshair and floating glassmorphic tooltip displaying the hovered date, total portfolio valuation, market value vs. cash balance breakdown, and cumulative return.

---

## 3. Acceptance Criteria

### 3.1 Backend Time-Series API & History Calculation
- [x] **AC-1 (History RPC)**: `proto/portfolio/v1/portfolio.proto` defines `GetPortfolioHistory(GetPortfolioHistoryRequest)` returning `GetPortfolioHistoryResponse` with valuation points and period return metrics.
- [x] **AC-2 (Timeframe Boundary Filtering)**: `history.go` queries `portfolio.portfolio_valuations` bounded by the calculated start date for `1D`, `1W`, `1M`, `1Y`, and `ALL`.
- [x] **AC-3 (Period Return Metrics)**: API returns `period_return_amount` and `period_return_percent` measuring exact growth across the selected window.

### 3.2 GraphQL BFF Resolution
- [x] **AC-4 (GraphQL Query)**: BFF exposes `portfolioHistory(timeframe: HistoryTimeframe!): PortfolioHistory!` mapping gRPC valuation points and exact decimal scalars.

### 3.3 Interactive Web Component (`PerformanceChart.tsx`)
- [x] **AC-5 (Dynamic SVG Curve)**: Replaces static images with an SVG path generated using Catmull-Rom or cubic Bezier smoothing.
- [x] **AC-6 (Timeframe Selector Tabs)**: Active timeframe buttons (`1D`, `1W`, `1M`, `1Y`, `ALL`) trigger queries and update the active button state with glassmorphic styling.
- [x] **AC-7 (Interactive Hover Crosshair)**: Mouse/touch tracking draws a vertical crosshair guide line and glowing anchor dot on the nearest date point.
- [x] **AC-8 (Floating Tooltip Card)**: Displays date, Total Value, Market Value, Cash Value, and cumulative TWR percentage formatted in portfolio currency.
- [x] **AC-9 (Responsive Scalability)**: Chart dimensions adapt dynamically to viewport width using SVG `viewBox` coordinates without clipping or distortion.

---

## 4. Technical & Implementation Notes

- **Protobuf**: `proto/portfolio/v1/portfolio.proto` defines `HistoryTimeframe`, `ValuationPoint`, and `PortfolioHistory`.
- **Backend Service**: `services/portfolio-api/internal/service/history.go` handles timeframe date math and period gain calculations.
- **Frontend Component**: `web/apps/main-app/src/components/PerformanceChart.tsx` and `.css`.
- **Benchmark Synergy**: Seamlessly supports benchmark comparison overlays (S&P 500, NASDAQ-100, ASX 200).
