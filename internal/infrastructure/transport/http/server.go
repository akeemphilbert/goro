package http

import (
	"crypto/tls"
	"time"

	"github.com/akeemphilbert/goro/internal/conf"
	"github.com/akeemphilbert/goro/internal/infrastructure/transport/http/handlers"
	"github.com/akeemphilbert/goro/internal/infrastructure/transport/http/middleware"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer creates a new HTTP server with the given configuration and logger
func NewHTTPServer(c *conf.HTTP, logger log.Logger, healthHandler *handlers.HealthHandler, requestResponseHandler *handlers.RequestResponseHandler) *http.Server {
	var opts = []http.ServerOption{
		http.Address(c.Addr),
		http.Timeout(c.Timeout),
		http.Middleware(
			recovery.Recovery(),
			middleware.Timeout(30*time.Second), // 30 second timeout for requests
			middleware.StructuredLogging(logger),
		),
		http.Filter(
			middleware.CORS(), // Add CORS support
		),
	}

	// Add TLS support if enabled
	if c.TLS.Enabled {
		cert, err := tls.LoadX509KeyPair(c.TLS.CertFile, c.TLS.KeyFile)
		if err != nil {
			log.Errorf("Failed to load TLS certificates: %v", err)
		} else {
			tlsConfig := &tls.Config{
				Certificates: []tls.Certificate{cert},
			}
			opts = append(opts, http.TLSConfig(tlsConfig))
		}
	}

	srv := http.NewServer(opts...)

	// Register basic routes
	RegisterRoutes(srv, healthHandler, requestResponseHandler)

	return srv
}

// RegisterRoutes registers basic routes on the HTTP server
func RegisterRoutes(srv *http.Server, healthHandler *handlers.HealthHandler, requestResponseHandler *handlers.RequestResponseHandler) {
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
