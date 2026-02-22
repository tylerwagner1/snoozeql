# Database migrations for SnoozeQL

This directory contains SQL migration files for the SnoozeQL database schema.

## Migration File Naming Convention

Migrations follow the pattern: `NNN_description.up.sql` and `NNN_description.down.sql`

Where `NNN` is a zero-padded sequential number (e.g., 001, 002, 003).

## Running Migrations

### Using Go migrate CLI (recommended)
```bash
# Install migrate
brew install golang-migrate

# Run migrations
migrate -path migrations -database "postgres://user:pass@localhost:5432/snoozeql" up

# Rollback
migrate -path migrations -database "postgres://user:pass@localhost:5432/snoozeql" down
```

### Using Docker Compose
```bash
# Run migrations in Docker
docker-compose run --rm app migrate -path /migrations -database "postgres://..." up
```

### Programmatic (in Go code)
```go
db, err := store.NewPostgres(os.Getenv("DATABASE_URL"))
if err != nil {
    log.Fatal(err)
}

if err := db.Migrate("./migrations"); err != nil {
    log.Fatal(err)
}
```

## Creating New Migrations

```bash
# Create new migration files
migrate create -ext sql -dir migrations -seq new_feature

# This creates: migrations/000001_new_feature.up.sql and migrations/000001_new_feature.down.sql
```

## Migration Best Practices

1. **Always write both up and down migrations**
2. **Keep migrations atomic** - one change per migration
3. **Don't modify existing migrations** - create new ones
4. **Use transactions** for complex migrations
5. **Add indexes** for frequently queried columns
6. **Test migrations** in a staging environment first

## Rollback Strategy

- Down migrations should safely reverse up migrations
- Handle data loss scenarios carefully
- Provide warnings for destructive operations
- Test rollback in isolation
