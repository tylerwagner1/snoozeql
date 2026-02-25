## ROOT CAUSE FOUND

**Debug Session:** `.planning/debug/resolved/metrics-collection-once.md`

**Root Cause:** `GetCloudAccount()` function in `internal/store/postgres.go` was not reading the `credentials` field from the database. The `cloud_accounts` table has a JSONB `credentials` field containing AWS `access_key_id` and `secret_access_key`, but `GetCloudAccount()` wasn't selecting it.

**Fix Applied:**
1. Added `credentials` to SELECT in `GetCloudAccount()`
2. Added `json.Unmarshal(credentialsJSON, &account.Credentials)` to parse the JSONB field

**Files Changed:**
- `internal/store/postgres.go` (lines 545-569)

**Verification:**
After fix, collector logs show credentials being read:
```
DEBUG: access_key_id found: AKIASFIXCX...
DEBUG: secret_access_key found: coblQriWMe...
```

Collector runs continuously (2 cycles confirmed). Zero metrics are stored for stopped instances.

**Note:** The issue where running instances show "no CloudWatch datapoints available" is separate from the original bug. This indicates CloudWatch doesn't have recent metrics data for those instances, which may be due to:
- CloudWatch data latency
- Instances being too new
- CloudWatch metrics being disabled
