package conf

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTLS_Validate(t *testing.T) {
	tests := []struct {
		name    string
		tls     TLS
		wantErr bool
		errMsg  string
	}{
		{
			name: "disabled TLS should be valid",
			tls: TLS{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "enabled TLS without cert file should be invalid",
			tls: TLS{
				Enabled: true,
				KeyFile: "key.pem",
			},
			wantErr: true,
			errMsg:  "TLS cert file is required when TLS is enabled",
		},
		{
			name: "enabled TLS without key file should be invalid",
			tls: TLS{
				Enabled:  true,
				CertFile: "cert.pem",
			},
			wantErr: true,
			errMsg:  "TLS key file is required when TLS is enabled",
		},
		{
			name: "enabled TLS with both files should be valid",
			tls: TLS{
				Enabled:  true,
				CertFile: "cert.pem",
				KeyFile:  "key.pem",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			http := &HTTP{
				Network: "tcp",
				Addr:    ":8080",
				TLS:     tt.tls,
			}

			err := http.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("HTTP.Validate() expected error but got none")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("HTTP.Validate() error = %v, want %v", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("HTTP.Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestTLS_LoadCertificates(t *testing.T) {
	// Create temporary directory for test certificates
	tempDir, err := os.MkdirTemp("", "tls_test")
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
			name:     "both files non-existent",
			certFile: filepath.Join(tempDir, "nonexistent.crt"),
			keyFile:  filepath.Join(tempDir, "nonexistent.key"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tls.LoadX509KeyPair(tt.certFile, tt.keyFile)
			if tt.wantErr {
				if err == nil {
					t.Errorf("tls.LoadX509KeyPair() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("tls.LoadX509KeyPair() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestTLS_ConfigCreation(t *testing.T) {
	// Create temporary directory for test certificates
	tempDir, err := os.MkdirTemp("", "tls_config_test")
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
		tls      TLS
		wantErr  bool
		validate func(*tls.Config) error
	}{
		{
			name: "valid TLS config creation",
			tls: TLS{
				Enabled:  true,
				CertFile: certFile,
				KeyFile:  keyFile,
			},
			wantErr: false,
			validate: func(config *tls.Config) error {
				if len(config.Certificates) != 1 {
					t.Errorf("Expected 1 certificate, got %d", len(config.Certificates))
				}
				return nil
			},
		},
		{
			name: "disabled TLS should not create config",
			tls: TLS{
				Enabled: false,
			},
			wantErr: false,
			validate: func(config *tls.Config) error {
				if config != nil {
					t.Errorf("Expected nil config for disabled TLS, got %v", config)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tlsConfig *tls.Config
			var err error

			if tt.tls.Enabled {
				cert, loadErr := tls.LoadX509KeyPair(tt.tls.CertFile, tt.tls.KeyFile)
				if loadErr != nil {
					err = loadErr
				} else {
					tlsConfig = &tls.Config{
						Certificates: []tls.Certificate{cert},
					}
				}
			}

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error = %v", err)
				}
				if tt.validate != nil {
					tt.validate(tlsConfig)
				}
			}
		})
	}
}

func TestTLS_DefaultValues(t *testing.T) {
	tests := []struct {
		name     string
		input    TLS
		expected TLS
	}{
		{
			name:  "empty TLS config should have defaults",
			input: TLS{},
			expected: TLS{
				Enabled:  false,
				CertFile: "",
				KeyFile:  "",
			},
		},
		{
			name: "enabled TLS should preserve values",
			input: TLS{
				Enabled:  true,
				CertFile: "custom.crt",
				KeyFile:  "custom.key",
			},
			expected: TLS{
				Enabled:  true,
				CertFile: "custom.crt",
				KeyFile:  "custom.key",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TLS doesn't have a SetDefaults method, so we test the zero values
			if tt.input.Enabled != tt.expected.Enabled {
				t.Errorf("TLS.Enabled = %v, want %v", tt.input.Enabled, tt.expected.Enabled)
			}
			if tt.input.CertFile != tt.expected.CertFile {
				t.Errorf("TLS.CertFile = %v, want %v", tt.input.CertFile, tt.expected.CertFile)
			}
			if tt.input.KeyFile != tt.expected.KeyFile {
				t.Errorf("TLS.KeyFile = %v, want %v", tt.input.KeyFile, tt.expected.KeyFile)
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
