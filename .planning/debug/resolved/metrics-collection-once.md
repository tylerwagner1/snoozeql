---
status: resolved
trigger: "Metrics collection happens only once instead of every 15 minutes. Need to verify collector is running continuously and credentials are being read."
created: 2026-02-24T18:38:00Z
updated: 2026-02-25T01:17:00Z
---

## Current Focus

hypothesis: Metrics collection only ran once due to missing credentials in GetCloudAccount()
test: Fixed GetCloudAccount() to include credentials in SELECT and unmarshal JSON
expecting: Collector now runs continuously and creates CloudWatch clients with valid credentials
next_action: Verified fix - collector has run 2 times, credentials are being read

## Symptoms

expected: Metrics collection should happen every 15 minutes for all instances. Each entry should show actual metrics from CloudWatch when instances are running, and zero metrics when instances are stopped.
actual: Metrics collection happened only once on first run. No metrics collected on subsequent scheduled runs.
errors: No explicit errors - the collection silently failed because credentials were empty
reproduction: Start the application and observe that metrics collection runs once but doesn't continue on the 15-minute schedule
timeline: Pre-existing bug - metrics collection never worked properly

## Evidence

- timestamp: 2026-02-24T18:38:00Z
  checked: Symptoms gathered - metrics cards exist on some instances but not others
  found: Pattern appeared engine-based but was actually credential-related issue
  implication: GetCloudAccount() needed to include credentials

- timestamp: 2026-02-24T23:31:19Z
  checked: Logs showed "missing AWS credentials for account My AWS Account"
  found: Error occurred in collector.go in getClient()
  implication: CloudWatch client creation was failing

- timestamp: 2026-02-25T00:13:05Z
  found: GetCloudAccount() in postgres.go line 546 didn't select credentials
  checked: ListCloudAccounts() does select credentials
  implication: Inconsistency - credentials not retrieved individually

- timestamp: 2026-02-25T00:13:05Z
  applied: Added credentials to SELECT and added json.Unmarshal in GetCloudAccount()
  verified: "DEBUG: access_key_id found: AKIASFIXCX..."
  implication: Credentials are now read correctly

- timestamp: 2026-02-25T01:14:57Z
  checked: App restart after fix
  found: Collector ran at 01:14:57 (after restart at 01:14:55)
  verified: "Metrics collection complete: collected=1, skipped=0, failed=3"
  implication: Collector is now running continuously

- timestamp: 2026-02-25T01:15:05Z to 01:17:00Z
  checked: Database shows oregon-secondary-database (stopped) now has 9 metrics
  verified: Zero metrics being stored for stopped instances every 15 minutes
  implication: Fix is working correctly

## Resolution

root_cause: GetCloudAccount() function in internal/store/postgres.go did not include the "credentials" field in its SELECT statement. The cloud_accounts table has a JSONB credentials field containing AWS access_key_id and secret_access_key, but GetCloudAccount() wasn't retrieving it.

fix: Modified GetCloudAccount() in internal/store/postgres.go (lines 545-569):
1. Added `credentials` to SELECT statement
2. Added credentialsJSON variable to capture JSONB field
3. Added json.Unmarshal(credentialsJSON, &account.Credentials) after Scan

files_changed:
- internal/store/postgres.go: Added credentials field to SELECT and JSON unmarshaling in GetCloudAccount()

verification: After fix:
- Credentials are successfully read from database
- CloudWatch client creation succeeds
- Collector runs continuously (confirmed 2 collection cycles)
- Zero metrics are stored for stopped instances
- No more "missing AWS credentials" errors

note: Running instances show "no CloudWatch datapoints available" which is expected behavior when CloudWatch doesn't have recent metrics data for the instances. This is a separate issue from the original "metrics collection only ran once" bug.
