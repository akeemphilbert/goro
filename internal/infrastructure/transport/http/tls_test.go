package http

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/akeemphilbert/goro/internal/conf"
	"github.com/akeemphilbert/goro/internal/infrastructure/transport/http/handlers"
	"github.com/go-kratos/kratos/v2/log"
)

func TestNewHTTPServer_TLS_Disabled(t *testing.T) {
	logger := log.NewStdLogger(os.Stdout)
	healthHandler := handlers.NewHealthHandler(logger)
	requestResponseHandler := handlers.NewRequestResponseHandler(logger)

	config := &conf.HTTP{
		Network: "tcp",
		Addr:    ":8080",
		Timeout: 30 * time.Second,
		TLS: conf.TLS{
			Enabled: false,
		},
	}

	server := NewHTTPServer(config, logger, healthHandler, requestResponseHandler)
	if server == nil {
		t.Fatal("Expected server to be created, got nil")
	}

	// Server should be created successfully without TLS
	// We can't easily test the internal TLS config without exposing it,
	// but we can verify the server was created
}

func TestNewHTTPServer_TLS_Enabled_ValidCerts(t *testing.T) {
	// Create temporary directory for test certificates
	tempDir, err := os.MkdirTemp("", "tls_server_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Generate test certificate and key
	certFile := filepath.Join(tempDir, "test.crt")
	keyFile := filepath.Join(tempDir, "test.key")

	if err := generateTestCertificate(certFile, keyFile); err != nil {
		t.Fatalf("Failed to generate test certificate: %v", err)
	}

	logger := log.NewStdLogger(os.Stdout)
	healthHandler := handlers.NewHealthHandler(logger)
	requestResponseHandler := handlers.NewRequestResponseHandler(logger)

	config := &conf.HTTP{
		Network: "tcp",
		Addr:    ":8081",
		Timeout: 30 * time.Second,
		TLS: conf.TLS{
			Enabled:  true,
			CertFile: certFile,
			KeyFile:  keyFile,
		},
	}

	server := NewHTTPServer(config, logger, healthHandler, requestResponseHandler)
	if server == nil {
		t.Fatal("Expected server to be created, got nil")
	}

	// Server should be created successfully with TLS
}

func TestNewHTTPServer_TLS_Enabled_InvalidCerts(t *testing.T) {
	logger := log.NewStdLogger(os.Stdout)
	healthHandler := handlers.NewHealthHandler(logger)
	requestResponseHandler := handlers.NewRequestResponseHandler(logger)

	config := &conf.HTTP{
		Network: "tcp",
		Addr:    ":8082",
		Timeout: 30 * time.Second,
		TLS: conf.TLS{
			Enabled:  true,
			CertFile: "nonexistent.crt",
			KeyFile:  "nonexistent.key",
		},
	}

	// This should not panic, but should log an error
	server := NewHTTPServer(config, logger, healthHandler, requestResponseHandler)
	if server == nil {
		t.Fatal("Expected server to be created even with invalid certs, got nil")
	}

	// Server should still be created, but without TLS due to cert loading failure
}

func TestTLS_CertificateLoading(t *testing.T) {
	// Create temporary directory for test certificates
	tempDir, err := os.MkdirTemp("", "cert_loading_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Generate test certificate and key
	certFile := filepath.Join(tempDir, "test.crt")
	keyFile := filepath.Join(tempDir, "test.key")

	if err := generateTestCertificate(certFile, keyFile); err != nil {
		t.Fatalf("Failed to generate test certificate: %v", err)
	}

	tests := []struct {
		name     string
		certFile string
		keyFile  string
		wantErr  bool
	}{
		{
			name:     "valid certificate files",
			certFile: certFile,
			keyFile:  keyFile,
			wantErr:  false,
		},
		{
			name:     "non-existent cert file",
			certFile: filepath.Join(tempDir, "nonexistent.crt"),
			keyFile:  keyFile,
			wantErr:  true,
		},
		{
			name:     "non-existent key file",
			certFile: certFile,
			keyFile:  filepath.Join(tempDir, "nonexistent.key"),
			wantErr:  true,
		},
		{
			name:     "empty cert file path",
			certFile: "",
			keyFile:  keyFile,
			wantErr:  true,
		},
		{
			name:     "empty key file path",
			certFile: certFile,
			keyFile:  "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cert, err := tls.LoadX509KeyPair(tt.certFile, tt.keyFile)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error loading certificate, got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error loading certificate: %v", err)
				}
				if len(cert.Certificate) == 0 {
					t.Errorf("Expected certificate to be loaded, got empty certificate")
				}
			}
		})
	}
}

func TestTLS_ConfigCreation(t *testing.T) {
	// Create temporary directory for test certificates
	tempDir, err := os.MkdirTemp("", "tls_config_creation_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Generate test certificate and key
	certFile := filepath.Join(tempDir, "test.crt")
	keyFile := filepath.Join(tempDir, "test.key")

	if err := generateTestCertificate(certFile, keyFile); err != nil {
		t.Fatalf("Failed to generate test certificate: %v", err)
	}

	tests := []struct {
		name        string
		tlsConfig   conf.TLS
		expectTLS   bool
		expectError bool
	}{
		{
			name: "disabled TLS",
			tlsConfig: conf.TLS{
				Enabled: false,
			},
			expectTLS:   false,
			expectError: false,
		},
		{
			name: "enabled TLS with valid certs",
			tlsConfig: conf.TLS{
				Enabled:  true,
				CertFile: certFile,
				KeyFile:  keyFile,
			},
			expectTLS:   true,
			expectError: false,
		},
		{
			name: "enabled TLS with invalid certs",
			tlsConfig: conf.TLS{
				Enabled:  true,
				CertFile: "nonexistent.crt",
				KeyFile:  "nonexistent.key",
			},
			expectTLS:   false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tlsConfig *tls.Config
			var err error

			if tt.tlsConfig.Enabled {
				cert, loadErr := tls.LoadX509KeyPair(tt.tlsConfig.CertFile, tt.tlsConfig.KeyFile)
				if loadErr != nil {
					err = loadErr
				} else {
					tlsConfig = &tls.Config{
						Certificates: []tls.Certificate{cert},
					}
				}
			}

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if tt.expectTLS {
				if tlsConfig == nil {
					t.Errorf("Expected TLS config to be created, got nil")
				} else {
					if len(tlsConfig.Certificates) != 1 {
						t.Errorf("Expected 1 certificate, got %d", len(tlsConfig.Certificates))
					}
				}
			} else {
				if tlsConfig != nil && !tt.expectError {
					t.Errorf("Expected no TLS config, got %v", tlsConfig)
				}
			}
		})
	}
}

func TestTLS_ServerOptions(t *testing.T) {
	// Create temporary directory for test certificates
	tempDir, err := os.MkdirTemp("", "server_options_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Generate test certificate and key
	certFile := filepath.Join(tempDir, "test.crt")
	keyFile := filepath.Join(tempDir, "test.key")

	if err := generateTestCertificate(certFile, keyFile); err != nil {
		t.Fatalf("Failed to generate test certificate: %v", err)
	}

	logger := log.NewStdLogger(os.Stdout)

	tests := []struct {
		name      string
		tlsConfig conf.TLS
		wantPanic bool
	}{
		{
			name: "valid TLS configuration should not panic",
			tlsConfig: conf.TLS{
				Enabled:  true,
				CertFile: certFile,
				KeyFile:  keyFile,
			},
			wantPanic: false,
		},
		{
			name: "disabled TLS should not panic",
			tlsConfig: conf.TLS{
				Enabled: false,
			},
			wantPanic: false,
		},
		{
			name: "invalid TLS configuration should not panic server creation",
			tlsConfig: conf.TLS{
				Enabled:  true,
				CertFile: "nonexistent.crt",
				KeyFile:  "nonexistent.key",
			},
			wantPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.wantPanic {
						t.Errorf("Unexpected panic: %v", r)
					}
				} else if tt.wantPanic {
					t.Errorf("Expected panic but got none")
				}
			}()

			config := &conf.HTTP{
				Network: "tcp",
				Addr:    ":0", // Use port 0 to let OS assign a free port
				Timeout: 30 * time.Second,
				TLS:     tt.tlsConfig,
			}

			healthHandler := handlers.NewHealthHandler(logger)
			requestResponseHandler := handlers.NewRequestResponseHandler(logger)

			server := NewHTTPServer(config, logger, healthHandler, requestResponseHandler)
			if server == nil {
				t.Errorf("Expected server to be created, got nil")
			}
		})
	}
}

// generateTestCertificate creates a self-signed certificate for testing
func generateTestCertificate(certFile, keyFile string) error {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"Test Org"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"Test City"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1)},
		DNSNames:    []string{"localhost"},
	}

	// Create certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return err
	}

	// Write certificate file
	certOut, err := os.Create(certFile)
	if err != nil {
		return err
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		return err
	}

	// Write private key file
	keyOut, err := os.Create(keyFile)
	if err != nil {
		return err
	}
	defer keyOut.Close()

	privateKeyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return err
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyDER}); err != nil {
		return err
	}

	return nil
}
func TestCreateTLSConfig(t *testing.T) {
	// Create temporary directory for test certificates
	tempDir, err := os.MkdirTemp("", "create_tls_config_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Generate test certificate and key
	certFile := filepath.Join(tempDir, "test.crt")
	keyFile := filepath.Join(tempDir, "test.key")

	if err := generateTestCertificate(certFile, keyFile); err != nil {
		t.Fatalf("Failed to generate test certificate: %v", err)
	}

	logger := log.NewStdLogger(os.Stdout)

	tests := []struct {
		name     string
		tlsConf  conf.TLS
		wantErr  bool
		errMsg   string
		validate func(*tls.Config) error
	}{
		{
			name: "disabled TLS should return error",
			tlsConf: conf.TLS{
				Enabled: false,
			},
			wantErr: true,
			errMsg:  "TLS is not enabled",
		},
		{
			name: "empty cert file should return error",
			tlsConf: conf.TLS{
				Enabled:  true,
				CertFile: "",
				KeyFile:  keyFile,
			},
			wantErr: true,
			errMsg:  "TLS certificate file path is empty",
		},
		{
			name: "empty key file should return error",
			tlsConf: conf.TLS{
				Enabled:  true,
				CertFile: certFile,
				KeyFile:  "",
			},
			wantErr: true,
			errMsg:  "TLS private key file path is empty",
		},
		{
			name: "non-existent cert file should return error",
			tlsConf: conf.TLS{
				Enabled:  true,
				CertFile: "nonexistent.crt",
				KeyFile:  keyFile,
			},
			wantErr: true,
		},
		{
			name: "valid TLS configuration",
			tlsConf: conf.TLS{
				Enabled:  true,
				CertFile: certFile,
				KeyFile:  keyFile,
			},
			wantErr: false,
			validate: func(config *tls.Config) error {
				if len(config.Certificates) != 1 {
					return fmt.Errorf("expected 1 certificate, got %d", len(config.Certificates))
				}
				if config.MinVersion != tls.VersionTLS12 {
					return fmt.Errorf("expected MinVersion TLS 1.2, got %d", config.MinVersion)
				}
				if config.MaxVersion != tls.VersionTLS13 {
					return fmt.Errorf("expected MaxVersion TLS 1.3, got %d", config.MaxVersion)
				}
				if !config.PreferServerCipherSuites {
					return fmt.Errorf("expected PreferServerCipherSuites to be true")
				}
				if config.ClientAuth != tls.NoClientCert {
					return fmt.Errorf("expected ClientAuth to be NoClientCert")
				}
				if len(config.CipherSuites) == 0 {
					return fmt.Errorf("expected cipher suites to be configured")
				}
				if len(config.CurvePreferences) == 0 {
					return fmt.Errorf("expected curve preferences to be configured")
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := createTLSConfig(tt.tlsConf, logger)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("Expected error message %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
					return
				}
				if config == nil {
					t.Errorf("Expected TLS config to be created, got nil")
					return
				}
				if tt.validate != nil {
					if validateErr := tt.validate(config); validateErr != nil {
						t.Errorf("TLS config validation failed: %v", validateErr)
					}
				}
			}
		})
	}
}

func TestTLS_SecurityConfiguration(t *testing.T) {
	// Create temporary directory for test certificates
	tempDir, err := os.MkdirTemp("", "tls_security_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Generate test certificate and key
	certFile := filepath.Join(tempDir, "test.crt")
	keyFile := filepath.Join(tempDir, "test.key")

	if err := generateTestCertificate(certFile, keyFile); err != nil {
		t.Fatalf("Failed to generate test certificate: %v", err)
	}

	logger := log.NewStdLogger(os.Stdout)
	tlsConf := conf.TLS{
		Enabled:  true,
		CertFile: certFile,
		KeyFile:  keyFile,
	}

	config, err := createTLSConfig(tlsConf, logger)
	if err != nil {
		t.Fatalf("Failed to create TLS config: %v", err)
	}

	// Test security configurations
	t.Run("minimum TLS version", func(t *testing.T) {
		if config.MinVersion != tls.VersionTLS12 {
			t.Errorf("Expected minimum TLS version 1.2, got %d", config.MinVersion)
		}
	})

	t.Run("maximum TLS version", func(t *testing.T) {
		if config.MaxVersion != tls.VersionTLS13 {
			t.Errorf("Expected maximum TLS version 1.3, got %d", config.MaxVersion)
		}
	})

	t.Run("cipher suites configured", func(t *testing.T) {
		expectedCiphers := []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		}

		if len(config.CipherSuites) != len(expectedCiphers) {
			t.Errorf("Expected %d cipher suites, got %d", len(expectedCiphers), len(config.CipherSuites))
		}

		for i, expected := range expectedCiphers {
			if i < len(config.CipherSuites) && config.CipherSuites[i] != expected {
				t.Errorf("Expected cipher suite %d at position %d, got %d", expected, i, config.CipherSuites[i])
			}
		}
	})

	t.Run("curve preferences configured", func(t *testing.T) {
		expectedCurves := []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
			tls.CurveP384,
		}

		if len(config.CurvePreferences) != len(expectedCurves) {
			t.Errorf("Expected %d curve preferences, got %d", len(expectedCurves), len(config.CurvePreferences))
		}

		for i, expected := range expectedCurves {
			if i < len(config.CurvePreferences) && config.CurvePreferences[i] != expected {
				t.Errorf("Expected curve %d at position %d, got %d", expected, i, config.CurvePreferences[i])
			}
		}
	})

	t.Run("server cipher suite preference", func(t *testing.T) {
		if !config.PreferServerCipherSuites {
			t.Errorf("Expected PreferServerCipherSuites to be true")
		}
	})

	t.Run("session tickets enabled", func(t *testing.T) {
		if config.SessionTicketsDisabled {
			t.Errorf("Expected session tickets to be enabled")
		}
	})

	t.Run("client authentication", func(t *testing.T) {
		if config.ClientAuth != tls.NoClientCert {
			t.Errorf("Expected ClientAuth to be NoClientCert, got %d", config.ClientAuth)
		}
	})
}

func TestTLS_IntegrationWithServer(t *testing.T) {
	// Create temporary directory for test certificates
	tempDir, err := os.MkdirTemp("", "tls_integration_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Generate test certificate and key
	certFile := filepath.Join(tempDir, "test.crt")
	keyFile := filepath.Join(tempDir, "test.key")

	if err := generateTestCertificate(certFile, keyFile); err != nil {
		t.Fatalf("Failed to generate test certificate: %v", err)
	}

	logger := log.NewStdLogger(os.Stdout)
	healthHandler := handlers.NewHealthHandler(logger)
	requestResponseHandler := handlers.NewRequestResponseHandler(logger)

	tests := []struct {
		name    string
		config  *conf.HTTP
		wantErr bool
	}{
		{
			name: "HTTPS server with valid TLS config",
			config: &conf.HTTP{
				Network: "tcp",
				Addr:    ":0",
				Timeout: 30 * time.Second,
				TLS: conf.TLS{
					Enabled:  true,
					CertFile: certFile,
					KeyFile:  keyFile,
				},
			},
			wantErr: false,
		},
		{
			name: "HTTP server with disabled TLS",
			config: &conf.HTTP{
				Network: "tcp",
				Addr:    ":0",
				Timeout: 30 * time.Second,
				TLS: conf.TLS{
					Enabled: false,
				},
			},
			wantErr: false,
		},
		{
			name: "HTTPS server with invalid TLS config should not fail server creation",
			config: &conf.HTTP{
				Network: "tcp",
				Addr:    ":0",
				Timeout: 30 * time.Second,
				TLS: conf.TLS{
					Enabled:  true,
					CertFile: "nonexistent.crt",
					KeyFile:  "nonexistent.key",
				},
			},
			wantErr: false, // Server creation should not fail, but TLS won't be enabled
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewHTTPServer(tt.config, logger, healthHandler, requestResponseHandler)

			if tt.wantErr {
				if server != nil {
					t.Errorf("Expected server creation to fail, but got server")
				}
			} else {
				if server == nil {
					t.Errorf("Expected server to be created, got nil")
				}
			}
		})
	}
}
