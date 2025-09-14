package http

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/akeemphilbert/goro/internal/conf"
	"github.com/akeemphilbert/goro/internal/infrastructure/transport/http/handlers"
	"github.com/akeemphilbert/goro/internal/infrastructure/transport/http/middleware"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer creates a new HTTP server with the given configuration and logger
func NewHTTPServer(c *conf.HTTP, logger log.Logger, healthHandler *handlers.HealthHandler, requestResponseHandler *handlers.RequestResponseHandler, resourceHandler *handlers.ResourceHandler, containerHandler *handlers.ContainerHandler, userHandler *handlers.UserHandler, accountHandler *handlers.AccountHandler) *http.Server {
	var opts = []http.ServerOption{
		http.Address(c.Addr),
		http.Timeout(time.Duration(c.Timeout)),
		http.Middleware(
			recovery.Recovery(),
			middleware.Timeout(time.Duration(c.Timeout)), // Use configured timeout
			middleware.StructuredLogging(logger),
		),
		http.Filter(
			middleware.CORS(),               // Add CORS support
			middleware.ContentNegotiation(), // Add content negotiation for RDF formats
		),
	}

	// Add shutdown timeout if configured
	if c.ShutdownTimeout > 0 {
		log.Infof("Configuring HTTP server shutdown timeout: %v", time.Duration(c.ShutdownTimeout))
		// Note: Kratos uses the timeout option for both request and shutdown timeout
		// We'll handle graceful shutdown at the app level
	}

	// Add TLS support if enabled
	if c.TLS.Enabled {
		tlsConfig, err := createTLSConfig(c.TLS, logger)
		if err != nil {
			log.Errorf("Failed to create TLS configuration: %v", err)
		} else {
			opts = append(opts, http.TLSConfig(tlsConfig))
			log.Infof("HTTPS server enabled with TLS configuration")
		}
	}

	srv := http.NewServer(opts...)

	// Register basic routes
	RegisterRoutes(srv, healthHandler, requestResponseHandler, resourceHandler, containerHandler, userHandler, accountHandler)

	return srv
}

// NewHTTPServerWithoutResourceHandler creates a new HTTP server without resource handler for backward compatibility
func NewHTTPServerWithoutResourceHandler(c *conf.HTTP, logger log.Logger, healthHandler *handlers.HealthHandler, requestResponseHandler *handlers.RequestResponseHandler) *http.Server {
	var opts = []http.ServerOption{
		http.Address(c.Addr),
		http.Timeout(time.Duration(c.Timeout)),
		http.Middleware(
			recovery.Recovery(),
			middleware.Timeout(time.Duration(c.Timeout)), // Use configured timeout
			middleware.StructuredLogging(logger),
		),
		http.Filter(
			middleware.CORS(), // Add CORS support
		),
	}

	// Add shutdown timeout if configured
	if c.ShutdownTimeout > 0 {
		log.Infof("Configuring HTTP server shutdown timeout: %v", time.Duration(c.ShutdownTimeout))
		// Note: Kratos uses the timeout option for both request and shutdown timeout
		// We'll handle graceful shutdown at the app level
	}

	// Add TLS support if enabled
	if c.TLS.Enabled {
		tlsConfig, err := createTLSConfig(c.TLS, logger)
		if err != nil {
			log.Errorf("Failed to create TLS configuration: %v", err)
		} else {
			opts = append(opts, http.TLSConfig(tlsConfig))
			log.Infof("HTTPS server enabled with TLS configuration")
		}
	}

	srv := http.NewServer(opts...)

	// Register basic routes without resource handler
	RegisterBasicRoutes(srv, healthHandler, requestResponseHandler)

	return srv
}

// RegisterBasicRoutes registers basic routes without resource endpoints
func RegisterBasicRoutes(srv *http.Server, healthHandler *handlers.HealthHandler, requestResponseHandler *handlers.RequestResponseHandler) {
	// Health check route using the proper handler
	srv.Route("/health").GET("/", healthHandler.Check)

	// Status route with path parameters and error handling
	srv.Route("/status").GET("/{id}", healthHandler.Status)

	// Request/Response processing demonstration routes
	srv.Route("/demo").GET("/path/{id}", requestResponseHandler.GetWithPathParams)
	srv.Route("/demo").GET("/query", requestResponseHandler.GetWithQueryParams)
	srv.Route("/demo").GET("/json", requestResponseHandler.GetJSONResponse)

	// Basic status route
	srv.Route("/status").GET("/", func(ctx http.Context) error {
		return ctx.JSON(200, map[string]string{
			"service": "running",
		})
	})
}

// RegisterResourceRoutes registers resource storage endpoints
func RegisterResourceRoutes(srv *http.Server, resourceHandler *handlers.ResourceHandler) {
	// Resource collection endpoints
	resourceRoute := srv.Route("/resources")

	// Collection operations
	resourceRoute.POST("/", resourceHandler.PostResource)
	resourceRoute.OPTIONS("/", resourceHandler.OptionsResource)

	// Individual resource operations
	resourceRoute.GET("/{id}", resourceHandler.GetResource)
	resourceRoute.PUT("/{id}", resourceHandler.PutResource)
	resourceRoute.DELETE("/{id}", resourceHandler.DeleteResource)
	resourceRoute.HEAD("/{id}", resourceHandler.HeadResource)
	resourceRoute.OPTIONS("/{id}", resourceHandler.OptionsResource)
}

// RegisterContainerRoutes registers container management endpoints
func RegisterContainerRoutes(srv *http.Server, containerHandler *handlers.ContainerHandler) {
	// Container collection endpoints
	containerRoute := srv.Route("/containers")

	// Collection operations - use PostResource for creating resources in containers
	containerRoute.POST("/", containerHandler.PostResource)
	containerRoute.OPTIONS("/", containerHandler.OptionsContainer)

	// Individual container operations
	containerRoute.GET("/{id}", containerHandler.GetContainer)
	containerRoute.PUT("/{id}", containerHandler.PutContainer)
	containerRoute.DELETE("/{id}", containerHandler.DeleteContainer)
	containerRoute.HEAD("/{id}", containerHandler.HeadContainer)
	containerRoute.OPTIONS("/{id}", containerHandler.OptionsContainer)

	// Container member operations - use PostResource for adding members
	containerRoute.POST("/{id}/members", containerHandler.PostResource)
}

// RegisterRoutes registers basic routes on the HTTP server
func RegisterRoutes(srv *http.Server, healthHandler *handlers.HealthHandler, requestResponseHandler *handlers.RequestResponseHandler, resourceHandler *handlers.ResourceHandler, containerHandler *handlers.ContainerHandler, userHandler *handlers.UserHandler, accountHandler *handlers.AccountHandler) {
	// Health check route using the proper handler
	srv.Route("/health").GET("/", healthHandler.Check)

	// Status route with path parameters and error handling
	srv.Route("/status").GET("/{id}", healthHandler.Status)

	// Request/Response processing demonstration routes
	srv.Route("/demo").GET("/path/{id}", requestResponseHandler.GetWithPathParams)
	srv.Route("/demo").GET("/query", requestResponseHandler.GetWithQueryParams)
	srv.Route("/demo").GET("/json", requestResponseHandler.GetJSONResponse)

	// Basic status route
	srv.Route("/status").GET("/", func(ctx http.Context) error {
		return ctx.JSON(200, map[string]string{
			"service": "running",
		})
	})

	// Resource storage endpoints
	RegisterResourceRoutes(srv, resourceHandler)

	// Container endpoints
	RegisterContainerRoutes(srv, containerHandler)

	// User management endpoints (only if handlers are provided)
	if userHandler != nil {
		RegisterUserRoutes(srv, userHandler)
	}
	if accountHandler != nil {
		RegisterAccountRoutes(srv, accountHandler)
	}
}

// RegisterRouteGroups registers route groups for organized routing
func RegisterRouteGroups(srv *http.Server) {
	// API v1 route group
	v1 := srv.Route("/api/v1")
	v1.GET("/users", func(ctx http.Context) error {
		return ctx.JSON(200, map[string]string{"message": "users endpoint"})
	})
	v1.GET("/resources", func(ctx http.Context) error {
		return ctx.JSON(200, map[string]string{"message": "resources endpoint"})
	})

	// Admin route group
	admin := srv.Route("/admin")
	admin.GET("/health", func(ctx http.Context) error {
		return ctx.JSON(200, map[string]string{"status": "admin healthy"})
	})
}

// RegisterParameterRoutes registers routes with path parameters
func RegisterParameterRoutes(srv *http.Server) {
	// Routes with path parameters
	srv.Route("/users").GET("/{id}", func(ctx http.Context) error {
		vars := ctx.Vars()
		id := ""
		if len(vars["id"]) > 0 {
			id = vars["id"][0]
		}
		return ctx.JSON(200, map[string]string{
			"user_id": id,
			"message": "user details",
		})
	})

	srv.Route("/resources").GET("/{id}/items/{item_id}", func(ctx http.Context) error {
		vars := ctx.Vars()
		resourceID := ""
		itemID := ""
		if len(vars["id"]) > 0 {
			resourceID = vars["id"][0]
		}
		if len(vars["item_id"]) > 0 {
			itemID = vars["item_id"][0]
		}
		return ctx.JSON(200, map[string]string{
			"resource_id": resourceID,
			"item_id":     itemID,
			"message":     "nested resource",
		})
	})

	// Wildcard route
	srv.Route("/files").GET("/{path:.*}", func(ctx http.Context) error {
		vars := ctx.Vars()
		path := ""
		if len(vars["path"]) > 0 {
			path = vars["path"][0]
		}
		return ctx.JSON(200, map[string]string{
			"file_path": path,
			"message":   "file access",
		})
	})
}

// RegisterMethodSpecificRoutes registers routes that demonstrate all HTTP method support
func RegisterMethodSpecificRoutes(srv *http.Server) {
	route := srv.Route("/api/resource")

	// GET method
	route.GET("/", func(ctx http.Context) error {
		return ctx.JSON(200, map[string]string{
			"method":  "GET",
			"message": "success",
		})
	})

	// POST method
	route.POST("/", func(ctx http.Context) error {
		return ctx.JSON(201, map[string]string{
			"method":  "POST",
			"message": "created",
		})
	})

	// PUT method
	route.PUT("/", func(ctx http.Context) error {
		return ctx.JSON(200, map[string]string{
			"method":  "PUT",
			"message": "updated",
		})
	})

	// DELETE method
	route.DELETE("/", func(ctx http.Context) error {
		return ctx.JSON(200, map[string]string{
			"method":  "DELETE",
			"message": "deleted",
		})
	})

	// PATCH method
	route.PATCH("/", func(ctx http.Context) error {
		return ctx.JSON(200, map[string]string{
			"method":  "PATCH",
			"message": "patched",
		})
	})

	// HEAD method - should return same headers as GET but no body
	route.HEAD("/", func(ctx http.Context) error {
		// Set the same headers as GET would
		ctx.Response().Header().Set("Content-Type", "application/json")
		ctx.Response().WriteHeader(200)
		return nil
	})

	// OPTIONS method - returns supported methods
	route.OPTIONS("/", func(ctx http.Context) error {
		return ctx.JSON(200, map[string]interface{}{
			"methods": []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"},
		})
	})
}

// RegisterOPTIONSTestRoutes registers routes for testing OPTIONS method discovery
func RegisterOPTIONSTestRoutes(srv *http.Server) {
	// Full CRUD resource with all methods
	crudRoute := srv.Route("/api/crud-resource")
	crudRoute.GET("/", func(ctx http.Context) error {
		return ctx.JSON(200, map[string]string{"method": "GET"})
	})
	crudRoute.POST("/", func(ctx http.Context) error {
		return ctx.JSON(201, map[string]string{"method": "POST"})
	})
	crudRoute.PUT("/", func(ctx http.Context) error {
		return ctx.JSON(200, map[string]string{"method": "PUT"})
	})
	crudRoute.DELETE("/", func(ctx http.Context) error {
		return ctx.JSON(200, map[string]string{"method": "DELETE"})
	})
	crudRoute.PATCH("/", func(ctx http.Context) error {
		return ctx.JSON(200, map[string]string{"method": "PATCH"})
	})
	crudRoute.HEAD("/", func(ctx http.Context) error {
		ctx.Response().Header().Set("Content-Type", "application/json")
		ctx.Response().WriteHeader(200)
		return nil
	})
	crudRoute.OPTIONS("/", func(ctx http.Context) error {
		return ctx.JSON(200, map[string]interface{}{
			"methods": []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"},
		})
	})

	// Read-only resource
	readOnlyRoute := srv.Route("/api/readonly-resource")
	readOnlyRoute.GET("/", func(ctx http.Context) error {
		return ctx.JSON(200, map[string]string{"method": "GET"})
	})
	readOnlyRoute.HEAD("/", func(ctx http.Context) error {
		ctx.Response().Header().Set("Content-Type", "application/json")
		ctx.Response().WriteHeader(200)
		return nil
	})
	readOnlyRoute.OPTIONS("/", func(ctx http.Context) error {
		return ctx.JSON(200, map[string]interface{}{
			"methods": []string{"GET", "HEAD", "OPTIONS"},
		})
	})

	// Write-only resource
	writeOnlyRoute := srv.Route("/api/writeonly-resource")
	writeOnlyRoute.POST("/", func(ctx http.Context) error {
		return ctx.JSON(201, map[string]string{"method": "POST"})
	})
	writeOnlyRoute.PUT("/", func(ctx http.Context) error {
		return ctx.JSON(200, map[string]string{"method": "PUT"})
	})
	writeOnlyRoute.OPTIONS("/", func(ctx http.Context) error {
		return ctx.JSON(200, map[string]interface{}{
			"methods": []string{"POST", "PUT", "OPTIONS"},
		})
	})
}

// RegisterParameterizedOPTIONSRoutes registers routes with path parameters for OPTIONS testing
func RegisterParameterizedOPTIONSRoutes(srv *http.Server) {
	// Collection endpoint
	collectionRoute := srv.Route("/api/items")
	collectionRoute.GET("/", func(ctx http.Context) error {
		return ctx.JSON(200, map[string]string{"method": "GET", "type": "collection"})
	})
	collectionRoute.POST("/", func(ctx http.Context) error {
		return ctx.JSON(201, map[string]string{"method": "POST", "type": "collection"})
	})
	collectionRoute.HEAD("/", func(ctx http.Context) error {
		ctx.Response().Header().Set("Content-Type", "application/json")
		ctx.Response().WriteHeader(200)
		return nil
	})
	collectionRoute.OPTIONS("/", func(ctx http.Context) error {
		return ctx.JSON(200, map[string]interface{}{
			"methods": []string{"GET", "POST", "HEAD", "OPTIONS"},
		})
	})

	// Item endpoint with ID parameter
	itemRoute := srv.Route("/api/items")
	itemRoute.GET("/{id}", func(ctx http.Context) error {
		vars := ctx.Vars()
		id := ""
		if len(vars["id"]) > 0 {
			id = vars["id"][0]
		}
		return ctx.JSON(200, map[string]string{"method": "GET", "type": "item", "id": id})
	})
	itemRoute.PUT("/{id}", func(ctx http.Context) error {
		vars := ctx.Vars()
		id := ""
		if len(vars["id"]) > 0 {
			id = vars["id"][0]
		}
		return ctx.JSON(200, map[string]string{"method": "PUT", "type": "item", "id": id})
	})
	itemRoute.DELETE("/{id}", func(ctx http.Context) error {
		vars := ctx.Vars()
		id := ""
		if len(vars["id"]) > 0 {
			id = vars["id"][0]
		}
		return ctx.JSON(200, map[string]string{"method": "DELETE", "type": "item", "id": id})
	})
	itemRoute.HEAD("/{id}", func(ctx http.Context) error {
		ctx.Response().Header().Set("Content-Type", "application/json")
		ctx.Response().WriteHeader(200)
		return nil
	})
	itemRoute.OPTIONS("/{id}", func(ctx http.Context) error {
		return ctx.JSON(200, map[string]interface{}{
			"methods": []string{"GET", "PUT", "DELETE", "HEAD", "OPTIONS"},
		})
	})
}

// RegisterHEADTestRoutes registers routes for testing HEAD method support
func RegisterHEADTestRoutes(srv *http.Server) {
	// Basic HEAD test route
	headRoute := srv.Route("/api/head-test")

	// GET handler that sets custom headers
	headRoute.GET("/", func(ctx http.Context) error {
		// Set custom headers that should be present in HEAD response
		ctx.Response().Header().Set("X-Resource-Type", "test-resource")
		ctx.Response().Header().Set("X-Version", "1.0")

		return ctx.JSON(200, map[string]interface{}{
			"message": "This is a test resource",
			"data":    []string{"item1", "item2", "item3"},
		})
	})

	// HEAD handler that returns same headers as GET but no body
	headRoute.HEAD("/", func(ctx http.Context) error {
		// Set the same headers as GET
		ctx.Response().Header().Set("X-Resource-Type", "test-resource")
		ctx.Response().Header().Set("X-Version", "1.0")
		ctx.Response().Header().Set("Content-Type", "application/json")

		// Write status code without body
		ctx.Response().WriteHeader(200)
		return nil
	})

	// GET handler with path parameter
	headRoute.GET("/{id}", func(ctx http.Context) error {
		vars := ctx.Vars()
		id := ""
		if len(vars["id"]) > 0 {
			id = vars["id"][0]
		}

		// Simulate not found
		if id == "notfound" {
			return ctx.JSON(404, map[string]string{"error": "Resource not found"})
		}

		// Set custom headers
		ctx.Response().Header().Set("X-Resource-Type", "test-resource")
		ctx.Response().Header().Set("X-Resource-ID", id)

		return ctx.JSON(200, map[string]interface{}{
			"id":      id,
			"message": "Resource found",
		})
	})

	// HEAD handler with path parameter
	headRoute.HEAD("/{id}", func(ctx http.Context) error {
		vars := ctx.Vars()
		id := ""
		if len(vars["id"]) > 0 {
			id = vars["id"][0]
		}

		// Simulate not found
		if id == "notfound" {
			ctx.Response().Header().Set("Content-Type", "application/json")
			ctx.Response().WriteHeader(404)
			return nil
		}

		// Set the same headers as GET
		ctx.Response().Header().Set("X-Resource-Type", "test-resource")
		ctx.Response().Header().Set("X-Resource-ID", id)
		ctx.Response().Header().Set("Content-Type", "application/json")

		ctx.Response().WriteHeader(200)
		return nil
	})
}

// RegisterContentTypeTestRoutes registers routes for testing different content types with HEAD
func RegisterContentTypeTestRoutes(srv *http.Server) {
	contentRoute := srv.Route("/api/content")

	// JSON content
	contentRoute.GET("/json", func(ctx http.Context) error {
		return ctx.JSON(200, map[string]string{"type": "json", "message": "JSON response"})
	})
	contentRoute.HEAD("/json", func(ctx http.Context) error {
		ctx.Response().Header().Set("Content-Type", "application/json")
		ctx.Response().WriteHeader(200)
		return nil
	})

	// Plain text content
	contentRoute.GET("/text", func(ctx http.Context) error {
		ctx.Response().Header().Set("Content-Type", "text/plain")
		ctx.Response().WriteHeader(200)
		ctx.Response().Write([]byte("This is plain text content"))
		return nil
	})
	contentRoute.HEAD("/text", func(ctx http.Context) error {
		ctx.Response().Header().Set("Content-Type", "text/plain")
		ctx.Response().WriteHeader(200)
		return nil
	})

	// XML content
	contentRoute.GET("/xml", func(ctx http.Context) error {
		ctx.Response().Header().Set("Content-Type", "application/xml")
		ctx.Response().WriteHeader(200)
		ctx.Response().Write([]byte(`<?xml version="1.0"?><root><message>XML response</message></root>`))
		return nil
	})
	contentRoute.HEAD("/xml", func(ctx http.Context) error {
		ctx.Response().Header().Set("Content-Type", "application/xml")
		ctx.Response().WriteHeader(200)
		return nil
	})
}

// RegisterErrorTestRoutes registers routes for testing HEAD method with error responses
func RegisterErrorTestRoutes(srv *http.Server) {
	errorRoute := srv.Route("/api/error")

	// 404 error
	errorRoute.GET("/404", func(ctx http.Context) error {
		return ctx.JSON(404, map[string]string{"error": "Not found"})
	})
	errorRoute.HEAD("/404", func(ctx http.Context) error {
		ctx.Response().Header().Set("Content-Type", "application/json")
		ctx.Response().WriteHeader(404)
		return nil
	})

	// 500 error
	errorRoute.GET("/500", func(ctx http.Context) error {
		return ctx.JSON(500, map[string]string{"error": "Internal server error"})
	})
	errorRoute.HEAD("/500", func(ctx http.Context) error {
		ctx.Response().Header().Set("Content-Type", "application/json")
		ctx.Response().WriteHeader(500)
		return nil
	})

	// 400 error
	errorRoute.GET("/400", func(ctx http.Context) error {
		return ctx.JSON(400, map[string]string{"error": "Bad request"})
	})
	errorRoute.HEAD("/400", func(ctx http.Context) error {
		ctx.Response().Header().Set("Content-Type", "application/json")
		ctx.Response().WriteHeader(400)
		return nil
	})
}

// createTLSConfig creates a TLS configuration from the provided TLS settings
func createTLSConfig(tlsConf conf.TLS, logger log.Logger) (*tls.Config, error) {
	if !tlsConf.Enabled {
		return nil, fmt.Errorf("TLS is not enabled")
	}

	// Validate certificate files exist and are readable
	if tlsConf.CertFile == "" {
		return nil, fmt.Errorf("TLS certificate file path is empty")
	}
	if tlsConf.KeyFile == "" {
		return nil, fmt.Errorf("TLS private key file path is empty")
	}

	// Load the certificate and private key
	cert, err := tls.LoadX509KeyPair(tlsConf.CertFile, tlsConf.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS certificate and key: %w", err)
	}

	// Create TLS configuration with security best practices
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},

		// Security configurations
		MinVersion: tls.VersionTLS12, // Minimum TLS 1.2
		MaxVersion: tls.VersionTLS13, // Maximum TLS 1.3

		// Cipher suites for TLS 1.2 (TLS 1.3 cipher suites are not configurable)
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		},

		// Prefer server cipher suites
		PreferServerCipherSuites: true,

		// Curve preferences
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
			tls.CurveP384,
		},

		// Enable session tickets for performance
		SessionTicketsDisabled: false,

		// Client authentication (can be configured later if needed)
		ClientAuth: tls.NoClientCert,
	}

	log.Infof("TLS configuration created successfully with certificate from %s", tlsConf.CertFile)
	return tlsConfig, nil
}

// RegisterUserRoutes registers user management routes
func RegisterUserRoutes(srv *http.Server, userHandler *handlers.UserHandler) {
	// User registration and management
	srv.Route("/api/v1/users").POST("/register", userHandler.RegisterUser)
	srv.Route("/api/v1/users").GET("/{id}", userHandler.GetUser)
	srv.Route("/api/v1/users").PUT("/{id}/profile", userHandler.UpdateProfile)
	srv.Route("/api/v1/users").DELETE("/{id}", userHandler.DeleteAccount)

	// WebID endpoints
	srv.Route("/api/v1/users").GET("/{id}/webid", userHandler.GetWebID)
}

// RegisterAccountRoutes registers account management routes
func RegisterAccountRoutes(srv *http.Server, accountHandler *handlers.AccountHandler) {
	// Account management
	srv.Route("/api/v1/accounts").POST("/", accountHandler.CreateAccount)
	srv.Route("/api/v1/accounts").GET("/{id}", accountHandler.GetAccount)
	srv.Route("/api/v1/accounts").PUT("/{id}", accountHandler.UpdateAccount)

	// Invitation management
	srv.Route("/api/v1/accounts").POST("/{id}/invitations", accountHandler.InviteUser)
	srv.Route("/api/v1/accounts").GET("/{id}/invitations", accountHandler.ListInvitations)
	srv.Route("/api/v1/invitations").POST("/{token}/accept", accountHandler.AcceptInvitation)

	// Member management
	srv.Route("/api/v1/accounts").GET("/{id}/members", accountHandler.ListMembers)
	srv.Route("/api/v1/accounts").PUT("/{id}/members/{userId}/role", accountHandler.UpdateMemberRole)
	srv.Route("/api/v1/accounts").DELETE("/{id}/members/{userId}", accountHandler.RemoveMember)
}
