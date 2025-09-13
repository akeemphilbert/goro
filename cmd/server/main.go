package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/env"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	"github.com/akeemphilbert/goro/internal/conf"

	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string = "goro-server"
	// Version is the version of the compiled software.
	Version string = "v1.0.0"
	// flagconf is the config flag.
	flagconf = flag.String("conf", "../../configs", "config path, eg: -conf config.yaml")
)

func newApp(logger log.Logger, hs *http.Server, gs *grpc.Server, config *conf.Server) *kratos.App {
	// Configure shutdown timeout from HTTP config
	var shutdownTimeout time.Duration = 10 * time.Second // default
	if config.HTTP != nil && config.HTTP.ShutdownTimeout > 0 {
		shutdownTimeout = config.HTTP.ShutdownTimeout
	}

	return kratos.New(
		kratos.ID(Name),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			hs,
			gs,
		),
		kratos.StopTimeout(shutdownTimeout),
	)
}

func main() {
	flag.Parse()
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", Name,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)

	c := config.New(
		config.WithSource(
			file.NewSource(*flagconf),
			env.NewSource("GORO_"),
		),
	)
	defer func() {
		if err := c.Close(); err != nil {
			log.Error(err)
		}
	}()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	app, cleanup, err := wireApp(bc.Server, logger)
	if err != nil {
		panic(err)
	}
	defer func() {
		log.Info("Cleaning up resources...")
		cleanup()
		log.Info("Cleanup completed")
	}()

	// Set up signal handling for graceful shutdown
	_, cancel := setupSignalHandling(logger, bc.Server)
	defer cancel()

	// start and wait for stop signal
	log.Info("Starting application...")
	if err := app.Run(); err != nil {
		log.Errorf("Application error: %v", err)
		panic(err)
	}
	log.Info("Application stopped gracefully")
}

// setupSignalHandling configures signal handling for graceful shutdown
func setupSignalHandling(logger log.Logger, config *conf.Server) (context.Context, context.CancelFunc) {
	// Create context for signal handling
	ctx, cancel := context.WithCancel(context.Background())

	// Create channel to receive OS signals
	sigChan := make(chan os.Signal, 1)

	// Register the channel to receive specific signals
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	// Start goroutine to handle signals
	go func() {
		defer signal.Stop(sigChan)

		select {
		case sig := <-sigChan:
			log.Infof("Received signal: %v", sig)
			log.Info("Initiating graceful shutdown...")

			// Get shutdown timeout from configuration
			shutdownTimeout := 10 * time.Second // default
			if config.HTTP != nil && config.HTTP.ShutdownTimeout > 0 {
				shutdownTimeout = config.HTTP.ShutdownTimeout
			}

			log.Infof("Shutdown timeout configured: %v", shutdownTimeout)

			// Cancel the context to signal shutdown
			cancel()

			// Set up forced termination after timeout
			go func() {
				time.Sleep(shutdownTimeout + 5*time.Second) // Add buffer for cleanup
				log.Warn("Forced termination after shutdown timeout")
				os.Exit(1)
			}()

		case <-ctx.Done():
			// Context was cancelled, stop signal handling
			return
		}
	}()

	log.Info("Signal handling configured for SIGTERM and SIGINT")
	return ctx, cancel
}
