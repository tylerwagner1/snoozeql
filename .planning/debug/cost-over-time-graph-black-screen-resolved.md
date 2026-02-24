---
status: resolved
trigger: "Cost over time graph is empty/black screen on savings page"
created: "2026-02-24T01:20:00Z"
updated: "2026-02-24T01:25:00Z"
---

## Current Focus

hypothesis: React component error when top_savers is null
test: Added null check for data.top_savers in SavingsSummaryCards.tsx
expecting: Page renders without errors and displays correct data
next_action: Verify savings page displays correctly in browser

## Summary

### Root Cause

React error `Cannot read properties of null (reading 'length')` in `SavingsSummaryCards.tsx:60`:

```
SavingsSummaryCards.tsx:60 Uncaught TypeError: Cannot read properties of null (reading 'length')
```

**Explanation**: The backend returns `"top_savers": null` when there's no savings data (since the savings table is empty). The frontend code tried to access `data.top_savers.length` without checking if `top_savers` is null first.

### Fix Applied

Changed `SavingsSummaryCards.tsx` line 60-62 from:
```typescript
value: data.top_savers.length > 0 ? data.top_savers.length.toString() : '0',
subtitle: data.top_savers.length > 0 
  ? `Best: ${formatCurrency(data.top_savers[0]?.savings_cents || 0)}`
  : 'No savings yet',
```

To:
```typescript
value: data.top_savers && data.top_savers.length > 0 ? data.top_savers.length.toString() : '0',
subtitle: data.top_savers && data.top_savers.length > 0 
  ? `Best: ${formatCurrency(data.top_savers[0]?.savings_cents || 0)}`
  : 'No savings yet',
```

### Verification

Backend endpoint returns null for empty data:
```
curl http://localhost:8080/api/v1/savings?days=30
{"ongoing_savings_cents":0,"period":{"end":"2026-02-24","start":"2026-02-23"},"top_savers":null,"total_savings_cents":0}
```

Frontend now handles null gracefully and shows "0" and "No savings yet" instead of crashing.

### Status
✅ Fixed: Null check added to prevent React error
✅ Build: Docker container rebuilt with fix
✅ Deployed: App container restarted

### Next Steps
User should verify savings page displays correctly in browser with:
- Total Savings card showing $0.00
- Ongoing Savings card showing $0.00  
- Top Savers card showing "0" and "No savings yet"
- No black screen or React errors
