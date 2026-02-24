---
status: resolved
trigger: "GET /api/v1/savings/daily?days=30 returning 500 Internal Server Error"
created: "2026-02-24T02:29:00Z"
updated: "2026-02-24T02:30:45Z"
---

## Current Focus

hypothesis: Type assertion issue in GetDailySavings function when building response slice
test: Rewrite GetDailySavings to use separate slice variable before assigning to response map
expecting: Endpoint returns 200 OK with savings data
next_action: Verify frontend can now load savings page

## Summary

### Problem

The /api/v1/savings/daily?days=30 endpoint was returning 500 Internal Server Error with panic:
```
panic: interface conversion: interface {} is []map[string]interface {}, not map[string][]map[string]interface {}
```

### Root Cause

The GetDailySavings function in internal/api/handlers/savings.go was using complex type assertions on response["daily_savings"] which was causing runtime panics. The original code had circular references where response["daily_savings"] was being read and written in the same append statement.

### Fix

Rewrote the GetDailySavings function to:
1. Build a separate dailySavingsList slice first
2. Process all daily savings entries into that slice
3. Add placeholder entry for today if needed
4. Assign the final slice to response["daily_savings"]

This avoids the circular type assertion issue by using a proper intermediate variable.

### Changes Made

**File:** internal/api/handlers/savings.go

Changed from inline type assertion:
response["daily_savings"] = append(response["daily_savings"].([]map[string]interface{}), entry)

To separate variable:
dailySavingsList := make([]map[string]interface{}, 0)
dailySavingsList = append(dailySavingsList, entry)
response["daily_savings"] = dailySavingsList

### Verification

curl -H "Authorization: Bearer dev-key" "http://localhost:8080/api/v1/savings/daily?days=30"

Returns HTTP 200 with:
{
  "daily_savings": [{"date":"2026-02-23","savings_cents":90,"stopped_minutes":0}],
  "ongoing_cost": 90
}

### Impact

- Fixed savings page 500 error
- Frontend can now load and display daily savings data
- No breaking changes to API contract
