package main

import (
  "context"
  "fmt"
  "log"
  "net/http"
  "os"
  "time"
)

// Feature toggle manager (simple version for demo)
type FeatureFlags struct {
  PaymentRetryEnabled bool
}

var features FeatureFlags

func init() {
  features.PaymentRetryEnabled = os.Getenv("FEATURE_RETRY") == "true"
}

// HTTP handlers
func handleHealth(w http.ResponseWriter, r *http.Request) {
  if r.URL.Path == "/health/live" {
    // Liveness: always return 200 if process alive
    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "OK")
    return
  }

  if r.URL.Path == "/health/ready" {
    // Readiness: check if service ready
    // In demo: always ready
    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "Ready")
    return
  }
}

func handlePayment(w http.ResponseWriter, r *http.Request) {
  if r.Method != http.MethodPost {
    http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    return
  }

  // Simulate payment processing
  amount := r.URL.Query().Get("amount")
  if amount == "" {
    amount = "100"
  }

  // Check if retry feature enabled
  if features.PaymentRetryEnabled {
    log.Printf("Processing payment (retry ENABLED): %s", amount)
    // Simulate: might fail once, then succeed with retry
    time.Sleep(50 * time.Millisecond)
    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, `{"status":"success","amount":%s,"retries":1}`, amount)
  } else {
    log.Printf("Processing payment (retry DISABLED): %s", amount)
    time.Sleep(30 * time.Millisecond)
    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, `{"status":"success","amount":%s,"retries":0}`, amount)
  }
}

func handleMetrics(w http.ResponseWriter, r *http.Request) {
  // Simple metrics endpoint (Prometheus scrapes this)
  w.Header().Set("Content-Type", "text/plain")
  fmt.Fprintf(w, `
# HELP payments_processed_total Total payments processed
# TYPE payments_processed_total counter
payments_processed_total{status="success"} 1000
payments_processed_total{status="failed"} 5

# HELP payment_processing_duration_seconds Payment processing duration
# TYPE payment_processing_duration_seconds histogram
payment_processing_duration_seconds_bucket{le="0.01"} 200
payment_processing_duration_seconds_bucket{le="0.05"} 850
payment_processing_duration_seconds_bucket{le="0.1"} 990
payment_processing_duration_seconds_bucket{le="+Inf"} 1005
payment_processing_duration_seconds_sum 45.6
payment_processing_duration_seconds_count 1005

# HELP feature_toggle_enabled Feature toggle status
# TYPE feature_toggle_enabled gauge
feature_toggle_enabled{feature="payment_retry"} %d
`, boolToInt(features.PaymentRetryEnabled))
}

func boolToInt(b bool) int {
  if b {
    return 1
  }
  return 0
}

func main() {
  log.Printf("Starting payment service on port 8080")
  log.Printf("Feature flags: PaymentRetryEnabled=%v", features.PaymentRetryEnabled)

  // Routes
  http.HandleFunc("/health/live", handleHealth)
  http.HandleFunc("/health/ready", handleHealth)
  http.HandleFunc("/api/payments", handlePayment)
  http.HandleFunc("/metrics", handleMetrics)

  // Start server
  if err := http.ListenAndServe(":8080", nil); err != nil {
    log.Fatalf("Server error: %v", err)
  }
}
