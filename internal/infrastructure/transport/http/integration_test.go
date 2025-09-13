package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/akeemphilbert/goro/internal/conf"
	"github.com/akeemphilbert/goro/internal/infrastructure/transport/http/handlers"
	"github.com/go-kratos/kratos/v2/log"
	kratosHttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEndToEndHTTPRequestResponseCycles tests complete HTTP request/response cycles
// Requirements: 1.1, 1.2, 1.3, 6.2, 6.3
func TestEndToEndHTTPRequestResponseCycles(t *testing.T) {
	server, baseURL, cleanup := createTestServer(t)
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := server.Start(ctx); err != nil {
			t.Logf("Server start error (expected for test): %v", err)
		}
	}()

	waitForServer(t, baseURL+"/health", 5*time.Second)

	t.Run("health check endpoint", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/health/")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var result map[string]any
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "ok", result["status"])
		assert.NotNil(t, result["timestamp"])
	})

	t.Run("path parameters", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/demo/path/test123")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]any
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "test123", result["id"])
	})

	t.Run("query parameters", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/demo/query?name=test&age=25&active=true")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)

		var result map[string]any
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "test", result["name"])
		assert.Equal(t, "25", result["age"])
		assert.Equal(t, "true", result["active"])
	})

	t.Run("JSON response", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/demo/json")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var result map[string]any
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "success", result["message"])
		assert.NotNil(t, result["data"])

		data, ok := result["data"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, float64(1), data["id"])
		assert.Equal(t, "example", data["name"])
	})

	t.Run("HTTP method handling", func(t *testing.T) {
		methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

		for _, method := range methods {
			t.Run(method, func(t *testing.T) {
				req, err := http.NewRequest(method, baseURL+"/health/", nil)
				require.NoError(t, err)

				client := &http.Client{Timeout: 5 * time.Second}
				resp, err := client.Do(req)
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 500,
					"Should get a valid HTTP response for method %s, got %d", method, resp.StatusCode)

				if method == "GET" {
					assert.Equal(t, 200, resp.StatusCode)
				}
			})
		}
	})

	t.Run("CORS headers", func(t *testing.T) {
		req, err := http.NewRequest("GET", baseURL+"/health/", nil)
		require.NoError(t, err)
		req.Header.Set("Origin", "https://example.com")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)
		corsHeader := resp.Header.Get("Access-Control-Allow-Origin")
		assert.Equal(t, "*", corsHeader)
	})
}

// TestConcurrentRequestHandling tests concurrent request handling and middleware chain execution
// Requirements: 1.1, 1.2, 1.3, 6.2, 6.3
func TestConcurrentRequestHandling(t *testing.T) {
	server, baseURL, cleanup := createTestServer(t)
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := server.Start(ctx); err != nil {
			t.Logf("Server start error (expected for test): %v", err)
		}
	}()

	waitForServer(t, baseURL+"/health", 5*time.Second)

	t.Run("concurrent health checks", func(t *testing.T) {
		const numRequests = 50
		const numWorkers = 10

		var wg sync.WaitGroup
		results := make(chan error, numRequests)
		requests := make(chan int, numRequests)

		for i := 0; i < numRequests; i++ {
			requests <- i
		}
		close(requests)

		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				client := &http.Client{Timeout: 10 * time.Second}

				for requestID := range requests {
					resp, err := client.Get(baseURL + "/health/")
					if err != nil {
						results <- fmt.Errorf("worker %d, request %d: %v", workerID, requestID, err)
						continue
					}

					if resp.StatusCode != 200 {
						resp.Body.Close()
						results <- fmt.Errorf("worker %d, request %d: unexpected status %d", workerID, requestID, resp.StatusCode)
						continue
					}

					var result map[string]any
					err = json.NewDecoder(resp.Body).Decode(&result)
					resp.Body.Close()
					if err != nil {
						results <- fmt.Errorf("worker %d, request %d: decode error %v", workerID, requestID, err)
						continue
					}

					if result["status"] != "ok" {
						results <- fmt.Errorf("worker %d, request %d: unexpected status %v", workerID, requestID, result["status"])
						continue
					}

					results <- nil
				}
			}(i)
		}

		wg.Wait()
		close(results)

		successCount := 0
		var errors []error
		for result := range results {
			if result == nil {
				successCount++
			} else {
				errors = append(errors, result)
			}
		}

		assert.Equal(t, numRequests, successCount, "All requests should succeed")
		if len(errors) > 0 {
			t.Logf("Errors encountered: %v", errors[:min(5, len(errors))])
		}
	})

	t.Run("middleware chain execution under load", func(t *testing.T) {
		const numRequests = 30
		var wg sync.WaitGroup
		results := make(chan bool, numRequests)

		for i := 0; i < numRequests; i++ {
			wg.Add(1)
			go func(requestID int) {
				defer wg.Done()
				client := &http.Client{Timeout: 10 * time.Second}

				req, err := http.NewRequest("GET", baseURL+"/health/", nil)
				if err != nil {
					results <- false
					return
				}
				req.Header.Set("Origin", "https://example.com")

				resp, err := client.Do(req)
				if err != nil {
					results <- false
					return
				}
				defer resp.Body.Close()

				corsHeader := resp.Header.Get("Access-Control-Allow-Origin")
				if corsHeader != "*" {
					results <- false
					return
				}

				if resp.StatusCode != 200 {
					results <- false
					return
				}

				results <- true
			}(i)
		}

		wg.Wait()
		close(results)

		successCount := 0
		for success := range results {
			if success {
				successCount++
			}
		}

		assert.Equal(t, numRequests, successCount, "All requests should have proper middleware execution")
	})
}

// TestGracefulShutdownWithActiveConnections tests graceful shutdown with active connections
// Requirements: 6.2, 6.3
func TestGracefulShutdownWithActiveConnections(t *testing.T) {
	t.SkipNow()
	server, baseURL, cleanup := createTestServer(t)
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())

	serverDone := make(chan error, 1)
	go func() {
		serverDone <- server.Start(ctx)
	}()

	waitForServer(t, baseURL+"/health", 5*time.Second)

	t.Run("graceful shutdown behavior", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			resp, err := http.Get(baseURL + "/health/")
			require.NoError(t, err)
			assert.Equal(t, 200, resp.StatusCode)
			resp.Body.Close()
		}

		shutdownStart := time.Now()
		cancel()

		select {
		case err := <-serverDone:
			shutdownDuration := time.Since(shutdownStart)
			t.Logf("Server shutdown completed in %v with error: %v", shutdownDuration, err)
			assert.Less(t, shutdownDuration, 10*time.Second, "Shutdown should complete within 10 seconds")

		case <-time.After(12 * time.Second):
			t.Fatal("Server shutdown timed out after 12 seconds")
		}

		client := &http.Client{Timeout: 1 * time.Second}
		_, err := client.Get(baseURL + "/health/")
		assert.Error(t, err, "Server should not accept new connections after shutdown")
	})
}

// Helper functions

func createTestServer(t *testing.T) (*kratosHttp.Server, string, func()) {
	logger := log.NewStdLogger(nil)
	healthHandler := handlers.NewHealthHandler(logger)
	requestResponseHandler := handlers.NewRequestResponseHandler(logger)

	config := &conf.HTTP{
		Network:         "tcp",
		Addr:            ":18080",
		Timeout:         30 * time.Second,
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		ShutdownTimeout: 5 * time.Second,
		MaxHeaderBytes:  1048576,
	}

	server := NewHTTPServer(config, logger, healthHandler, requestResponseHandler)
	baseURL := "http://localhost:18080"

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Stop(ctx); err != nil {
			t.Logf("Server stop error: %v", err)
		}
	}

	return server, baseURL, cleanup
}

func waitForServer(t *testing.T, url string, timeout time.Duration) {
	client := &http.Client{Timeout: 1 * time.Second}
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				return
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	t.Fatalf("Server did not start within %v", timeout)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
