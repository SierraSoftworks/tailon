package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"tailscale.com/tsnet"

	"github.com/sierrasoftworks/tailon/pkg/api"
	"github.com/sierrasoftworks/tailon/pkg/apps"
	"github.com/sierrasoftworks/tailon/pkg/config"
	"github.com/sierrasoftworks/tailon/pkg/ui"
)

var (
	configFile string
	verbose    bool
)

var Version = "dev"

var rootCmd = &cobra.Command{
	Use:   "tailon",
	Short: "A web service for managing applications over Tailscale",
	Long: `tailon is a web service that allows you to manage and monitor applications
over your Tailscale network. It provides APIs to start, stop, and stream logs
from configured applications.`,
	Run:     runServer,
	Version: Version,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "config.yaml", "config file")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose logging")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		logrus.WithError(err).Fatal("Failed to execute command")
	}
}

func runServer(cmd *cobra.Command, args []string) {
	// Configure logging
	if verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}
	logrus.SetFormatter(&logrus.TextFormatter{})

	// Load configuration
	cfg, err := config.Load(configFile)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to load configuration")
	}

	logrus.WithField("config", configFile).Info("Configuration loaded")

	// Validate that at least one server is enabled
	if !cfg.Tailscale.Enabled && cfg.Listen == "" {
		logrus.Fatal("At least one server must be enabled. Please either:\n" +
			"  - Enable Tailscale integration by setting 'tailscale.enabled: true' in your config, or\n" +
			"  - Configure a local HTTP server by setting 'listen: \"localhost:8080\"' (or similar) in your config")
	}

	// Security warning for non-localhost bindings
	if cfg.Listen != "" && !isLocalhostBinding(cfg.Listen) {
		logrus.Warn("WARNING: Your 'listen' address is not bound to localhost. " +
			"This will allow ANYONE with network access to your machine to control your applications. " +
			"For security, consider using 'localhost:PORT' or '127.0.0.1:PORT' instead, " +
			"or rely on Tailscale integration for secure remote access.")
	}

	// Create application manager
	appManager := apps.NewManager(cfg.Applications)

	// Create servers
	var apiServer *api.Server
	var uiServer *ui.Server
	var mainRouter *mux.Router

	// Start Tailscale server if enabled
	var tailscaleServer *http.Server
	if cfg.Tailscale.Enabled {
		// Setup Tailscale server
		tsServer := &tsnet.Server{
			Hostname:  cfg.Tailscale.Name,
			Dir:       cfg.Tailscale.StateDir,
			Ephemeral: cfg.Tailscale.Ephemeral,
		}

		// Get the LocalClient for user context
		localClient, err := tsServer.LocalClient()
		if err != nil {
			logrus.WithError(err).Fatal("Failed to get Tailscale LocalClient")
		}

		// Create servers with Tailscale LocalClient for user context
		apiServer = api.NewServerWithTailscale(appManager, localClient)
		uiServer = ui.NewServer(appManager)

		// Create main router and mount sub-routers
		mainRouter = mux.NewRouter()
		mainRouter.PathPrefix("/api/").Handler(apiServer.Routes())
		mainRouter.PathPrefix("/docs/").Handler(uiServer.Routes())
		mainRouter.PathPrefix("/").Handler(uiServer.Routes())

		// Create HTTP server for Tailscale
		tailscaleServer = &http.Server{
			Handler:      mainRouter,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
		}

		// Start Tailscale server
		go func() {
			var listener net.Listener
			var err error

			// Try to listen on HTTPS port first (Tailscale will auto-configure HTTPS if available)
			listener, err = tsServer.Listen("tcp", ":443")
			if err != nil {
				// Fallback to HTTP
				listener, err = tsServer.Listen("tcp", ":80")
				if err != nil {
					logrus.WithError(err).Fatal("Failed to create Tailscale listener")
				}
				logrus.Info("Starting HTTP server on Tailscale network at :80")
			} else {
				logrus.Info("Starting HTTPS server on Tailscale network at :443")
			}

			if err := tailscaleServer.Serve(listener); err != nil && err != http.ErrServerClosed {
				logrus.WithError(err).Fatal("Tailscale server failed")
			}
		}()
	} else {
		logrus.Info("Tailscale integration disabled")

		// Create servers without Tailscale (anonymous users only)
		apiServer = api.NewServer(appManager)
		uiServer = ui.NewServer(appManager)

		// Create main router and mount sub-routers
		mainRouter = mux.NewRouter()
		mainRouter.PathPrefix("/api/").Handler(apiServer.Routes())
		mainRouter.PathPrefix("/docs/").Handler(uiServer.Routes())
		mainRouter.PathPrefix("/").Handler(uiServer.Routes())
	}

	// Also listen on local interface if configured
	var localServer *http.Server
	if cfg.Listen != "" {
		localServer = &http.Server{
			Addr:         cfg.Listen,
			Handler:      mainRouter,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
		}

		go func() {
			logrus.WithField("addr", cfg.Listen).Info("Starting local HTTP server")
			if err := localServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logrus.WithError(err).Fatal("Local HTTP server failed")
			}
		}()
	} else {
		logrus.Info("Local HTTP server disabled")
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logrus.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown both servers if they exist
	if tailscaleServer != nil {
		if err := tailscaleServer.Shutdown(ctx); err != nil {
			logrus.WithError(err).Error("Tailscale server shutdown failed")
		}
	}

	if localServer != nil {
		if err := localServer.Shutdown(ctx); err != nil {
			logrus.WithError(err).Error("Local server shutdown failed")
		}
	}

	logrus.Info("Server stopped")
}

// isLocalhostBinding checks if the given address is bound to localhost or 127.0.0.1
func isLocalhostBinding(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		// If we can't parse it, assume it's not safe
		return false
	}

	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}
