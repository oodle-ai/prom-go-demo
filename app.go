package main

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	requests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "app_requests_total",
		Help: "The total number of requests",
	}, []string{"operation", "status", "customer"})

	latency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "app_latency_seconds",
		Help:    "The latency of requests",
		Buckets: prometheus.DefBuckets,
	}, []string{"operation", "status", "customer"})

	errors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "app_errors_total",
		Help: "The total number of errors",
	}, []string{"cause", "customer", "operation"})

	dbConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "app_db_connections",
		Help: "The number of active database connections",
	})

	dbConnectionLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "app_db_connection_latency_seconds",
		Help:    "The latency of database connections",
		Buckets: prometheus.DefBuckets,
	}, []string{"customer"})
)

const (
	maxDBConnections = 1000
	numCustomers     = 5
	numOperations    = 10
)

var (
	dbPool               = make(chan struct{}, maxDBConnections)
	lastSpikeTime        time.Time
	spikeDuration        = 30 * time.Second
	spikeInterval        = 5 * time.Minute
	spikeMutex           sync.Mutex
	customers            = []string{"Alpha", "Beta", "Gamma", "Delta", "Epsilon"}
	operations           = []string{"Create", "Read", "Update", "Delete", "List", "Search", "Export", "Import", "Analyze", "Report"}
	currentSpikeCustomer string
)

func main() {
	go runClient()

	http.HandleFunc("/", handleRequest)
	http.Handle("/metrics", promhttp.Handler())

	fmt.Println("Server starting on :8080")
	http.ListenAndServe(":8080", nil)
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	operation := r.Header.Get("X-Operation")
	if operation == "" {
		operation = "unknown"
	}

	status := http.StatusOK
	customer := r.Header.Get("X-Customer")
	if customer == "" {
		customer = "unknown"
	}

	// Simulate database connection with customer-based delay
	if !acquireDBConnection(customer) {
		errors.WithLabelValues("db_connection_timeout", customer, operation).Inc()
		http.Error(w, "Database Connection Timeout", http.StatusServiceUnavailable)
		return
	}
	defer releaseDBConnection()

	// simulated processing time
	time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)

	if rand.Float32() < 0.05 {
		errors.WithLabelValues("internal_server_error", customer, operation).Inc()
		status = http.StatusInternalServerError
		http.Error(w, "Internal Server Error", status)
	} else {
		if customer != "unknown" && operation != "unknown" {
			fmt.Fprintf(w, "Hello, %s! Operation: %s", customer, operation)
		} else {
			fmt.Fprintf(w, "Hello!")
		}
	}

	requests.WithLabelValues(operation, fmt.Sprintf("%d", status), customer).Inc()
	latency.WithLabelValues(operation, fmt.Sprintf("%d", status), customer).Observe(time.Since(start).Seconds())
}

func runClient() {
	for {
		rps := getRequestsPerSecond()
		ticker := time.NewTicker(time.Second / time.Duration(rps))

		var wg sync.WaitGroup
		ctx, cancel := context.WithCancel(context.Background())

		for i := 0; i < rps; i++ {
			select {
			case <-ticker.C:
				wg.Add(1)
				go func() {
					defer wg.Done()
					customer := customers[rand.Intn(numCustomers)]
					operation := operations[rand.Intn(numOperations)]
					req, _ := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080", nil)
					req.Header.Set("X-Customer", customer)
					req.Header.Set("X-Operation", operation)
					resp, err := http.DefaultClient.Do(req)
					if err != nil {
						fmt.Println("Error making request:", err)
						return
					}
					defer resp.Body.Close()

					// Discard response body
					_, _ = io.Copy(io.Discard, resp.Body)
				}()
			case <-ctx.Done():
				break
			}
		}

		ticker.Stop()
		wg.Wait()
		cancel()
	}
}

func getRequestsPerSecond() int {
	return 500
}

func acquireDBConnection(customer string) bool {
	start := time.Now()
	delay := calculateDBConnectionDelay(customer)
	time.Sleep(delay)

	select {
	case dbPool <- struct{}{}:
		dbConnections.Inc()
		dbConnectionLatency.WithLabelValues(customer).Observe(time.Since(start).Seconds())
		return true
	case <-time.After(50 * time.Millisecond):
		return false
	}
}

func releaseDBConnection() {
	<-dbPool
	dbConnections.Dec()
}

func calculateDBConnectionDelay(customer string) time.Duration {
	spikeMutex.Lock()
	defer spikeMutex.Unlock()

	now := time.Now()
	if now.Sub(lastSpikeTime) > spikeInterval {
		lastSpikeTime = now
		currentSpikeCustomer = customers[rand.Intn(len(customers))]
	}

	if now.Sub(lastSpikeTime) > spikeDuration {
		currentSpikeCustomer = ""
	}

	if now.Sub(lastSpikeTime) < spikeDuration && customer == currentSpikeCustomer {
		// High latency for specific customer during spike: 1500-2000ms
		return time.Duration(rand.Intn(500)+1500) * time.Millisecond
	}

	// Base connection delay for all customers: 0-50ms
	return time.Duration(rand.Intn(50)) * time.Millisecond
}
