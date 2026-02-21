# Phase 1: Multi-Cloud Discovery - Research

**Researched:** 2026-02-20
**Domain:** Cloud provider integration (AWS RDS, GCP Cloud SQL) and React data tables
**Confidence:** HIGH

## Summary

Phase 1 requires building upon a substantial existing foundation. The codebase already has:
- AWS RDS provider with `DescribeDBInstances`, `StartDBInstance`, `StopDBInstance` fully implemented using AWS SDK v2
- GCP Cloud SQL provider with `Instances.List`, start/stop via `ActivationPolicy` 
- Provider registry pattern for multi-provider support
- CloudAccount model with JSONB credentials storage
- Basic UI for cloud accounts and instances listing

The primary work is:
1. **Enhancing the instance persistence** - Currently instances are fetched live from cloud APIs; need to sync to database
2. **Improving discovery service** - Add connection status tracking, automatic provider re-registration when accounts change
3. **Enriching the UI** - Add sorting, filtering, connection status indicators, dashboard stats cards with pre-filtering

**Primary recommendation:** Focus on wiring existing backend capabilities to the database and UI rather than building new integrations.

## Standard Stack

The established libraries/tools for this domain:

### Core (Already in Use)
| Library | Version | Purpose | Status |
|---------|---------|---------|--------|
| aws-sdk-go-v2/service/rds | v1.116+ | AWS RDS API operations | Implemented |
| google.golang.org/api/sqladmin/v1 | latest | GCP Cloud SQL Admin API | Implemented |
| pgx/v5 | v5.x | PostgreSQL driver with JSONB support | Implemented |
| chi/v5 | v5.x | HTTP router | Implemented |

### Frontend (Already in Use)
| Library | Version | Purpose | Status |
|---------|---------|---------|--------|
| React | 18.2 | UI framework | Implemented |
| react-router-dom | 6.20 | Routing with URL params | Implemented |
| Recharts | 2.10 | Charts for cost visualization | Implemented |
| lucide-react | 0.300 | Icons | Implemented |
| tailwindcss | 3.4 | Styling | Implemented |

### Supporting (May Need to Add)
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| @tanstack/react-table | 8.x | Advanced data table with sorting/filtering | If built-in HTML table becomes unwieldy |
| react-hot-toast | 2.x | Toast notifications | For connection errors (lightweight alternative) |
| sonner | 1.x | Toast notifications | Alternative to react-hot-toast |

**Note:** The current table implementation uses native HTML `<table>` with Tailwind styling. This works well for the requirements. @tanstack/react-table adds complexity that may not be needed for sorting a few columns.

## Architecture Patterns

### Existing Backend Structure
```
internal/
├── api/
│   ├── handlers/       # HTTP handlers (partially stubbed)
│   └── middleware/     # Chi middleware (implemented)
├── config/            # Configuration loading
├── discovery/         # Discovery service (needs enhancement)
├── models/            # Data models (complete)
├── provider/
│   ├── aws/rds.go     # AWS RDS provider (implemented)
│   ├── gcp/cloudsql.go # GCP Cloud SQL provider (implemented)
│   ├── provider.go    # Provider interface
│   └── registry.go    # Multi-provider registry
└── store/
    └── postgres.go    # Database operations (partial)
```

### Pattern 1: Provider Registration per Region
**What:** Each cloud account + region combination gets its own provider instance
**Current implementation:**
```go
// Source: cmd/server/main.go lines 115-124
for _, region := range regions {
    awsProvider, err := awsprovider.NewRDSProvider(region, "", []string{}, accessKey, secretKey)
    if err == nil {
        providerKey := "aws_" + region
        providerRegistry.Register(providerKey, awsProvider)
    }
}
```
**Issue:** Provider key doesn't include account ID, causing collisions with multiple accounts in same region.

### Pattern 2: Instance Discovery Flow
**What:** Fetch from cloud APIs → sync to database → return from database
**Current state:** Fetches directly from cloud APIs, skips database sync
```go
// Source: internal/discovery/discovery.go lines 84-86
// TODO: Sync instances to database in Phase 3
// For now, just return the count
```

### Pattern 3: Connection Form Separation
**What:** Separate forms for AWS and GCP with different credential fields
**Current implementation:** Single modal with provider toggle
```typescript
// Source: web/src/pages/CloudAccountsPage.tsx
// AWS: accessKeyId, secretAccessKey
// GCP: gcpProjectId, gcpServiceKey
```

### Anti-Patterns to Avoid
- **Storing plaintext credentials in localStorage:** Already handled correctly - credentials stored server-side in encrypted JSONB
- **Single RDS client per account:** Already uses per-region clients
- **Polling without rate limiting:** Need to implement in Phase 1
- **Blocking API calls:** Already uses goroutines for background discovery

## Don't Hand-Roll

Problems that look simple but have existing solutions:

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Data table sorting | Custom sort functions | Native array.sort() + state | Simple enough for 5 columns |
| Status badge colors | Hardcoded switch statements | Status map constant | Already in codebase, extend it |
| AWS pagination | Manual marker tracking | AWS SDK v2 Paginator | SDK handles this properly |
| GCP pagination | Manual token tracking | Standard pageToken pattern | Already in SDK response |
| Toast notifications | Custom component | react-hot-toast or sonner | Handles positioning, animations, stacking |
| Form validation | Manual if-statements | HTML5 validation + required | Sufficient for MVP |

**Key insight:** The existing codebase already has most patterns implemented. The work is connecting them, not creating new abstractions.

## Common Pitfalls

### Pitfall 1: AWS 7-Day Auto-Restart
**What goes wrong:** AWS automatically restarts stopped RDS instances after 7 days
**Why it happens:** AWS feature to prevent indefinite stopped state
**How to avoid:** Track `stopped_at` timestamp, implement re-stop mechanism
**Warning signs:** Instance shows "available" when schedule says it should be stopped
**Phase 1 action:** Add `last_stopped_at` field to Instance model, UI warning indicator

### Pitfall 2: Read Replicas Cannot Be Stopped
**What goes wrong:** Attempting to stop read replica returns error
**Why it happens:** AWS doesn't allow stopping read replicas
**How to avoid:** Check `ReadReplicaSourceDBInstanceIdentifier` field in DescribeDBInstances response
**Warning signs:** Stop API call fails with "InvalidDBInstanceState"
**Phase 1 action:** Add `is_read_replica` flag during discovery, disable stop button in UI

### Pitfall 3: Instance State Race Conditions
**What goes wrong:** UI shows stale state after start/stop action
**Why it happens:** State transitions (stopping→stopped, starting→available) take time
**How to avoid:** Optimistic UI update + polling for confirmation
**Warning signs:** Button shows "Start" immediately after user clicked "Stop"
**Phase 1 action:** Show transitional states (starting, stopping) in UI with pulse animation (already partially implemented)

### Pitfall 4: Multi-Account Region Collision
**What goes wrong:** Provider registry overwrites entries when two accounts have same region
**Why it happens:** Current provider key is just `aws_{region}`, not `{accountId}_{region}`
**How to avoid:** Include account ID in provider registry key
**Phase 1 action:** Update provider registration: `providerKey := fmt.Sprintf("aws_%s_%s", account.ID, region)`

### Pitfall 5: GCP Service Account JSON Validation
**What goes wrong:** User pastes malformed JSON, connection silently fails
**Why it happens:** No validation on JSON structure before saving
**How to avoid:** Validate JSON structure client-side, test connection server-side
**Phase 1 action:** Add JSON.parse() check in frontend, test-connection endpoint for GCP

### Pitfall 6: Discovery Runs Before Providers Registered
**What goes wrong:** Discovery service starts before cloud accounts loaded from DB
**Why it happens:** Async initialization order
**How to avoid:** Wait for provider registration before starting discovery goroutine
**Current code handles this:** Provider registration happens before `discoveryService.RunContinuous(ctx)`

## Code Examples

### AWS RDS Instance Listing (Already Implemented)
```go
// Source: internal/provider/aws/rds.go lines 67-85
func (p *RDSProvider) ListDatabases(ctx context.Context) ([]models.Instance, error) {
    var instances []models.Instance
    result, err := p.rdsClient.DescribeDBInstances(ctx, &rds.DescribeDBInstancesInput{})
    if err != nil {
        return nil, fmt.Errorf("failed to describe DB instances: %w", err)
    }
    for _, db := range result.DBInstances {
        instance, err := p.dbInstanceToModel(db)
        if err != nil {
            return nil, err
        }
        instances = append(instances, instance)
    }
    return instances, nil
}
```

### AWS RDS Pagination (Needs Implementation)
```go
// Source: AWS SDK v2 documentation
func (p *RDSProvider) ListDatabasesPaginated(ctx context.Context) ([]models.Instance, error) {
    var instances []models.Instance
    paginator := rds.NewDescribeDBInstancesPaginator(p.rdsClient, &rds.DescribeDBInstancesInput{})
    
    for paginator.HasMorePages() {
        page, err := paginator.NextPage(ctx)
        if err != nil {
            return nil, fmt.Errorf("failed to get page: %w", err)
        }
        for _, db := range page.DBInstances {
            instance, err := p.dbInstanceToModel(db)
            if err != nil {
                return nil, err
            }
            instances = append(instances, instance)
        }
    }
    return instances, nil
}
```

### GCP Cloud SQL Instance Listing (Already Implemented)
```go
// Source: internal/provider/gcp/cloudsql.go lines 36-54
func (p *CloudSQLProvider) ListDatabases(ctx context.Context) ([]models.Instance, error) {
    var instances []models.Instance
    result, err := p.sqlAdminService.Instances.List(p.projectID).Context(ctx).Do()
    if err != nil {
        return nil, fmt.Errorf("failed to list Cloud SQL instances: %w", err)
    }
    for _, db := range result.Items {
        instance, err := p.instanceToModel(db)
        if err != nil {
            return nil, err
        }
        instances = append(instances, instance)
    }
    return instances, nil
}
```

### Instance Sync to Database (Needs Implementation)
```go
// Pattern: Upsert discovered instances
func (s *InstanceStore) UpsertInstance(ctx context.Context, instance *models.Instance) error {
    query := `
        INSERT INTO instances (
            cloud_account_id, provider, provider_id, name, region, 
            instance_type, engine, status, managed, tags, hourly_cost_cents
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        ON CONFLICT (provider, provider_id) DO UPDATE SET
            name = EXCLUDED.name,
            status = EXCLUDED.status,
            tags = EXCLUDED.tags,
            hourly_cost_cents = EXCLUDED.hourly_cost_cents,
            updated_at = NOW()
        RETURNING id`
    
    tagsJSON, _ := json.Marshal(instance.Tags)
    return s.db.QueryRowContext(ctx, query,
        instance.CloudAccountID, instance.Provider, instance.ProviderID,
        instance.Name, instance.Region, instance.InstanceType, instance.Engine,
        instance.Status, instance.Managed, tagsJSON, instance.HourlyCostCents,
    ).Scan(&instance.ID)
}
```

### React Table with Sorting (Pattern to Follow)
```typescript
// Pattern: Simple sorting without external library
const [sortConfig, setSortConfig] = useState<{key: keyof Instance, direction: 'asc' | 'desc'}>({
  key: 'name',
  direction: 'asc'
});

const sortedInstances = useMemo(() => {
  const sorted = [...instances].sort((a, b) => {
    if (a[sortConfig.key] < b[sortConfig.key]) return sortConfig.direction === 'asc' ? -1 : 1;
    if (a[sortConfig.key] > b[sortConfig.key]) return sortConfig.direction === 'asc' ? 1 : -1;
    return 0;
  });
  return sorted;
}, [instances, sortConfig]);

const handleSort = (key: keyof Instance) => {
  setSortConfig(prev => ({
    key,
    direction: prev.key === key && prev.direction === 'asc' ? 'desc' : 'asc'
  }));
};
```

### Status Chip Component (Existing Pattern to Extend)
```typescript
// Source: web/src/pages/InstancesPage.tsx - extend this pattern
const statusColors: Record<string, string> = {
  'available': 'bg-green-500/10 text-green-400 border-green-500/30',
  'running': 'bg-green-500/10 text-green-400 border-green-500/30',
  'stopped': 'bg-slate-500/10 text-slate-400 border-slate-500/30',
  'stopping': 'bg-orange-500/10 text-orange-400 border-orange-500/30 animate-pulse',
  'starting': 'bg-blue-500/10 text-blue-400 border-blue-500/30 animate-pulse',
  'pending': 'bg-yellow-500/10 text-yellow-400 border-yellow-500/30',
  // GCP states
  'RUNNABLE': 'bg-green-500/10 text-green-400 border-green-500/30',
  'SUSPENDED': 'bg-slate-500/10 text-slate-400 border-slate-500/30',
  'MAINTENANCE': 'bg-yellow-500/10 text-yellow-400 border-yellow-500/30',
};
```

### Connection Status Indicator (New Pattern Needed)
```typescript
// Pattern: Connection status chip for cloud accounts
type ConnectionStatus = 'connected' | 'syncing' | 'failed' | 'unknown';

const connectionStatusColors: Record<ConnectionStatus, string> = {
  connected: 'bg-green-500/10 text-green-400 border-green-500/30',
  syncing: 'bg-blue-500/10 text-blue-400 border-blue-500/30 animate-pulse',
  failed: 'bg-red-500/10 text-red-400 border-red-500/30',
  unknown: 'bg-slate-500/10 text-slate-400 border-slate-500/30',
};
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| AWS SDK v1 | AWS SDK v2 | 2021 | Already using v2, no migration needed |
| google-api-go-client | Same, still current | - | No change needed |
| Polling cloud APIs on every request | Cache in DB + periodic sync | Best practice | Needs implementation in Phase 1 |

**Current in codebase:**
- AWS SDK v2 already in use
- GCP Cloud SQL Admin API already in use
- React 18 with hooks pattern

**Patterns already established:**
- Provider interface abstraction
- Registry pattern for multi-provider
- Chi router with middleware chain
- Tailwind + CSS variables for theming

## What Exists vs What Needs Building

### Backend - Exists
- [x] AWS RDS provider with credentials
- [x] GCP Cloud SQL provider (basic)
- [x] Provider registry pattern
- [x] CloudAccount CRUD in database
- [x] Instance model with all required fields
- [x] Database schema for instances table
- [x] API endpoints for instances, cloud-accounts

### Backend - Needs Building
- [ ] Instance persistence (UpsertInstance)
- [ ] Instance listing from database (not just cloud API)
- [ ] Connection status tracking (connected/syncing/failed)
- [ ] Provider re-registration when accounts added/deleted
- [ ] Read replica detection during discovery
- [ ] GCP service account JSON credential support
- [ ] Dashboard stats endpoint enhancement

### Frontend - Exists
- [x] InstancesPage with table layout
- [x] CloudAccountsPage with add/delete forms
- [x] Dashboard with stats cards
- [x] Status chip styling
- [x] Navigation

### Frontend - Needs Building
- [ ] Column sorting on instances table
- [ ] Column filters (dropdown menus)
- [ ] Account name column in instances table
- [ ] Connection status chips on CloudAccountsPage
- [ ] Skeleton loading states
- [ ] Toast notifications for errors
- [ ] Clickable stats cards that pre-filter
- [ ] Empty state for no connected accounts

## Open Questions

Things that couldn't be fully resolved:

1. **GCP Service Account Authentication**
   - What we know: Current code uses `cloudsql.NewService(context.Background())` which relies on ADC
   - What's unclear: How to use service account JSON key passed from frontend
   - Recommendation: Use `google.golang.org/api/option` with `option.WithCredentialsJSON()`

2. **Instance Sync Frequency**
   - What we know: Config has `Discovery_interval` in seconds
   - What's unclear: Optimal interval (too frequent = rate limiting, too slow = stale data)
   - Recommendation: Default 60 seconds, configurable per-account

3. **Multi-Region Discovery Performance**
   - What we know: Each region requires separate API calls
   - What's unclear: Should we parallelize region discovery?
   - Recommendation: Use goroutines with semaphore to limit concurrent calls (3-5 max)

## Sources

### Primary (HIGH confidence)
- AWS SDK v2 Go documentation - https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/rds
- Existing codebase: internal/provider/aws/rds.go (implemented and working)
- Existing codebase: internal/provider/gcp/cloudsql.go (implemented and working)
- Existing codebase: cmd/server/main.go (server initialization pattern)

### Secondary (MEDIUM confidence)
- GCP Cloud SQL Admin API - https://pkg.go.dev/google.golang.org/api/sqladmin/v1

### Tertiary (LOW confidence)
- React table patterns - general React documentation (verify current best practices)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Verified from existing codebase, all libraries already in use
- Architecture: HIGH - Patterns established in existing code
- Pitfalls: HIGH - AWS/GCP documented behaviors, code review confirms gaps
- UI Patterns: MEDIUM - Following existing patterns, some new components needed

**Research date:** 2026-02-20
**Valid until:** 2026-03-20 (30 days - stable libraries, established patterns)

---

*Phase: 01-multi-cloud-discovery*
*Research completed: 2026-02-20*
