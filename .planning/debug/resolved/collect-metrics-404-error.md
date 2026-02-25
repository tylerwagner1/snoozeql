---
status: verified
trigger: "i pressed the Test Metrics button, i get failed to collect metrics in UI pop-up badge and the console tab shows 404"
created: 2026-02-25T00:00:00Z
updated: 2026-02-25T00:00:00Z
---

## Current Focus

hypothesis: Backend server in Docker is running old code without collect-metrics route
test: Check docker container logs and rebuild/restart app container
expecting: collect-metrics endpoint exists in current code but Docker container needs restart
next_action: Verified - endpoint works after Docker container rebuild

## Symptoms

expected: Test Metrics button should collect metrics and display results in modal
actual: Button returns 404 Not Found for POST /api/v1/instances/:id/collect-metrics
errors: 
  - UI shows: "Failed to collect metrics"
  - Console: POST http://localhost:8080/api/v1/instances/15728f75-8407-4e44-83a5-b8c6793ee79f/collect-metrics 404 (Not Found)
  - curl test confirmed: 404 page not found
  - curl test after fix: {"message":"Metrics collected","success":"true"}
reproduction: Click Test Metrics button on instance details page
started: Docker container running old server code without collect-metrics endpoint
timeline: Backend code had endpoint added but Docker container not rebuilt - FIXED by rebuilding container

## Eliminated

- hypothesis: API URL is wrong
  evidence: Frontend URL is http://localhost:8080/api/v1/instances/:id/collect-metrics which is standard REST format
  timestamp: 2026-02-25T00:00:00Z

- hypothesis: Instance ID is invalid
  evidence: Instance ID 15728f75-8407-4e44-83a5-b8c6793ee79f is valid and used elsewhere in page
  timestamp: 2026-02-25T00:00:00Z

- hypothesis: collect-metrics endpoint missing from backend code
  evidence: cmd/server/main.go line 672 shows r.Post("/instances/{id}/collect-metrics", ...) is present
  timestamp: 2026-02-25T00:00:00Z

## Evidence

- timestamp: 2026-02-25T00:00:00Z
  checked: Frontend API call in api.ts line 254
  found: collectInstanceMetrics uses POST /instances/${instanceId}/collect-metrics
  implication: This should map to /api/v1/instances/:id/collect-metrics on backend

- timestamp: 2026-02-25T00:00:00Z
  checked: Backend routes in Go server cmd/server/main.go line 672
  found: r.Post("/instances/{id}/collect-metrics", ...) endpoint exists at line 672
  implication: Code has endpoint, but Docker container not rebuilt

- timestamp: 2026-02-25T00:00:00Z
  checked: Docker containers running
  found: snoozeql-app-1 container running old code
  implication: Container needs rebuild to pick up collect-metrics endpoint

- timestamp: 2026-02-25T00:00:00Z
  checked: curl to collect-metrics endpoint BEFORE rebuild
  found: Returns 404 page not found with Authorization header
  implication: Endpoint does not exist in running container

- timestamp: 2026-02-25T00:00:00Z
  checked: docker-compose build app
  found: Container rebuilt successfully with updated code
  implication: New container includes collect-metrics endpoint

- timestamp: 2026-02-25T00:00:00Z
  checked: docker-compose up -d app
  found: Container restarted successfully
  implication: New container is now running with updated code

- timestamp: 2026-02-25T00:00:00Z
  checked: curl to collect-metrics endpoint AFTER rebuild
  found: Returns {"message":"Metrics collected","success":"true"}
  implication: Endpoint is now working correctly

## Resolution

root_cause: The collect-metrics endpoint exists in the backend code (cmd/server/main.go line 672) but the Docker container running the Go server was not rebuilt after the code change. The container was running stale code without the endpoint.

fix: Rebuilt and restarted the Docker container using:
1. docker-compose build app
2. docker-compose up -d app

This picked up the current backend code including the collect-metrics endpoint at line 672.

files_changed:
- cmd/server/main.go line 672 (endpoint already exists in code)
- Docker image rebuilt from source

verification: 
- Endpoint tested with curl after rebuild: {"message":"Metrics collected","success":"true"}
- Test Metrics button should now work in the UI

## Debug Session Location

.planning/debug/resolved/collect-metrics-404-error.md
