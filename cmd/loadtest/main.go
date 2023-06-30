package main

import (
	"context"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/bufbuild/connect-go"

	fsv1 "github.com/stillmatic/featuresheet/gen/featuresheet/v1"
	"github.com/stillmatic/featuresheet/gen/featuresheet/v1/featuresheetv1connect"
)

var (
	qps         = 1000 // configure your QPS here
	testSeconds = 10   // configure your test duration here
)

func main() {
	client := featuresheetv1connect.NewFeatureSheetServiceClient(
		http.DefaultClient,
		"http://localhost:8080",
	)

	// Create a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Create a ticker for QPS
	ticker := time.NewTicker(time.Second / time.Duration(qps))
	defer ticker.Stop()

	// Create a slice to hold all latencies
	var latencies []time.Duration

	// Run test for testSeconds
	timeout := time.After(time.Duration(testSeconds) * time.Second)

	// Mutex for synchronization
	var mu sync.Mutex

	// Run the load test until the timeout
	for {
		select {
		case <-timeout:
			// Wait for all requests to finish
			wg.Wait()
			printLatencyReport(latencies)
			return
		case <-ticker.C:
			wg.Add(1)
			go func() {
				defer wg.Done()
				start := time.Now()
				res, err := client.Evaluate(
					context.Background(),
					connect.NewRequest(&fsv1.EvaluateRequest{
						Feature:  "my_key",
						EntityId: "my_id",
					}),
				)
				elapsed := time.Since(start)
				mu.Lock()
				latencies = append(latencies, elapsed)
				mu.Unlock()
				if err != nil {
					log.Printf("Request failed: %v\n", err)
				}
				_ = res
			}()
		}
	}
}

func printLatencyReport(latencies []time.Duration) {
	var total time.Duration
	for _, latency := range latencies {
		total += latency
	}

	avgLatency := total / time.Duration(len(latencies))
	// report p90/p99
	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})
	p90 := latencies[int(float64(len(latencies))*0.9)]
	p99 := latencies[int(float64(len(latencies))*0.99)]
	log.Printf("total: %s, count: %d\n", total, len(latencies))
	log.Printf("avg: %s, p90: %s, p99: %s\n", avgLatency, p90, p99)
}
