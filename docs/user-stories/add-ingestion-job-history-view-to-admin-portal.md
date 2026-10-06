# User Story: Add Ingestion Job History View to Admin Portal

> **Target**: Internal Operations Portal (`web/apps/admin-app`)  
> **Status**: Ready for Implementation  
> **Implementation Plan**: [ingestion-job-history-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/ingestion-job-history-implementation-plan.md)  
> **Roadmap Reference**: [FEATURES.md (Feature #18)](file:///Users/oscargarcia/workspace/graphfolio/FEATURES.md#L27)

---

## 1. User Story

> **As an** Admin Portal user,  
> **I want to** view a historical log of past data ingestion jobs,  
> **So that** I can monitor system health and troubleshoot failed imports.

---

## 2. Acceptance Criteria

- [ ] **Data Display**: Create a new table/view in the Admin Portal displaying past ingestion jobs.
- [ ] **Execution Time**: The table must show the exact timestamp of when the job started and finished.
- [ ] **Job Status**: Include a clear status indicator for each job (e.g., `Success`, `Failed`, `In Progress`, `Partial Success`).
- [ ] **Records Affected**: Display the total number of records processed (split by Successful vs. Failed rows).
- [ ] **Sorting**: Users should be able to sort the jobs by date (newest first by default).

---

## 3. Related Documentation

- **Architecture & Implementation Plan**: [ingestion-job-history-implementation-plan.md](file:///Users/oscargarcia/workspace/graphfolio/docs/plans/ingestion-job-history-implementation-plan.md)
- **Feature Matrix & Roadmap**: [FEATURES.md](file:///Users/oscargarcia/workspace/graphfolio/FEATURES.md)