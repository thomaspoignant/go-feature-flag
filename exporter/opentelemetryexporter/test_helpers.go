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

// NewPersistentInMemoryExporter returns a new PersistentInMemoryExporter.
func NewPersistentInMemoryExporter() *PersistentInMemoryExporter {
	return new(PersistentInMemoryExporter)
}

// PersistentInMemoryExporter is an exporter that stores all received spans in-memory.
type PersistentInMemoryExporter struct {
	tracetest.InMemoryExporter
}

func (imsb *PersistentInMemoryExporter) Shutdown(context.Context) error {
	return nil
}

// SliceLogConsumer buffers log content into a slice
type SliceLogConsumer struct {
	logs []string
	lock sync.Mutex
}

func (lc *SliceLogConsumer) Size() int {
	lc.lock.Lock()
	defer lc.lock.Unlock()
	return len(lc.logs)
}

// Accept prints the log to stdout
func (lc *SliceLogConsumer) Accept(l testcontainers.Log) {
	lc.lock.Lock()
	defer lc.lock.Unlock()
	lc.logs = append(lc.logs, string(l.Content))
}

// Exists checks if the target exists anywhere in the log output
func (lc *SliceLogConsumer) Exists(target string) bool {
	lc.lock.Lock()
	defer lc.lock.Unlock()
	for _, s := range lc.logs {
		if strings.Contains(s, target) {
			return true
		}
	}
	return false
}

// opentelCollectorContainer struct for the test container and URI
type opentelCollectorContainer struct {
	testcontainers.Container
	URI string
}

// setupOtelCollectorContainer sets up an otel container with a log consumer
func setupOtelCollectorContainer(ctx context.Context,
	consumer testcontainers.LogConsumer) (*opentelCollectorContainer, error) {
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
