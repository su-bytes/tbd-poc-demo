package main

import (
	"fmt"
	"log"
	"net/http"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "order_svc_http_requests_total",
			Help: "Total HTTP requests",
		},
		[]string{"method", "endpoint", "status", "version"},
	)
	httpDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "order_svc_http_duration_seconds",
			Help: "HTTP request latency",
			Buckets: []float64{.1, .25, .5, 1, 2.5, 5},
		},
		[]string{"method", "endpoint", "version"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal, httpDuration)
}

func orderHandler(w http.ResponseWriter, r *http.Request) {
	timer := prometheus.NewTimer(httpDuration.WithLabelValues(r.Method, "/order/create", "blue"))
	defer timer.ObserveDuration()
	
	httpRequestsTotal.WithLabelValues(r.Method, "/order/create", "200", "blue").Inc()
	fmt.Fprintf(w, "Order created successfully\n")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"healthy","version":"v1.0.0"}`)
}

func main() {
	http.HandleFunc("/order/create", orderHandler)
	http.HandleFunc("/health/live", healthHandler)
	http.HandleFunc("/health/ready", healthHandler)
	http.Handle("/metrics", promhttp.Handler())
	
	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
