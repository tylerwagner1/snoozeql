package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	awssdkconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/rds"

	"snoozeql/internal/analyzer"
	"snoozeql/internal/api/handlers"
	"snoozeql/internal/api/middleware"
	"snoozeql/internal/config"
	"snoozeql/internal/discovery"
	"snoozeql/internal/metrics"
	"snoozeql/internal/models"
	"snoozeql/internal/provider"
	awsprovider "snoozeql/internal/provider/aws"
	gcpprovider "snoozeql/internal/provider/gcp"
	"snoozeql/internal/store"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

// Global instances
var (
	discoveryService    *discovery.DiscoveryService
	instanceStore       *store.InstanceStore
	accountStore        *store.CloudAccountStore
	eventStore          *store.EventStore
	scheduleStore       *store.ScheduleStore
	recommendationStore *store.RecommendationStore
	metricsStore        *metrics.MetricsStore
	metricsCollector    *metrics.MetricsCollector
)

// BulkOperationRequest represents a request to start/stop multiple instances
type BulkOperationRequest struct {
	InstanceIDs []string `json:"instance_ids"`
}

// BulkOperationResponse represents the result of a bulk operation
type BulkOperationResponse struct {
	Success []string         `json:"success"`
	Failed  []OperationError `json:"failed"`
}

// OperationError represents a failed operation
type OperationError struct {
	InstanceID string `json:"instance_id"`
	Error      string `json:"error"`
}

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)

	fmt.Println("SnoozeQL - Database Sleeper for Dev/Test Environments")
	fmt.Println("========================================================")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("failed to validate config: %v", err)
	}

	// Initialize database
	db, err := store.NewPostgres(cfg.Database_url)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	migrationDir := "./deployments/docker/migrations"
	if err := db.Migrate(migrationDir); err != nil {
		log.Printf("warning: migration check failed: %v", err)
	}

	log.Printf("✓ Connected to database")
	log.Printf("✓ AWS_ACCESS_KEY_ID set: %v", cfg.AWS_access_key != "")
	log.Printf("✓ AWS_SECRET_ACCESS_KEY set: %v", cfg.AWS_secret_key != "")
	log.Printf("✓ Server will start on http://%s:%s", cfg.Server_host, cfg.Server_port)

	// Initialize provider registry and discovery
	providerRegistry := provider.NewRegistry()

	// Load cloud accounts from database and register providers
	storeInstance := store.NewCloudAccountStore(db)
	cloudAccounts, err := storeInstance.ListCloudAccounts()
	if err != nil {
		log.Printf("Warning: Failed to load cloud accounts: %v", err)
	} else {
		log.Printf("✓ Loaded %d cloud accounts from database", len(cloudAccounts))

		for _, account := range cloudAccounts {
			if account.Provider == "aws" {
				var accessKey, secretKey string
				if cred, ok := account.Credentials["aws_access_key_id"]; ok {
					if str, ok := cred.(string); ok {
						accessKey = str
						log.Printf("DEBUG: Found access_key_id: %s", str[:10]+"...")
					}
				}
				if cred, ok := account.Credentials["aws_secret_access_key"]; ok {
					if str, ok := cred.(string); ok {
						secretKey = str
						log.Printf("DEBUG: Found secret_access_key: %s", str[:10]+"...")
					}
				}

				if accessKey == "" || secretKey == "" {
					log.Printf("Warning: Skipping AWS account %s - missing credentials", account.Name)
					continue
				}

				regions := account.Regions
				if len(regions) == 0 {
					regions = []string{"us-east-1"}
				}

				for _, region := range regions {
					awsProvider, err := awsprovider.NewRDSProvider(region, "", []string{}, accessKey, secretKey)
					if err == nil {
						providerKey := fmt.Sprintf("aws_%s_%s", account.ID, region)
						providerRegistry.Register(providerKey, awsProvider)
						log.Printf("✓ Registered AWS provider for account: %s (region: %s, key: %s)", account.Name, region, providerKey)
					} else {
						log.Printf("Warning: Failed to register AWS provider for %s in region %s: %v", account.Name, region, err)
					}
				}
			} else if account.Provider == "gcp" {
				var projectID, serviceAccountKey string
				if cred, ok := account.Credentials["gcp_project_id"]; ok {
					if str, ok := cred.(string); ok {
						projectID = str
					}
				}
				if cred, ok := account.Credentials["gcp_service_account_key"]; ok {
					if str, ok := cred.(string); ok {
						serviceAccountKey = str
					}
				}

				if projectID == "" {
					log.Printf("Warning: Skipping GCP account %s - missing project ID", account.Name)
					continue
				}

				gcpProvider, err := gcpprovider.NewCloudSQLProvider(projectID, "", []string{}, serviceAccountKey)
				if err != nil {
					log.Printf("Warning: Failed to create GCP provider for %s: %v", account.Name, err)
					continue
				}

				providerKey := fmt.Sprintf("gcp_%s", account.ID)
				providerRegistry.Register(providerKey, gcpProvider)
				log.Printf("✓ Registered GCP provider for account: %s (project: %s, key: %s)", account.Name, projectID, providerKey)
			} else {
				log.Printf("Skipping %s provider (not supported yet): %s", account.Provider, account.Name)
			}
		}
	}

	if len(providerRegistry.Providers) == 0 {
		log.Printf("Warning: No cloud accounts registered, instances will not be discovered")
	}

	// Create store instances for discovery
	instanceStore = store.NewInstanceStore(db)
	accountStore = store.NewCloudAccountStore(db)
	eventStore = store.NewEventStore(db)
	scheduleStore = store.NewScheduleStore(db)
	recommendationStore = store.NewRecommendationStore(db)

	// Initialize metrics store and collector first (before analyzer)
	metricsStore = metrics.NewMetricsStore(db)
	metricsCollector = metrics.NewMetricsCollector(
		metricsStore,
		instanceStore,
		accountStore,
		15, // 15-minute collection interval per CONTEXT.md
	)

	// Initialize recommendation analyzer
	thresholdConfig := analyzer.ThresholdConfig{
		DefaultInactivityHours: 8,
		DetectionDays:          7,
		ConfidenceMinimum:      0.5,
	}
	analyzer := analyzer.NewAnalyzer(providerRegistry, recommendationStore, metricsStore, thresholdConfig)

	discoveryService = discovery.NewDiscoveryService(providerRegistry, instanceStore, accountStore, eventStore, cfg.Discovery_enabled, cfg.Discovery_interval, []string{})

	// Start discovery in background
	ctx := context.Background()
	go discoveryService.RunContinuous(ctx)

	// Start metrics collection in background
	go metricsCollector.RunContinuous(ctx)
	log.Printf("✓ Started metrics collector (15-minute interval)")

	// Initialize router
	r := chi.NewRouter()

	// Chi middleware
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.Logger)

	// Custom middleware
	r.Use(middleware.NewRequestID)
	r.Use(middleware.TrackRequestDuration)
	r.Use(middleware.CORS)

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Use(middleware.APIKeyAuth(cfg))
		r.Route("/v1", func(r chi.Router) {
			// Health/Status
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"ok","service":"snoozeql"}`))
			})

			// Statistics
			r.Get("/stats", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)

				// Get instance counts from database
				instances, err := instanceStore.ListInstances(r.Context())
				if err != nil {
					log.Printf("ERROR listing instances for stats: %v", err)
					// Return default response on error
					w.Write([]byte(`{
						"total_instances": 0,
						"running_instances": 0,
						"stopped_instances": 0,
						"savings_7d": 0,
						"pending_actions": 0
					}`))
					return
				}

				runningCount := 0
				stoppedCount := 0
				savings7d := 0.0
				for _, inst := range instances {
					// Map instance status to running/stopped
					switch inst.Status {
					case "available", "running", "starting":
						runningCount++
						savings7d += float64(inst.HourlyCostCents) * 24 * 7 / 100
					case "stopped", "stopping":
						stoppedCount++
					}
				}

				// Get pending recommendations count
				recommendations, err := instanceStore.ListRecommendationsByStatus(r.Context(), "pending")
				if err != nil {
					log.Printf("ERROR listing recommendations: %v", err)
					recommendations = []map[string]interface{}{}
				}

				stats := map[string]interface{}{
					"total_instances":   len(instances),
					"running_instances": runningCount,
					"stopped_instances": stoppedCount,
					"savings_7d":        savings7d,
					"pending_actions":   len(recommendations),
				}

				json.NewEncoder(w).Encode(stats)
			})

			// Instances - returns persisted instances from database
			r.Get("/instances", func(w http.ResponseWriter, r *http.Request) {
				instances, err := instanceStore.ListInstances(r.Context())
				if err != nil {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					log.Printf("ERROR listing instances from database: %v", err)
					w.Write([]byte(`{"error":"Failed to list instances"}`))
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(instances)
			})

			r.Get("/instances/{id}", func(w http.ResponseWriter, r *http.Request) {
				instanceID := chi.URLParam(r, "id")
				log.Printf("DEBUG: Get instance by ID: %s", instanceID)

				ctx := r.Context()
				instance, err := instanceStore.GetInstanceByID(ctx, instanceID)
				if err != nil {
					log.Printf("ERROR: Instance not found: %s", instanceID)
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusNotFound)
					w.Write([]byte(`{"error":"Instance not found"}`))
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(instance)
			})

			r.Post("/instances/{id}/start", func(w http.ResponseWriter, r *http.Request) {
				instanceID := chi.URLParam(r, "id")

				ctx := r.Context()
				// Try to find instance in database first
				// Note: instance ProviderName is required to find the correct provider
				instance, err := instanceStore.GetInstanceByProviderID(ctx, "", instanceID)
				var providerName string
				if err != nil {
					// fallback to discovery
					instances, err := discoveryService.ListAllDatabases(ctx)
					if err != nil {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusInternalServerError)
						log.Printf("ERROR listing instances: %v", err)
						w.Write([]byte(`{"error":"Failed to list instances"}`))
						return
					}

					var foundInstance *models.Instance
					for i := range instances {
						if instances[i].ID == instanceID {
							foundInstance = &instances[i]
							break
						}
					}

					if foundInstance == nil {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusNotFound)
						log.Printf("ERROR: Instance not found: %s", instanceID)
						w.Write([]byte(fmt.Sprintf(`{"error":"Instance not found: %s"}`, instanceID)))
						return
					}

					providerName = foundInstance.ProviderName
				} else {
					// Found in database, use ProviderName from database
					providerName = instance.ProviderName
				}

				log.Printf("DEBUG: Calling StartDatabase with provider=%s, id=%s", providerName, instanceID)
				if err := discoveryService.StartDatabase(ctx, providerName, instanceID); err != nil {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					log.Printf("ERROR starting instance %s: %v", instanceID, err)
					w.Write([]byte(fmt.Sprintf(`{"error":"Failed to start instance: %v","instance_id":"%s"}`, err, instanceID)))
					return
				}

				log.Printf("SUCCESS: StartDatabase call completed for instance %s (provider=%s)", instanceID, providerName)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(fmt.Sprintf(`{"success":true,"instance_id":"%s","provider":"%s","status":"starting"}`, instanceID, providerName)))
			})

			r.Post("/instances/{id}/stop", func(w http.ResponseWriter, r *http.Request) {
				instanceID := chi.URLParam(r, "id")

				log.Printf("DEBUG: Stop endpoint called with instanceID=%s", instanceID)

				ctx := r.Context()
				// Try to find instance in database first
				// Note: provider_name is required to find the correct provider, so we use empty provider
				instance, err := instanceStore.GetInstanceByProviderID(ctx, "", instanceID)
				log.Printf("DEBUG: GetInstanceByProviderID returned: instance=%v, err=%v", instance, err)

				var providerName string
				if err != nil {
					// fallback to discovery
					instances, err := discoveryService.ListAllDatabases(ctx)
					if err != nil {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusInternalServerError)
						log.Printf("ERROR listing instances: %v", err)
						w.Write([]byte(`{"error":"Failed to list instances"}`))
						return
					}

					var foundInstance *models.Instance
					for i := range instances {
						if instances[i].ID == instanceID {
							foundInstance = &instances[i]
							break
						}
					}

					if foundInstance == nil {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusNotFound)
						log.Printf("ERROR: Instance not found: %s", instanceID)
						w.Write([]byte(fmt.Sprintf(`{"error":"Instance not found: %s"}`, instanceID)))
						return
					}

					providerName = foundInstance.ProviderName
				} else {
					// Found in database, use provider_name from database
					providerName = instance.ProviderName
				}

				log.Printf("DEBUG: Calling StopDatabase with provider=%s, id=%s", providerName, instanceID)
				if err := discoveryService.StopDatabase(ctx, providerName, instanceID); err != nil {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					log.Printf("ERROR stopping instance %s: %v", instanceID, err)
					w.Write([]byte(fmt.Sprintf(`{"error":"Failed to stop instance: %v","instance_id":"%s"}`, err, instanceID)))
					return
				}

				log.Printf("SUCCESS: StopDatabase call completed for instance %s (provider=%s)", instanceID, providerName)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(fmt.Sprintf(`{"success":true,"instance_id":"%s","provider":"%s","status":"stopping"}`, instanceID, providerName)))
			})

			// Bulk stop instances
			r.Post("/instances/bulk-stop", func(w http.ResponseWriter, r *http.Request) {
				var req BulkOperationRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`{"error":"Invalid request body"}`))
					return
				}

				if len(req.InstanceIDs) == 0 {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`{"error":"No instance IDs provided"}`))
					return
				}

				ctx := r.Context()
				success := []string{}
				failed := []OperationError{}

				for _, instanceID := range req.InstanceIDs {
					// Get instance from database to find provider and current status
					instance, err := instanceStore.GetInstanceByProviderID(ctx, "", instanceID)
					if err != nil {
						// Try by ID directly
						instances, listErr := instanceStore.ListInstances(ctx)
						if listErr != nil {
							failed = append(failed, OperationError{InstanceID: instanceID, Error: "Instance not found"})
							continue
						}
						var found *models.Instance
						for i := range instances {
							if instances[i].ID == instanceID {
								found = &instances[i]
								break
							}
						}
						if found == nil {
							failed = append(failed, OperationError{InstanceID: instanceID, Error: "Instance not found"})
							continue
						}
						instance = found
					}

					// Check if instance is in a stoppable state
					if instance.Status != "available" && instance.Status != "running" {
						failed = append(failed, OperationError{
							InstanceID: instanceID,
							Error:      fmt.Sprintf("Instance not in stoppable state (current: %s)", instance.Status),
						})
						continue
					}

					// Stop the instance using provider_name (the registered provider)
					if err := discoveryService.StopDatabase(ctx, instance.ProviderName, instance.ProviderID); err != nil {
						failed = append(failed, OperationError{InstanceID: instanceID, Error: err.Error()})
						continue
					}

					// Log the event
					event := &models.Event{
						InstanceID:     instance.ID,
						EventType:      "sleep",
						TriggeredBy:    "manual",
						PreviousStatus: instance.Status,
						NewStatus:      "stopping",
					}
					if err := eventStore.CreateEvent(ctx, event); err != nil {
						log.Printf("Warning: Failed to log event for instance %s: %v", instanceID, err)
					}

					success = append(success, instanceID)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(BulkOperationResponse{Success: success, Failed: failed})
			})

			// Bulk start instances
			r.Post("/instances/bulk-start", func(w http.ResponseWriter, r *http.Request) {
				var req BulkOperationRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`{"error":"Invalid request body"}`))
					return
				}

				if len(req.InstanceIDs) == 0 {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`{"error":"No instance IDs provided"}`))
					return
				}

				ctx := r.Context()
				success := []string{}
				failed := []OperationError{}

				for _, instanceID := range req.InstanceIDs {
					// Get instance from database
					instance, err := instanceStore.GetInstanceByProviderID(ctx, "", instanceID)
					if err != nil {
						instances, listErr := instanceStore.ListInstances(ctx)
						if listErr != nil {
							failed = append(failed, OperationError{InstanceID: instanceID, Error: "Instance not found"})
							continue
						}
						var found *models.Instance
						for i := range instances {
							if instances[i].ID == instanceID {
								found = &instances[i]
								break
							}
						}
						if found == nil {
							failed = append(failed, OperationError{InstanceID: instanceID, Error: "Instance not found"})
							continue
						}
						instance = found
					}

					// Check if instance is in a startable state
					if instance.Status != "stopped" {
						failed = append(failed, OperationError{
							InstanceID: instanceID,
							Error:      fmt.Sprintf("Instance not in startable state (current: %s)", instance.Status),
						})
						continue
					}

					// Start the instance
					if err := discoveryService.StartDatabase(ctx, instance.ProviderName, instance.ProviderID); err != nil {
						failed = append(failed, OperationError{InstanceID: instanceID, Error: err.Error()})
						continue
					}

					// Log the event
					event := &models.Event{
						InstanceID:     instance.ID,
						EventType:      "wake",
						TriggeredBy:    "manual",
						PreviousStatus: instance.Status,
						NewStatus:      "starting",
					}
					if err := eventStore.CreateEvent(ctx, event); err != nil {
						log.Printf("Warning: Failed to log event for instance %s: %v", instanceID, err)
					}

					success = append(success, instanceID)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(BulkOperationResponse{Success: success, Failed: failed})
			})

			// Schedules (using ScheduleHandler with real store)
			scheduleHandler := handlers.NewScheduleHandler(scheduleStore, instanceStore, eventStore)
			r.Get("/schedules", scheduleHandler.GetAllSchedules)
			r.Get("/schedules/{id}", func(w http.ResponseWriter, r *http.Request) {
				id := chi.URLParam(r, "id")
				scheduleHandler.GetSchedule(w, r, id)
			})
			r.Post("/schedules", scheduleHandler.CreateSchedule)
			r.Put("/schedules/{id}", func(w http.ResponseWriter, r *http.Request) {
				id := chi.URLParam(r, "id")
				scheduleHandler.UpdateSchedule(w, r, id)
			})
			r.Delete("/schedules/{id}", func(w http.ResponseWriter, r *http.Request) {
				id := chi.URLParam(r, "id")
				scheduleHandler.DeleteSchedule(w, r, id)
			})
			r.Post("/schedules/{id}/enable", func(w http.ResponseWriter, r *http.Request) {
				id := chi.URLParam(r, "id")
				scheduleHandler.EnableSchedule(w, r, id)
			})
			r.Post("/schedules/{id}/disable", func(w http.ResponseWriter, r *http.Request) {
				id := chi.URLParam(r, "id")
				scheduleHandler.DisableSchedule(w, r, id)
			})
			r.Post("/schedules/preview-filter", scheduleHandler.PreviewFilter)

			// Recommendations
			recommendationHandler := handlers.NewRecommendationHandler(
				recommendationStore, instanceStore, scheduleStore, providerRegistry, analyzer,
			)
			r.Get("/recommendations", func(w http.ResponseWriter, r *http.Request) {
				recommendationHandler.GetAllRecommendations(w, r)
			})
			r.Get("/recommendations/{id}", func(w http.ResponseWriter, r *http.Request) {
				id := chi.URLParam(r, "id")
				recommendationHandler.GetRecommendation(w, r, id)
			})
			r.Post("/recommendations/generate", func(w http.ResponseWriter, r *http.Request) {
				recommendationHandler.GenerateRecommendations(w, r)
			})
			r.Post("/recommendations/{id}/apply", func(w http.ResponseWriter, r *http.Request) {
				id := chi.URLParam(r, "id")
				recommendationHandler.ConfirmRecommendation(w, r, id)
			})
			r.Post("/recommendations/{id}/ignore", func(w http.ResponseWriter, r *http.Request) {
				id := chi.URLParam(r, "id")
				recommendationHandler.DismissRecommendation(w, r, id)
			})

			// Metrics
			r.Get("/instances/{id}/metrics", func(w http.ResponseWriter, r *http.Request) {
				instanceID := chi.URLParam(r, "id")
				log.Printf("DEBUG: Get metrics for instance: %s", instanceID)

				ctx := r.Context()
				metrics, err := metricsStore.GetLatestMetrics(ctx, instanceID)
				if err != nil {
					log.Printf("ERROR: Failed to get metrics for instance %s: %v", instanceID, err)
					// Return empty array on error (not 404 - instance exists, just no metrics)
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode([]models.HourlyMetric{})
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(metrics)
			})

			r.Post("/instances/{id}/collect-metrics", func(w http.ResponseWriter, r *http.Request) {
				instanceID := chi.URLParam(r, "id")
				log.Printf("DEBUG: Collect metrics for instance: %s", instanceID)

				ctx := r.Context()
				instance, err := instanceStore.GetInstanceByID(ctx, instanceID)
				if err != nil {
					log.Printf("ERROR: Instance not found: %s", instanceID)
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusNotFound)
					w.Write([]byte(`{"error":"Instance not found"}`))
					return
				}

				// Check if AWS provider (GCP not supported for collection yet)
				if instance.Provider != "aws" {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`{"error":"Metrics collection not supported for this provider"}`))
					return
				}

				// Call metrics collector for single instance
				if err := metricsCollector.CollectInstance(ctx, *instance); err != nil {
					log.Printf("ERROR: Failed to collect metrics for %s: %v", instanceID, err)
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(fmt.Sprintf(`{"error":"Failed to collect metrics: %v"}`, err)))
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"success": "true", "message": "Metrics collected"})
			})

			// Events/Audit Log
			r.Get("/events", func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()

				// Parse optional pagination params
				limit := 50
				offset := 0
				if l := r.URL.Query().Get("limit"); l != "" {
					if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
						limit = parsed
					}
				}
				if o := r.URL.Query().Get("offset"); o != "" {
					if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
						offset = parsed
					}
				}

				events, err := eventStore.ListEvents(ctx, limit, offset)
				if err != nil {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					log.Printf("ERROR listing events: %v", err)
					w.Write([]byte(`{"error":"Failed to list events"}`))
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(events)
			})

			// Test connection endpoint
			r.Post("/test-connection", func(w http.ResponseWriter, r *http.Request) {
				var input struct {
					Provider string         `json:"provider"`
					Region   string         `json:"region"`
					Creds    map[string]any `json:"credentials"`
				}
				if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`{"error":"Invalid request body"}`))
					return
				}

				if input.Provider != "aws" {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`{"error":"Only AWS connections are supported for testing"}`))
					return
				}

				awsAccessKey, ok1 := input.Creds["aws_access_key_id"].(string)
				awsSecretKey, ok2 := input.Creds["aws_secret_access_key"].(string)
				if !ok1 || !ok2 {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`{"error":"Missing AWS credentials"}`))
					return
				}

				// Create AWS config with provided credentials
				awsCfg, err := awssdkconfig.LoadDefaultConfig(context.Background(),
					awssdkconfig.WithRegion(input.Region),
					awssdkconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(awsAccessKey, awsSecretKey, "")))
				if err != nil {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"error":"Failed to load AWS config"}`))
					return
				}

				rdsClient := rds.NewFromConfig(awsCfg)

				// Test connection by listing RDS instances
				_, err = rdsClient.DescribeDBInstances(context.Background(), &rds.DescribeDBInstancesInput{
					MaxRecords: awssdk.Int32(20),
				})
				if err != nil {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(fmt.Sprintf(`{"error":"Connection failed: %v","details":"%v"}`, err, err)))
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"success":true,"message":"Connection successful"}`))
			})

			// Cloud accounts
			r.Get("/cloud-accounts", func(w http.ResponseWriter, r *http.Request) {
				log.Printf("DEBUG: Listing cloud accounts...")
				accounts, err := store.NewCloudAccountStore(db).ListCloudAccounts()
				if err != nil {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					log.Printf("ERROR listing cloud accounts: %v", err)
					w.Write([]byte(`{"error":"Failed to load accounts"}`))
					return
				}
				log.Printf("DEBUG: Found %d cloud accounts", len(accounts))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				response := make([]map[string]any, len(accounts))
				for i, a := range accounts {
					response[i] = map[string]any{
						"id":                a.ID,
						"name":              a.Name,
						"provider":          a.Provider,
						"regions":           a.Regions,
						"connection_status": a.ConnectionStatus,
						"last_sync_at":      a.LastSyncAt,
						"last_error":        a.LastError,
						"created_at":        a.CreatedAt.Format("2006-01-02T15:04:05Z"),
					}
				}
				json.NewEncoder(w).Encode(response)
			})

			r.Post("/cloud-accounts", func(w http.ResponseWriter, r *http.Request) {
				var input struct {
					Name        string         `json:"name"`
					Provider    string         `json:"provider"`
					Regions     []string       `json:"regions"`
					Credentials map[string]any `json:"credentials"`
				}
				if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusBadRequest)
					log.Printf("ERROR: Invalid request body: %v", err)
					w.Write([]byte(`{"error":"Invalid request body"}`))
					return
				}
				log.Printf("DEBUG: Creating account: name=%s, provider=%s, regions=%v", input.Name, input.Provider, input.Regions)

				account := &models.CloudAccount{
					ID:          "",
					Name:        input.Name,
					Provider:    input.Provider,
					Credentials: input.Credentials,
					Regions:     input.Regions,
					CreatedAt:   time.Now(),
				}

				if err := store.NewCloudAccountStore(db).CreateCloudAccount(account); err != nil {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					log.Printf("ERROR creating cloud account: %v", err)
					w.Write([]byte(`{"error":"Failed to save account"}`))
					return
				}

				log.Printf("SUCCESS: Created account with ID: %s", account.ID)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(`{"success":true}`))
			})

			r.Delete("/cloud-accounts/{id}", func(w http.ResponseWriter, r *http.Request) {
				id := chi.URLParam(r, "id")
				if err := store.NewCloudAccountStore(db).DeleteCloudAccount(id); err != nil {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"error":"Failed to delete account"}`))
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"success":true}`))
			})
		})
	})

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = cfg.Server_port
	}

	addr := fmt.Sprintf("%s:%s", cfg.Server_host, port)
	log.Printf("starting server on %s", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
