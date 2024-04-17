// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package tracetest is a testing helper package for the SDK. User can
// configure no-op or in-memory exporters to verify different SDK behaviors or
// custom instrumentation.
package opentelemetryexporter

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

var _ trace.SpanExporter = (*PersistentInMemoryExporter)(nil)

// NewInMemoryExporter returns a new InMemoryExporter.
func NewPersistentInMemoryExporter() *PersistentInMemoryExporter {
	return new(PersistentInMemoryExporter)
}

// InMemoryExporter is an exporter that stores all received spans in-memory.
type PersistentInMemoryExporter struct {
	mu sync.Mutex
	ss tracetest.SpanStubs
}

// ExportSpans handles export of spans by storing them in memory.
func (imsb *PersistentInMemoryExporter) ExportSpans(_ context.Context, spans []trace.ReadOnlySpan) error {
	imsb.mu.Lock()
	defer imsb.mu.Unlock()
	imsb.ss = append(imsb.ss, tracetest.SpanStubsFromReadOnlySpans(spans)...)
	return nil
}

// Shutdown stops the exporter by clearing spans held in memory.
func (imsb *PersistentInMemoryExporter) Shutdown(context.Context) error {
	imsb.Reset()
	return nil
}

// Reset the current in-memory storage.
func (imsb *PersistentInMemoryExporter) Reset() {
	imsb.mu.Lock()
	defer imsb.mu.Unlock()
	//imsb.ss = nil
}

// GetSpans returns the current in-memory stored spans.
func (imsb *PersistentInMemoryExporter) GetSpans() tracetest.SpanStubs {
	imsb.mu.Lock()
	defer imsb.mu.Unlock()
	ret := make(tracetest.SpanStubs, len(imsb.ss))
	copy(ret, imsb.ss)
	return ret
}

func (imsb *PersistentInMemoryExporter) Clear(context.Context) error {
	imsb.ss = nil
	return nil
}

type SliceLogConsumer struct {
	logs []string
}

// Accept prints the log to stdout
func (lc *SliceLogConsumer) Accept(l testcontainers.Log) {
	lc.logs = append(lc.logs, string(l.Content))
}

func (lc *SliceLogConsumer) Exists(target string) bool {
	for _, s := range lc.logs {

		if strings.Contains(s, target) {
			return true
		}

	}

	return false
}

type opentelCollectorContainer struct {
	testcontainers.Container
	URI string
}

func setupotelCollectorContainer(ctx context.Context, consumer testcontainers.LogConsumer) (*opentelCollectorContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "otel/opentelemetry-collector:0.98.0",
		ExposedPorts: []string{"4317/tcp", "55679/tcp"},
		WaitingFor:   wait.ForLog("Everything is ready. Begin running and processing data"),
	}

	logConsumerConfig := testcontainers.LogConsumerConfig{
		Opts:      []testcontainers.LogProductionOption{testcontainers.WithLogProductionTimeout(10 * time.Second)},
		Consumers: []testcontainers.LogConsumer{consumer},
	}

	request := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}
	request.LogConsumerCfg = &logConsumerConfig
	container, err := testcontainers.GenericContainer(ctx, request)
	if err != nil {
		return nil, err
	}

	ip, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, "4317")
	if err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("%s:%s", ip, mappedPort.Port())

	return &opentelCollectorContainer{Container: container, URI: uri}, nil
}
