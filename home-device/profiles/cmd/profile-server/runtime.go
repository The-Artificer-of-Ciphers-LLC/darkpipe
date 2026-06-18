// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/darkpipe/darkpipe/monitoring/cert"
	"github.com/darkpipe/darkpipe/monitoring/delivery"
	"github.com/darkpipe/darkpipe/monitoring/health"
	"github.com/darkpipe/darkpipe/monitoring/queue"
	"github.com/darkpipe/darkpipe/monitoring/status"
	"github.com/darkpipe/darkpipe/profiles/pkg/apppassword"
	"github.com/darkpipe/darkpipe/profiles/pkg/mobileconfig"
	"github.com/darkpipe/darkpipe/profiles/pkg/qrcode"
)

type profileRuntimeConfig struct {
	Server                ServerConfig
	Port                  string
	MailServerType        string
	AppPasswordStorePath  string
	MonitorCertPaths      []string
	MonitorHealthcheckURL string
	MonitorLogPath        string
}

type profileRuntime struct {
	config          profileRuntimeConfig
	server          *http.Server
	aggregator      *status.StatusAggregator
	deliveryTracker *delivery.DeliveryTracker
}

func runProfileRuntime() error {
	runtime, err := newProfileRuntime(loadProfileRuntimeConfig())
	if err != nil {
		return err
	}
	return runtime.Run()
}

func loadProfileRuntimeConfig() profileRuntimeConfig {
	return profileRuntimeConfig{
		Server: ServerConfig{
			Domain:      getEnv("MAIL_DOMAIN", "example.com"),
			Hostname:    getEnv("MAIL_HOSTNAME", "mail.example.com"),
			CalDAVURL:   getEnv("CALDAV_URL", ""),
			CardDAVURL:  getEnv("CARDDAV_URL", ""),
			CalDAVPort:  getEnvInt("CALDAV_PORT", 443),
			CardDAVPort: getEnvInt("CARDDAV_PORT", 443),
			AdminUser:   getEnv("ADMIN_USER", "admin"),
			AdminPass:   getEnv("ADMIN_PASSWORD", ""),
		},
		Port:                  getEnv("PROFILE_SERVER_PORT", "8090"),
		MailServerType:        getEnv("MAIL_SERVER_TYPE", "stalwart"),
		AppPasswordStorePath:  getEnv("APP_PASSWORD_STORE_PATH", "/data/app-passwords.json"),
		MonitorCertPaths:      splitEnvList(os.Getenv("MONITOR_CERT_PATHS")),
		MonitorHealthcheckURL: os.Getenv("MONITOR_HEALTHCHECK_URL"),
		MonitorLogPath:        getEnv("MONITOR_LOG_PATH", "/var/log/mail.log"),
	}
}

func newProfileRuntime(cfg profileRuntimeConfig) (*profileRuntime, error) {
	if cfg.Server.AdminPass == "" {
		log.Println("WARNING: ADMIN_PASSWORD not set, QR generation endpoints will be insecure")
	}

	appPassStore, storeLabel, err := newAppPasswordStore(cfg.MailServerType, cfg.AppPasswordStorePath)
	if err != nil {
		return nil, err
	}
	log.Print(storeLabel)

	tokenStore := qrcode.NewMemoryTokenStore()
	handler := &ProfileHandler{
		ProfileGen:   &mobileconfig.ProfileGenerator{},
		TokenStore:   tokenStore,
		AppPassStore: appPassStore,
		Config:       cfg.Server,
	}

	healthChecker := health.NewChecker()
	healthChecker.RegisterCheck("postfix", health.CheckPostfix)
	healthChecker.RegisterCheck("imap", health.CheckIMAP)

	deliveryTracker := delivery.NewDeliveryTracker(0)
	aggregator := status.NewStatusAggregator(
		healthChecker,
		queue.GetQueueStats,
		deliveryTracker,
		cert.NewCertWatcher(cfg.MonitorCertPaths),
	)

	webUI := NewWebUIHandler(appPassStore, tokenStore, cfg.Server)
	mux := http.NewServeMux()
	registerProfileRoutes(mux, handler, webUI, healthChecker, aggregator, loadStatusTemplate())

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      LogRequest(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &profileRuntime{
		config:          cfg,
		server:          server,
		aggregator:      aggregator,
		deliveryTracker: deliveryTracker,
	}, nil
}

func newAppPasswordStore(mailServerType, storePath string) (apppassword.Store, string, error) {
	switch mailServerType {
	case "stalwart":
		return apppassword.NewStalwartStore(), "Using Stalwart app password store", nil
	case "dovecot":
		return apppassword.NewDovecotStore(storePath), fmt.Sprintf("Using Dovecot app password store (path: %s)", storePath), nil
	case "maddy":
		return apppassword.NewMaddyStore(storePath), fmt.Sprintf("Using Maddy app password store (path: %s)", storePath), nil
	case "postfix-dovecot":
		return apppassword.NewDovecotStore(storePath), fmt.Sprintf("Using Dovecot app password store (path: %s)", storePath), nil
	default:
		return nil, "", fmt.Errorf("unknown MAIL_SERVER_TYPE: %s (supported: stalwart, dovecot, maddy, postfix-dovecot)", mailServerType)
	}
}

func registerProfileRoutes(
	mux *http.ServeMux,
	handler *ProfileHandler,
	webUI *WebUIHandler,
	healthChecker *health.Checker,
	aggregator *status.StatusAggregator,
	statusTmpl *template.Template,
) {
	mux.HandleFunc("/profile/download", handler.HandleProfileDownload)
	mux.HandleFunc("/mail/config-v1.1.xml", handler.HandleAutoconfig)
	mux.HandleFunc("/autodiscover/autodiscover.xml", handler.HandleAutodiscover)
	mux.HandleFunc("/health", handler.HandleHealth)
	mux.HandleFunc("/health/live", health.LivenessHandler(healthChecker))
	mux.HandleFunc("/health/ready", health.ReadinessHandler(healthChecker))
	mux.HandleFunc("/qr/generate", handler.HandleQRGenerate)
	mux.HandleFunc("/qr/image", handler.HandleQRImage)

	if statusTmpl != nil {
		mux.HandleFunc("/status", status.HandleDashboard(aggregator, statusTmpl))
	}
	mux.HandleFunc("/status/api", status.HandleStatusAPI(aggregator))

	mux.HandleFunc("/devices", webUI.HandleDeviceList)
	mux.HandleFunc("/devices/add", webUI.HandleAddDevice)
	mux.HandleFunc("/devices/revoke", webUI.HandleRevokeDevice)
	mux.HandleFunc("/static/", webUI.ServeStatic)
}

func loadStatusTemplate() *template.Template {
	statusTmpl, err := template.New("status.html").Funcs(template.FuncMap{
		"mul": func(value int, multiplier float64) float64 {
			return float64(value) * multiplier
		},
	}).ParseFiles("templates/status.html")
	if err != nil {
		log.Printf("WARNING: Could not load status template: %v (dashboard disabled)", err)
		return nil
	}
	return statusTmpl
}

func (r *profileRuntime) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if r.config.MonitorHealthcheckURL != "" {
		pinger := status.NewHealthchecksPinger(r.config.MonitorHealthcheckURL, 0)
		go pinger.Run(ctx, func() (*status.SystemStatus, error) {
			return r.aggregator.GetStatus(ctx)
		})
		log.Printf("Push monitoring enabled (URL: %s)", r.config.MonitorHealthcheckURL)
	}

	go startLogParser(ctx, r.config.MonitorLogPath, r.deliveryTracker)

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("Profile server listening on port %s", r.config.Port)
		if err := r.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	select {
	case <-quit:
		log.Println("Shutting down server...")
	case err := <-serverErr:
		if err != nil {
			return fmt.Errorf("server error: %w", err)
		}
		return nil
	}

	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := r.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	log.Println("Server exited")
	return nil
}

func splitEnvList(value string) []string {
	if value == "" {
		return nil
	}
	return strings.Split(value, ",")
}

// getEnv gets environment variable with fallback to default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets environment variable as integer with fallback to default value.
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
