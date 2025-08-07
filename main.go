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

	"github.com/sierrasoftworks/tail-on/pkg/api"
	"github.com/sierrasoftworks/tail-on/pkg/apps"
	"github.com/sierrasoftworks/tail-on/pkg/config"
	"github.com/sierrasoftworks/tail-on/pkg/ui"
)

var (
	configFile string
	verbose    bool
)

var rootCmd = &cobra.Command{
	Use:   "tail-on",
	Short: "A web service for managing applications over Tailscale",
	Long: `tail-on is a web service that allows you to manage and monitor applications
over your Tailscale network. It provides APIs to start, stop, and stream logs
from configured applications.`,
	Run: runServer,
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
	logrus.SetFormatter(&logrus.JSONFormatter{})

	// Load configuration
	cfg, err := config.Load(configFile)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to load configuration")
	}

	logrus.WithField("config", configFile).Info("Configuration loaded")

	// Validate that at least one server is enabled
	if !cfg.Tailscale.Enabled && cfg.Listen == "" {
		logrus.Fatal("At least one server must be enabled (Tailscale or local HTTP)")
	}

	// Create application manager
	appManager := apps.NewManager(cfg.Applications)

	// Create servers
	apiServer := api.NewServer(appManager)
	uiServer := ui.NewServer(appManager)

	// Create main router and mount sub-routers
	mainRouter := mux.NewRouter()
	mainRouter.PathPrefix("/api/").Handler(apiServer.Routes())
	mainRouter.PathPrefix("/docs/").Handler(apiServer.Routes())
	mainRouter.PathPrefix("/").Handler(uiServer.Routes())

	// Start Tailscale server if enabled
	var tailscaleServer *http.Server
	if cfg.Tailscale.Enabled {
		// Setup Tailscale server
		tsServer := &tsnet.Server{
			Hostname: cfg.Tailscale.Name,
			Dir:      cfg.Tailscale.StateDir,
		}

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
