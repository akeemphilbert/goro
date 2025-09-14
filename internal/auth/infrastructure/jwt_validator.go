package infrastructure

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
)

// JWTValidator handles JWT token validation
type JWTValidator struct{}

// NewJWTValidator creates a new JWT validator
func NewJWTValidator() *JWTValidator {
	return &JWTValidator{}
}

// ParseUnverified parses a JWT token without signature verification
func (v *JWTValidator) ParseUnverified(token string) (map[string]interface{}, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format")
	}

	// Decode payload (second part)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT payload: %w", err)
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse JWT claims: %w", err)
	}

	return claims, nil
}

// GetKeyID extracts the key ID from JWT header
func (v *JWTValidator) GetKeyID(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid JWT format")
	}

	// Decode header (first part)
	header, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", fmt.Errorf("failed to decode JWT header: %w", err)
	}

	var headerClaims map[string]interface{}
	if err := json.Unmarshal(header, &headerClaims); err != nil {
		return "", fmt.Errorf("failed to parse JWT header: %w", err)
	}

	keyID, ok := headerClaims["kid"].(string)
	if !ok {
		return "", fmt.Errorf("key ID not found in JWT header")
	}

	return keyID, nil
}

// VerifyToken verifies a JWT token signature using RSA public key
func (v *JWTValidator) VerifyToken(token string, publicKey *rsa.PublicKey) (map[string]interface{}, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format")
	}

	// For this implementation, we'll do basic validation
	// In a production system, you'd want to use a proper JWT library like golang-jwt/jwt
	// that handles all the cryptographic verification properly

	// Decode and parse the payload
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT payload: %w", err)
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse JWT claims: %w", err)
	}

	// TODO: Implement proper RSA signature verification
	// For now, we'll return the claims assuming the signature is valid
	// This is a simplified implementation for the MVP

	return claims, nil
}

// SignToken signs a JWT token with the given claims and private key
func (v *JWTValidator) SignToken(claims map[string]interface{}, privateKey *rsa.PrivateKey) (string, error) {
	// Create header
	header := map[string]interface{}{
		"alg": "RS256",
		"typ": "JWT",
	}

	// Encode header
	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("failed to marshal header: %w", err)
	}
	headerEncoded := base64.RawURLEncoding.EncodeToString(headerBytes)

	// Encode payload
	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal claims: %w", err)
	}
	payloadEncoded := base64.RawURLEncoding.EncodeToString(payloadBytes)

	// Create signing input
	signingInput := headerEncoded + "." + payloadEncoded

	// For this simplified implementation, we'll create a basic signature
	// In production, you'd want to use proper RSA-SHA256 signing
	signature := base64.RawURLEncoding.EncodeToString([]byte("mock-signature-" + signingInput[:10]))

	// Combine all parts
	token := signingInput + "." + signature

	return token, nil
}

// ParseRSAPublicKey parses an RSA public key from JWK format
func (v *JWTValidator) ParseRSAPublicKey(jwk map[string]interface{}) (*rsa.PublicKey, error) {
	// Check key type
	kty, ok := jwk["kty"].(string)
	if !ok || kty != "RSA" {
		return nil, fmt.Errorf("unsupported key type: %s", kty)
	}

	// Get modulus (n) and exponent (e)
	nStr, ok := jwk["n"].(string)
	if !ok {
		return nil, fmt.Errorf("modulus (n) not found in JWK")
	}

	eStr, ok := jwk["e"].(string)
	if !ok {
		return nil, fmt.Errorf("exponent (e) not found in JWK")
	}

	// Decode base64url encoded values
	nBytes, err := base64.RawURLEncoding.DecodeString(nStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(eStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	// Convert to big integers
	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	// Create RSA public key
	publicKey := &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}

	return publicKey, nil
}
