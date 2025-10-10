package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP Metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "omnidrop_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "omnidrop_http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "status"},
	)

	HTTPRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "omnidrop_http_request_size_bytes",
			Help:    "HTTP request size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 7), // 100B to 100MB
		},
		[]string{"method", "endpoint"},
	)

	HTTPResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "omnidrop_http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 7),
		},
		[]string{"method", "endpoint"},
	)

	// Business Metrics - Task Operations
	TaskCreationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "omnidrop_task_creations_total",
			Help: "Total number of task creation attempts",
		},
		[]string{"status"}, // success, failure
	)

	TaskCreationDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "omnidrop_task_creation_duration_seconds",
			Help:    "Task creation duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
	)

	TasksWithProjectTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "omnidrop_tasks_with_project_total",
			Help: "Total number of tasks created with project assignment",
		},
	)

	TasksWithTagsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "omnidrop_tasks_with_tags_total",
			Help: "Total number of tasks created with tags",
		},
	)

	// Business Metrics - File Operations
	FileCreationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "omnidrop_file_creations_total",
			Help: "Total number of file creation attempts",
		},
		[]string{"status"}, // success, failure
	)

	FileCreationDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "omnidrop_file_creation_duration_seconds",
			Help:    "File creation duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
	)

	FilesSizeBytes = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "omnidrop_files_size_bytes",
			Help:    "Size of files created in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 7),
		},
	)

	// AppleScript Metrics
	AppleScriptExecutionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "omnidrop_applescript_executions_total",
			Help: "Total number of AppleScript executions",
		},
		[]string{"status"}, // success, failure
	)

	AppleScriptExecutionDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "omnidrop_applescript_execution_duration_seconds",
			Help:    "AppleScript execution duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
	)

	AppleScriptErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "omnidrop_applescript_errors_total",
			Help: "Total number of AppleScript errors by type",
		},
		[]string{"error_type"}, // compilation, runtime, unknown
	)

	// OAuth Metrics
	TokenIssuedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "omnidrop_oauth_tokens_issued_total",
			Help: "Total number of OAuth tokens issued",
		},
		[]string{"client_id"},
	)

	TokenValidationTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "omnidrop_oauth_token_validations_total",
			Help: "Total number of token validation attempts",
		},
		[]string{"result"}, // success, expired, invalid
	)

	ScopeValidationFailures = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "omnidrop_oauth_scope_validation_failures_total",
			Help: "Total number of scope validation failures",
		},
		[]string{"client_id", "required_scope"},
	)
)