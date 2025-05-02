# Metric Agent

**Metric Agent** is a lightweight Go utility that periodically collects runtime metrics from your
application and ships them to a remote HTTP endpoint. Data is serialized as JSON,
compressed via GZIP, and sent in configurable batches.


## Features

- **Collector**: samples Go runtime metrics (e.g. Alloc, HeapSys, NumGC, custom gauges) at a configurable interval.
- **Sender**: accumulates metrics into a slice and sends the batch every N seconds.
- **JSON + GZIP**: payloads are encoded as JSON and compressed with GZIP to reduce bandwidth.
- **Graceful Shutdown**: honors `context.Context` cancellation and signals (SIGINT/SIGTERM).
- **Buffered Channel**: decouples collector and sender, avoiding backpressure.
- **Tests**: unit tests using `httptest.Server`, gzip round-trip, context cancellation.

## Installation

```bash
git clone https://github.com/sanchey92/metric-agent.git
cd metric-agent
go build -o metric-agent cmd/agent/main.go
```

**Or run directly**: 

```bash
go run cmd/agent/main.go --address=http://localhost:8080/metrics --poll=2s --report=10s
```

## Testing

**Run all unit tests with**:
```bash
go test ./... -v
```
