---
quick: 011
description: GCP Cloud Account Coming Soon message
completed: 2026-02-27
commit: 2b4b9365
---

# Quick Task 011: GCP Cloud Account Coming Soon

## What Was Done

Added a "Coming Soon!" message when users select GCP as the provider in the Cloud Accounts page. Since GCP Cloud SQL integration was never fully implemented, this provides a friendly user experience instead of showing a broken form.

## Changes Made

**File:** `web/src/pages/CloudAccountsPage.tsx`

1. **Added Rocket icon import** from lucide-react
2. **Added conditional rendering** for when `form.provider === 'gcp' && !editingAccount`:
   - Purple gradient styled info box
   - Rocket icon in decorative container
   - "GCP Support Coming Soon!" heading
   - Explanation text about AWS being available
   - "Connect AWS Instead" button that switches provider back to AWS
3. **Wrapped existing credential form** in the else branch of the conditional
4. **AWS form unchanged** - continues to work exactly as before

## UI Preview

When user clicks GCP button:
```
┌─────────────────────────────────────┐
│         🚀 (Rocket icon)            │
│                                     │
│   GCP Support Coming Soon!          │
│                                     │
│   We're working on GCP Cloud SQL    │
│   integration. AWS RDS support is   │
│   fully available now.              │
│                                     │
│   [Connect AWS Instead]             │
└─────────────────────────────────────┘
```

## Verification

- ✅ `npm run build` passes (2.00s)
- ✅ GCP selection shows Coming Soon message
- ✅ AWS credential form still works
- ✅ Can switch between AWS and GCP
- ✅ Editing existing GCP accounts (if any existed) would still show form

## Commit

`2b4b9365` - feat(quick-011): show Coming Soon message for GCP cloud accounts

---

*Quick task 011 completed 2026-02-27*
