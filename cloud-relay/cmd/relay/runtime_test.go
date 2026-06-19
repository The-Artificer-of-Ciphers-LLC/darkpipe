// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/darkpipe/darkpipe/cloud-relay/relay/config"
	"github.com/darkpipe/darkpipe/cloud-relay/relay/forward"
	"github.com/darkpipe/darkpipe/cloud-relay/relay/queue"
)

func TestNewRelayRuntime_WireGuardWithoutQueue(t *testing.T) {
	cfg := testRelayConfig()
	cfg.TransportType = "wireguard"
	cfg.QueueEnabled = false

	wireGuardForwarder := &fakeForwarder{name: "wireguard"}
	server := &fakeRelayServer{}
	var serverForwarder forward.Forwarder

	runtime, err := newRelayRuntimeWithFactories(cfg, relayRuntimeFactories{
		newWireGuard: func(homeAddr string) forward.Forwarder {
			if homeAddr != cfg.HomeDeviceAddr {
				t.Fatalf("homeAddr = %q, want %q", homeAddr, cfg.HomeDeviceAddr)
			}
			return wireGuardForwarder
		},
		newSMTPServer: func(forwarder forward.Forwarder, cfg *config.Config) relayServer {
			serverForwarder = forwarder
			return server
		},
	})
	if err != nil {
		t.Fatalf("newRelayRuntimeWithFactories() error: %v", err)
	}

	if serverForwarder != wireGuardForwarder {
		t.Fatal("server did not receive the WireGuard forwarder directly")
	}

	if err := runtime.ListenAndServe(); err != nil {
		t.Fatalf("ListenAndServe() error: %v", err)
	}
	if !server.listened {
		t.Fatal("server was not started")
	}

	if err := runtime.Close(); err != nil {
		t.Fatalf("Close() error: %v", err)
	}
	if !wireGuardForwarder.closed {
		t.Fatal("transport forwarder was not closed")
	}
}

func TestNewRelayRuntime_MTLSQueueAndOverflowAssembly(t *testing.T) {
	cfg := testRelayConfig()
	cfg.TransportType = "mtls"
	cfg.QueueEnabled = true
	cfg.QueueKeyPath = "/queue/key"
	cfg.QueueMaxRAMBytes = 64 * 1024 * 1024
	cfg.QueueMaxMessages = 42
	cfg.QueueTTLHours = 12
	cfg.QueueSnapshotPath = "/queue/snapshot.json"
	cfg.OverflowEnabled = true
	cfg.OverflowEndpoint = "s3.example.test"
	cfg.OverflowAccessKey = "access"
	cfg.OverflowSecretKey = "secret"
	cfg.OverflowBucket = "bucket"
	cfg.OverflowUseSSL = true

	mtlsForwarder := &fakeForwarder{name: "mtls"}
	queuedForwarder := &fakeForwarder{name: "queued"}
	server := &fakeRelayServer{}
	msgQueue := &queue.MessageQueue{}
	var queueCfg queue.QueueConfig
	var serverForwarder forward.Forwarder
	var processorCtx context.Context
	var processorInterval time.Duration

	runtime, err := newRelayRuntimeWithFactories(cfg, relayRuntimeFactories{
		newMTLSForwarder: func(caCert, clientCert, clientKey, homeAddr string) (forward.Forwarder, error) {
			if caCert != cfg.CACertPath || clientCert != cfg.ClientCertPath || clientKey != cfg.ClientKeyPath || homeAddr != cfg.HomeDeviceAddr {
				t.Fatalf("mTLS factory received unexpected config")
			}
			return mtlsForwarder, nil
		},
		newMessageQueue: func(cfg queue.QueueConfig) (*queue.MessageQueue, error) {
			queueCfg = cfg
			return msgQueue, nil
		},
		newOverflowStorage: func(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*queue.OverflowStorage, error) {
			if endpoint != cfg.OverflowEndpoint || accessKey != cfg.OverflowAccessKey || secretKey != cfg.OverflowSecretKey || bucket != cfg.OverflowBucket || useSSL != cfg.OverflowUseSSL {
				t.Fatalf("overflow factory received unexpected config")
			}
			return nil, nil
		},
		newQueuedForwarder: func(transport forward.Forwarder, q *queue.MessageQueue, enabled bool) forward.Forwarder {
			if transport != mtlsForwarder {
				t.Fatal("queued forwarder did not receive mTLS transport")
			}
			if q != msgQueue {
				t.Fatal("queued forwarder did not receive assembled queue")
			}
			if !enabled {
				t.Fatal("queued forwarder was not enabled")
			}
			return queuedForwarder
		},
		startProcessor: func(ctx context.Context, q *queue.MessageQueue, transport forward.Forwarder, interval time.Duration) {
			if q != msgQueue {
				t.Fatal("processor did not receive assembled queue")
			}
			if transport != mtlsForwarder {
				t.Fatal("processor did not receive mTLS transport")
			}
			processorCtx = ctx
			processorInterval = interval
		},
		newSMTPServer: func(forwarder forward.Forwarder, cfg *config.Config) relayServer {
			serverForwarder = forwarder
			return server
		},
	})
	if err != nil {
		t.Fatalf("newRelayRuntimeWithFactories() error: %v", err)
	}

	if queueCfg.KeyPath != cfg.QueueKeyPath || queueCfg.MaxRAMBytes != cfg.QueueMaxRAMBytes || queueCfg.MaxMessages != cfg.QueueMaxMessages || queueCfg.TTLHours != cfg.QueueTTLHours || queueCfg.SnapshotPath != cfg.QueueSnapshotPath {
		t.Fatalf("queue config = %+v, want values from relay config", queueCfg)
	}
	if serverForwarder != queuedForwarder {
		t.Fatal("server did not receive queued forwarder")
	}
	if processorInterval != 30*time.Second {
		t.Fatalf("processor interval = %s, want 30s", processorInterval)
	}

	if err := runtime.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown() error: %v", err)
	}
	if !server.shutdown {
		t.Fatal("server was not shut down")
	}
	if processorCtx == nil || processorCtx.Err() == nil {
		t.Fatal("processor context was not canceled")
	}
}

func TestNewRelayRuntime_StrictModeWarningsDoNotFailAssembly(t *testing.T) {
	cfg := testRelayConfig()
	cfg.StrictModeEnabled = true
	cfg.QueueEnabled = false

	runtime, err := newRelayRuntimeWithFactories(cfg, relayRuntimeFactories{
		newStrictMode: func(enabled bool) strictMode {
			if !enabled {
				t.Fatal("strict mode was not enabled")
			}
			return &fakeStrictMode{
				generateErr: errors.New("generate failed"),
				applyErr:    errors.New("apply failed"),
			}
		},
		newWireGuard: func(homeAddr string) forward.Forwarder {
			return &fakeForwarder{name: "wireguard"}
		},
		newSMTPServer: func(forwarder forward.Forwarder, cfg *config.Config) relayServer {
			return &fakeRelayServer{}
		},
	})
	if err != nil {
		t.Fatalf("newRelayRuntimeWithFactories() error: %v", err)
	}
	if runtime == nil {
		t.Fatal("runtime is nil")
	}
}

func TestNewRelayRuntime_ReturnsTransportErrors(t *testing.T) {
	cfg := testRelayConfig()
	cfg.TransportType = "mtls"
	cfg.QueueEnabled = false

	wantErr := errors.New("missing cert")
	_, err := newRelayRuntimeWithFactories(cfg, relayRuntimeFactories{
		newMTLSForwarder: func(caCert, clientCert, clientKey, homeAddr string) (forward.Forwarder, error) {
			return nil, wantErr
		},
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("error = %v, want wrapped %v", err, wantErr)
	}
}

func testRelayConfig() *config.Config {
	return &config.Config{
		ListenAddr:       "127.0.0.1:10025",
		TransportType:    "wireguard",
		HomeDeviceAddr:   "10.8.0.2:25",
		CACertPath:       "/certs/ca.pem",
		ClientCertPath:   "/certs/client.pem",
		ClientKeyPath:    "/certs/client-key.pem",
		MaxMessageBytes:  1024,
		ReadTimeout:      time.Second,
		WriteTimeout:     time.Second,
		QueueMaxRAMBytes: 200 * 1024 * 1024,
		QueueMaxMessages: 10000,
		QueueTTLHours:    168,
	}
}

type fakeForwarder struct {
	name   string
	closed bool
}

func (f *fakeForwarder) Forward(ctx context.Context, from string, to []string, data io.Reader) error {
	return nil
}

func (f *fakeForwarder) Close() error {
	f.closed = true
	return nil
}

type fakeRelayServer struct {
	listened bool
	shutdown bool
}

func (s *fakeRelayServer) ListenAndServe() error {
	s.listened = true
	return nil
}

func (s *fakeRelayServer) Shutdown(ctx context.Context) error {
	s.shutdown = true
	return nil
}

type fakeStrictMode struct {
	generateErr error
	applyErr    error
}

func (s *fakeStrictMode) GeneratePolicyMap() error {
	return s.generateErr
}

func (s *fakeStrictMode) ApplyToPostfix() error {
	return s.applyErr
}
