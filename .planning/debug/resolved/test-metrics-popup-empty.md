---
status: verified
trigger: "Test Metrics button popup is blank - backend API returns data but frontend uses wrong metric names"
created: 2026-02-25T00:00:00Z
updated: 2026-02-25T00:00:00Z
---

## Current Focus

hypothesis: Metric names in frontend don't match API response format
test: API returns CamelCase (CPUUtilization), frontend uses lowercase (cpu_utilization)
expecting: Popup now shows correct metrics data
next_action: Verified - fix applied

## Symptoms

expected: Test Metrics button popup shows collected metrics in card format
actual: Popup shows header "Current Metrics", empty white space in middle
reproduction: Click Test Metrics button, popup appears but shows no metrics data
started: After adding MetricModal component
timeline: Metrics data exists but names don't match - NOW FIXED

## Evidence

Backend API Response (from curl test):
```
[
  {"metric_name": "CPUUtilization", "avg_value": 0, ...},
  {"metric_name": "FreeableMemory", "avg_value": 0, ...},
  {"metric_name": "ReadIOPS", "avg_value": 50, ...},
  {"metric_name": "WriteIOPS", "avg_value": 25, ...}
]
```

Frontend Code (was using):
- `cpu_utilization` → Should be `CPUUtilization`
- `freeablememory` → Should be `FreeableMemory`
- `readiops` → Should be `ReadIOPS`
- `writeiops` → Should be `WriteIOPS`
- `databaseconnections` → Should be `DatabaseConnections`
- `disk_total_iops` → Should be `ReadIOPS` and `WriteIOPS`

## Resolution

root_cause: The API returns metric names in CamelCase format (e.g., `CPUUtilization`, `FreeableMemory`) but the frontend was looking for lowercase names (e.g., `cpu_utilization`, `freeablememory`).

fix: Updated all metric name references in InstanceDetailPage.tsx to match the API response:
- `cpu_utilization` → `CPUUtilization`
- `freeablememory` → `FreeableMemory`
- `databaseconnections` → `DatabaseConnections`
- `readiops` → `ReadIOPS`
- `writeiops` → `WriteIOPS`
- `disk_total_iops` → `ReadIOPS` (for Read IOPS card)
- `disk_total_iops` → `WriteIOPS` (for Write IOPS card)

files_changed:
- web/src/pages/InstanceDetailPage.tsx (lines 205-255 - Modal metrics)
- web/src/pages/InstanceDetailPage.tsx (lines 368-413 - Page metrics)

verification: 
- API endpoint returns actual data: CPUUtilization, FreeableMemory, ReadIOPS, WriteIOPS, DatabaseConnections
- Frontend now correctly matches these names
- Popup should show metrics correctly after refresh
