// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/darkpipe/darkpipe/cloud-relay/relay/config"
	"github.com/darkpipe/darkpipe/cloud-relay/relay/forward"
	"github.com/darkpipe/darkpipe/cloud-relay/relay/notify"
	"github.com/darkpipe/darkpipe/cloud-relay/relay/queue"
	"github.com/darkpipe/darkpipe/cloud-relay/relay/smtp"
	"github.com/darkpipe/darkpipe/cloud-relay/relay/tls"
)

type relayServer interface {
	ListenAndServe() error
	Shutdown(context.Context) error
}

type strictMode interface {
	GeneratePolicyMap() error
	ApplyToPostfix() error
}

type relayRuntime struct {
	server             relayServer
	notifier           notify.Notifier
	transportForwarder forward.Forwarder
	processorCancel    context.CancelFunc
}

type relayRuntimeFactories struct {
	newWebhookNotifier func(string) notify.Notifier
	newMultiNotifier   func(...notify.Notifier) notify.Notifier
	newStrictMode      func(bool) strictMode
	newMTLSForwarder   func(string, string, string, string) (forward.Forwarder, error)
	newWireGuard       func(string) forward.Forwarder
	newMessageQueue    func(queue.QueueConfig) (*queue.MessageQueue, error)
	newOverflowStorage func(string, string, string, string, bool) (*queue.OverflowStorage, error)
	newQueuedForwarder func(forward.Forwarder, *queue.MessageQueue, bool) forward.Forwarder
	newSMTPServer      func(forward.Forwarder, *config.Config) relayServer
	startProcessor     func(context.Context, *queue.MessageQueue, forward.Forwarder, time.Duration)
}

func newRelayRuntime(cfg *config.Config) (*relayRuntime, error) {
	return newRelayRuntimeWithFactories(cfg, relayRuntimeFactories{})
}

func newRelayRuntimeWithFactories(cfg *config.Config, factories relayRuntimeFactories) (*relayRuntime, error) {
	factories = factories.withDefaults()

	notifier := buildNotifier(cfg, factories)
	applyStrictMode(cfg, factories)

	transportForwarder, err := buildTransportForwarder(cfg, factories)
	if err != nil {
		notifier.Close()
		return nil, err
	}

	activeForwarder, processorCancel, err := buildActiveForwarder(cfg, factories, transportForwarder)
	if err != nil {
		notifier.Close()
		transportForwarder.Close()
		return nil, err
	}

	server := factories.newSMTPServer(activeForwarder, cfg)
	log.Printf("Relay daemon listening on %s (forwarding to %s via %s)", cfg.ListenAddr, cfg.HomeDeviceAddr, cfg.TransportType)

	return &relayRuntime{
		server:             server,
		notifier:           notifier,
		transportForwarder: transportForwarder,
		processorCancel:    processorCancel,
	}, nil
}

func (r *relayRuntime) ListenAndServe() error {
	return r.server.ListenAndServe()
}

func (r *relayRuntime) Shutdown(ctx context.Context) error {
	if r.processorCancel != nil {
		r.processorCancel()
		r.processorCancel = nil
	}
	return r.server.Shutdown(ctx)
}

func (r *relayRuntime) Close() error {
	var firstErr error
	if r.notifier != nil {
		if err := r.notifier.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if r.transportForwarder != nil {
		if err := r.transportForwarder.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func buildNotifier(cfg *config.Config, factories relayRuntimeFactories) notify.Notifier {
	if cfg.WebhookURL == "" {
		log.Println("TLS monitor ready (will be activated via entrypoint.sh)")
		return &noopNotifier{}
	}

	log.Printf("Enabling webhook notifications to %s", cfg.WebhookURL)
	webhookNotifier := factories.newWebhookNotifier(cfg.WebhookURL)
	notifier := factories.newMultiNotifier(webhookNotifier)
	log.Println("TLS monitor ready (will be activated via entrypoint.sh)")
	return notifier
}

func applyStrictMode(cfg *config.Config, factories relayRuntimeFactories) {
	if !cfg.StrictModeEnabled {
		return
	}

	log.Println("Applying strict TLS mode to Postfix...")
	strictMode := factories.newStrictMode(true)
	if err := strictMode.GeneratePolicyMap(); err != nil {
		log.Printf("WARNING: Failed to generate TLS policy map: %v", err)
	}
	if err := strictMode.ApplyToPostfix(); err != nil {
		log.Printf("WARNING: Failed to apply strict mode to Postfix: %v", err)
	}
}

func buildTransportForwarder(cfg *config.Config, factories relayRuntimeFactories) (forward.Forwarder, error) {
	if cfg.TransportType == "mtls" {
		log.Println("Initializing mTLS forwarder...")
		transportForwarder, err := factories.newMTLSForwarder(
			cfg.CACertPath,
			cfg.ClientCertPath,
			cfg.ClientKeyPath,
			cfg.HomeDeviceAddr,
		)
		if err != nil {
			return nil, fmt.Errorf("create mTLS forwarder: %w", err)
		}
		return transportForwarder, nil
	}

	log.Println("Initializing WireGuard forwarder...")
	return factories.newWireGuard(cfg.HomeDeviceAddr), nil
}

func buildActiveForwarder(
	cfg *config.Config,
	factories relayRuntimeFactories,
	transportForwarder forward.Forwarder,
) (forward.Forwarder, context.CancelFunc, error) {
	if !cfg.QueueEnabled {
		log.Println("Queue disabled: mail will bounce when home device is offline")
		return transportForwarder, nil, nil
	}

	log.Println("Initializing encrypted message queue...")
	msgQueue, err := factories.newMessageQueue(queue.QueueConfig{
		KeyPath:      cfg.QueueKeyPath,
		MaxRAMBytes:  cfg.QueueMaxRAMBytes,
		MaxMessages:  cfg.QueueMaxMessages,
		TTLHours:     cfg.QueueTTLHours,
		SnapshotPath: cfg.QueueSnapshotPath,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("initialize message queue: %w", err)
	}

	if cfg.OverflowEnabled {
		log.Printf("Initializing S3 overflow storage: endpoint=%s bucket=%s", cfg.OverflowEndpoint, cfg.OverflowBucket)
		overflow, err := factories.newOverflowStorage(
			cfg.OverflowEndpoint,
			cfg.OverflowAccessKey,
			cfg.OverflowSecretKey,
			cfg.OverflowBucket,
			cfg.OverflowUseSSL,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("initialize overflow storage: %w", err)
		}
		msgQueue.SetOverflow(overflow)
		log.Println("S3 overflow storage enabled")
	}

	processorCtx, cancel := context.WithCancel(context.Background())
	factories.startProcessor(processorCtx, msgQueue, transportForwarder, 30*time.Second)

	log.Printf("Queue enabled: max_ram=%dMB max_messages=%d ttl=%dh",
		cfg.QueueMaxRAMBytes/(1024*1024), cfg.QueueMaxMessages, cfg.QueueTTLHours)
	return factories.newQueuedForwarder(transportForwarder, msgQueue, true), cancel, nil
}

func (f relayRuntimeFactories) withDefaults() relayRuntimeFactories {
	if f.newWebhookNotifier == nil {
		f.newWebhookNotifier = func(url string) notify.Notifier {
			return notify.NewWebhookNotifier(url)
		}
	}
	if f.newMultiNotifier == nil {
		f.newMultiNotifier = func(notifiers ...notify.Notifier) notify.Notifier {
			return notify.NewMultiNotifier(notifiers...)
		}
	}
	if f.newStrictMode == nil {
		f.newStrictMode = func(enabled bool) strictMode {
			return tls.NewStrictMode(enabled)
		}
	}
	if f.newMTLSForwarder == nil {
		f.newMTLSForwarder = func(caCert, clientCert, clientKey, homeAddr string) (forward.Forwarder, error) {
			return forward.NewMTLSForwarder(caCert, clientCert, clientKey, homeAddr)
		}
	}
	if f.newWireGuard == nil {
		f.newWireGuard = func(homeAddr string) forward.Forwarder {
			return forward.NewWireGuardForwarder(homeAddr)
		}
	}
	if f.newMessageQueue == nil {
		f.newMessageQueue = queue.NewMessageQueue
	}
	if f.newOverflowStorage == nil {
		f.newOverflowStorage = queue.NewOverflowStorage
	}
	if f.newQueuedForwarder == nil {
		f.newQueuedForwarder = func(transport forward.Forwarder, msgQueue *queue.MessageQueue, enabled bool) forward.Forwarder {
			return forward.NewQueuedForwarder(transport, msgQueue, enabled)
		}
	}
	if f.newSMTPServer == nil {
		f.newSMTPServer = func(forwarder forward.Forwarder, cfg *config.Config) relayServer {
			return smtp.NewServer(forwarder, cfg)
		}
	}
	if f.startProcessor == nil {
		f.startProcessor = func(ctx context.Context, msgQueue *queue.MessageQueue, transport forward.Forwarder, interval time.Duration) {
			go msgQueue.StartProcessor(ctx, transport, interval)
		}
	}
	return f
}

// noopNotifier is a no-op notifier used when webhook notifications are disabled.
type noopNotifier struct{}

func (n *noopNotifier) Send(ctx context.Context, event notify.Event) error {
	return nil
}

func (n *noopNotifier) Close() error {
	return nil
}
