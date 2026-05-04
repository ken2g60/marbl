# Golang Takehome Task

This repository contains two Go microservices (Producer and Consumer) that communicate via gRPC. 
They track tasks in a PostgreSQL database and expose Prometheus metrics and pprof profiling endpoints.

## Architecture & Tooling
- **Communication**: gRPC (efficient, typed, bi-directional capability).
- **Database**: PostgreSQL with `sqlc` for type-safe query generation and `golang-migrate` for schema migrations (embedded via `go:embed`).
- **Monitoring**: Prometheus & Grafana for metrics visualization.
- **Profiling**: `net/http/pprof` for generating flame graphs and analyzing performance.
- **Standard Library**: Used `log/slog` for structured logging, `flag` & `os` for config, and standard `http` for metrics/profiling. `golang.org/x/time/rate` used for rate limiting.

## GOGC and GOMEMLIMIT
We set `GOMEMLIMIT` in `docker-compose.yml` to define a soft memory limit for the Go runtime, which helps prevent OOM errors in containerized environments. `GOGC` can also be tuned, but `GOMEMLIMIT` usually provides better control over memory usage tradeoffs by automatically adjusting GC frequency when approaching the limit.

## Build Flags & Optimizations
We extensively use Go compiler and linker flags to optimize the production binaries. As requested in the task, you can see the full list of available flags by running `go tool link` or `go tool compile`. 

The optimizations we use include:

**Linker Flags (`go tool link`)**:
- `-s`: Disables symbol table generation.
- `-w`: Disables DWARF generation.
Together, `-ldflags="-s -w"` significantly reduce the final containerized binary size.
- `-X 'main.Version=...'`: Dynamically injects the current version string at build time. This is accessible in the app by running the binary with the `-version` flag.

**Compiler Flags (`go tool compile`)**:
- `-pgo=auto`: Enables Profile-Guided Optimization. If you capture a CPU profile using the `/debug/pprof/profile` endpoint and save it as `default.pgo` in the package directory, the Go compiler will automatically use it to inline hot functions, reducing CPU usage overhead.
```sh
make run
```
Alternatively, you can use docker-compose directly:
```sh
docker-compose up --build
```
This will start:
- PostgreSQL (port 5432)
- Producer (gRPC - inside docker, Prometheus :9092, pprof :6062)
- Consumer (gRPC :50051, Prometheus :9091, pprof :6061)
- Prometheus (:9090)
- Grafana (:3000)

## Profiling & Flame Graphs
Profiling endpoints are exposed on `6061` (Consumer) and `6062` (Producer).

To capture a CPU profile and generate a flame graph, you need Go installed locally:
```sh
# 1. Capture a 30s CPU profile from the consumer
curl -o cpu.prof http://localhost:6061/debug/pprof/profile?seconds=30

# 2. View in web browser (including Flame Graph view)
go tool pprof -http=:8080 cpu.prof
```

## Testing
Run unit tests with:
```sh
go test ./... -v -cover
```
