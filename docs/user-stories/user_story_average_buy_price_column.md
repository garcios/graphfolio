# User Story: Add "AVG BUY PRICE" Column to Investments Holding Table

**Story ID:** US-INV-104  
**Epic:** Portfolio Holdings Management  
**Status:** Ready for Refinement  

---

## 1. User Story Statement

**As an** active investor,  
**I want to** view my average buy price directly alongside the current market price for each asset in my holdings table,  
**So that** I can instantly evaluate position performance and determine whether each asset is currently trading above or below my cost basis.

---

## 2. Acceptance Criteria

### AC 1: Placement & Layout
* **Column Name:** A new column labeled **`AVG BUY PRICE`** (or `Avg Buy Price`, following system case standards) is added to the holdings table.
* **Positioning:** The column must be placed directly adjacent to the existing **`Price`** column (e.g., `... | Price | AVG BUY PRICE | ...`).
* **Visual Formatting:** Numbers are right-aligned to maintain consistency with financial metrics.

### AC 2: Data Calculation & Precision
* Displays the volume-weighted average purchase price across all open lots for that holding:
  $$\text{Average Buy Price} = \frac{\sum (\text{Buy Quantity} \times \text{Execution Price})}{\text{Total Current Shares/Units}}$$
* Values are formatted in the active portfolio currency with appropriate decimal precision (standard 2 decimal places for equities, higher precision for crypto/fractional shares where applicable).
* If cost basis data is missing or unavailable (e.g., external transfer without trade history), display an empty placeholder (`—` or `N/A`) rather than `$0.00`.

### AC 3: Sorting & Interaction
* Users can sort the holdings table in ascending and descending order by clicking the **`AVG BUY PRICE`** column header.
* Hovering over the column header displays an informational tooltip:  
  > *"The weighted average price paid per share/unit across all open lots."*

### AC 4: Responsive & Customization Behavior
* On mobile/tablet viewports, if columns are condensed or scrollable, `Price` and `AVG BUY PRICE` should remain visually grouped together.
* If user-configurable columns are supported, `AVG BUY PRICE` is included in the column manager and set to **visible by default**.

---

## 3. Technical & Implementation Notes

* **API Payload:** Verify that the holdings API endpoint (`GET /api/v1/portfolio/holdings`) exposes `average_buy_price` / `cost_basis_per_share` directly to avoid computationally expensive client-side calculations over extensive order histories.
* **Corporate Actions:** Ensure backend calculations adjust the average cost basis appropriately following events like stock splits or reverse splits.