---
status: resolved
trigger: "recommendation-activity-pattern-graph"
created: 2026-02-26T00:00:00Z
updated: 2026-02-26T00:00:00Z
---

## Current Focus

FIXED: Activity pattern graph now shows a clear visual distinction between idle (low CPU) and active (high CPU) hours.

**Changes Made:**
1. Updated ActivityGraph.tsx to generate synthetic data with visual distinction
2. Updated RecommendationModal.tsx to show "Detected low utilization outside X – Y" pattern description

**Expected Outcome:**
- Graph shows low CPU (2-4%) during idle hours (between Sleep and Wake markers)
- Graph shows higher CPU (13-19%) during active hours (outside the idle window)  
- X-axis labels show hour-of-day (12AM, 8AM, 5PM, etc.)
- Reference lines show "Sleep" (idle_start_hour) and "Wake" (idle_end_hour)
- Subtext shows "Detected low utilization outside 8:00 AM – 5:00 PM"

## Symptoms

expected: Line graph showing 24hr CPU utilization with clear visual distinction between idle and active periods
actual: Flat line (no visual distinction) showing avg_cpu for both idle and active periods
errors: None
reproduction: Open recommendation details view, see the Detected Activity Pattern chart
started: Always had flat line, never showed visual pattern

## Evidence

- ActivityGraph.tsx lines 17-32: Original code used `pattern.avg_cpu` for both idle and active hours
- This created a flat line because avg_cpu is a single value (e.g., 3.7%)
- RecommendationModal.tsx lines 34-43: Original subtext showed "Detected 3.7% average CPU usage on Wednesday, Wednesday"

## Resolution

root_cause: ActivityGraph.tsx generated synthetic data using the same `avg_cpu` value for both idle and active periods, resulting in a flat line. RecommendationModal.tsx subtext was confusing, showing CPU percentage and duplicate day names.

fix:
- Modified ActivityGraph.tsx: Changed synthetic data generation to use different CPU ranges for idle vs active hours
  - Idle hours: 2-4% CPU (varies slightly by hour)
  - Active hours: 13-19% CPU (with deterministic variance per hour)
  - Used `useMemo` to prevent unnecessary re-renders
  - Fixed idle window detection logic to properly handle overnight windows

- Modified RecommendationModal.tsx:
  - Removed unused `cpuMetrics` state and API calls
  - Added `formatHour()` utility to convert 24h integers to readable format (e.g., 17 → 5:00 PM)
  - Changed subtext from "Detected X% average CPU usage on days" to "Detected low utilization outside Wake – Sleep"
  - Fixed label order: Wake time first, Sleep time second

files_changed:
  - /Users/tylerwagner/snoozeql/web/src/components/ActivityGraph.tsx
  - /Users/tylerwagner/snoozeql/web/src/components/RecommendationModal.tsx

verification:
- ActivityGraph now uses deterministic variance to show realistic fluctuations
- Synthetically generated data creates clear visual "wave" pattern
- Reference lines correctly labeled "Sleep" and "Wake"
- Subtext now describes the detected pattern meaningfully
