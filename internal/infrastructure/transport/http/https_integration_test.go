package http

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/akeemphilbert/goro/internal/conf"
	"github.com/akeemphilbert/goro/internal/infrastructure/transport/http/handlers"
	"github.com/go-kratos/kratos/v2/log"
)

func TestHTTPS_EndToEnd(t *testing.T) {
	// Create temporary directory for test certificates
	tempDir, err := os.MkdirTemp("", "https_e2e_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Generate test certificate and key
	certFile := filepath.Join(tempDir, "server.crt")
	keyFile := filepath.Join(tempDir, "server.key")

	if err := generateTestCertificate(certFile, keyFile); err != nil {
		t.Fatalf("Failed to generate test certificate: %v", err)
	}

	// Create server configuration
	logger := log.NewStdLogger(os.Stdout)
	healthHandler := handlers.NewHealthHandler(logger)
	requestResponseHandler := handlers.NewRequestResponseHandler(logger)

	config := &conf.HTTP{
		Network: "tcp",
		Addr:    ":0", // Let OS assign a free port
		Timeout: conf.Duration(30 * time.Second),
		TLS: conf.TLS{
			Enabled:  true,
			CertFile: certFile,
			KeyFile:  keyFile,
		},
	}

	// Create HTTPS server
	server := NewHTTPServerWithoutResourceHandler(config, logger, healthHandler, requestResponseHandler)
	if server == nil {
		t.Fatal("Failed to create HTTPS server")
	}

	// Start server in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := server.Start(ctx); err != nil {
			t.Logf("Server start error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Get the actual server address
	endpoint, err := server.Endpoint()
	if err != nil {
		t.Fatalf("Failed to get server endpoint: %v", err)
	}
	serverAddr := endpoint.Host

	// Create HTTPS client that accepts self-signed certificates
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // Accept self-signed cert for testing
			},
		},
		Timeout: 10 * time.Second,
	}

	// Test HTTPS health endpoint
	t.Run("HTTPS health check", func(t *testing.T) {
		resp, err := client.Get(fmt.Sprintf("https://%s/health", serverAddr))
		if err != nil {
			t.Fatalf("Failed to make HTTPS request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		var healthResp map[string]interface{}
		if err := json.Unmarshal(body, &healthResp); err != nil {
			t.Fatalf("Failed to parse JSON response: %v", err)
		}

		if healthResp["status"] != "ok" {
			t.Errorf("Expected status 'ok', got %v", healthResp["status"])
		}
	})

	// Test HTTPS with different endpoints
	t.Run("HTTPS status endpoint", func(t *testing.T) {
		resp, err := client.Get(fmt.Sprintf("https://%s/status", serverAddr))
		if err != nil {
			t.Fatalf("Failed to make HTTPS request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	// Verify TLS connection details
	t.Run("TLS connection verification", func(t *testing.T) {
		conn, err := tls.Dial("tcp", serverAddr, &tls.Config{
			InsecureSkipVerify: true,
		})
		if err != nil {
			t.Fatalf("Failed to establish TLS connection: %v", err)
		}
		defer conn.Close()

		state := conn.ConnectionState()

		// Verify TLS version
		if state.Version < tls.VersionTLS12 {
			t.Errorf("Expected TLS version >= 1.2, got %d", state.Version)
		}

		// Verify cipher suite is secure
		if state.CipherSuite == 0 {
			t.Errorf("No cipher suite negotiated")
		}

		// Verify certificate
		if len(state.PeerCertificates) == 0 {
			t.Errorf("No peer certificates received")
		}
	})

	// Stop server
	if err := server.Stop(ctx); err != nil {
		t.Logf("Server stop error: %v", err)
	}
}

func TestHTTP_vs_HTTPS_Comparison(t *testing.T) {
	logger := log.NewStdLogger(os.Stdout)
	healthHandler := handlers.NewHealthHandler(logger)
	requestResponseHandler := handlers.NewRequestResponseHandler(logger)

	// Test HTTP server
	t.Run("HTTP server", func(t *testing.T) {
		config := &conf.HTTP{
			Network: "tcp",
			Addr:    ":0",
			Timeout: conf.Duration(30 * time.Second),
			TLS: conf.TLS{
				Enabled: false,
			},
		}

		server := NewHTTPServerWithoutResourceHandler(config, logger, healthHandler, requestResponseHandler)
		if server == nil {
			t.Fatal("Failed to create HTTP server")
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			if err := server.Start(ctx); err != nil {
				t.Logf("HTTP server start error: %v", err)
			}
		}()

		time.Sleep(100 * time.Millisecond)

		// Test HTTP request
		client := &http.Client{Timeout: 5 * time.Second}
		endpoint, err := server.Endpoint()
		if err != nil {
			t.Fatalf("Failed to get server endpoint: %v", err)
		}
		resp, err := client.Get(fmt.Sprintf("http://%s/health", endpoint.Host))
		if err != nil {
			t.Fatalf("Failed to make HTTP request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		if err := server.Stop(ctx); err != nil {
			t.Logf("HTTP server stop error: %v", err)
		}
	})

	// Test HTTPS server
	t.Run("HTTPS server", func(t *testing.T) {
		// Create temporary directory for test certificates
		tempDir, err := os.MkdirTemp("", "https_comparison_test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Generate test certificate and key
		certFile := filepath.Join(tempDir, "server.crt")
		keyFile := filepath.Join(tempDir, "server.key")

		if err := generateTestCertificate(certFile, keyFile); err != nil {
			t.Fatalf("Failed to generate test certificate: %v", err)
		}

		config := &conf.HTTP{
			Network: "tcp",
			Addr:    ":0",
			Timeout: conf.Duration(30 * time.Second),
			TLS: conf.TLS{
				Enabled:  true,
				CertFile: certFile,
				KeyFile:  keyFile,
			},
		}

		server := NewHTTPServerWithoutResourceHandler(config, logger, healthHandler, requestResponseHandler)
		if server == nil {
			t.Fatal("Failed to create HTTPS server")
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			if err := server.Start(ctx); err != nil {
				t.Logf("HTTPS server start error: %v", err)
			}
		}()

		time.Sleep(100 * time.Millisecond)

		// Test HTTPS request
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
			Timeout: 5 * time.Second,
		}
		endpoint, err := server.Endpoint()
		if err != nil {
			t.Fatalf("Failed to get server endpoint: %v", err)
		}
		resp, err := client.Get(fmt.Sprintf("https://%s/health", endpoint.Host))
		if err != nil {
			t.Fatalf("Failed to make HTTPS request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		if err := server.Stop(ctx); err != nil {
			t.Logf("HTTPS server stop error: %v", err)
		}
	})
}
