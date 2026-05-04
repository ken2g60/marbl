.PHONY: run down build test proto sqlc clean

# Start the complete environment (Producer, Consumer, DB, Prometheus, Grafana)
run:
	docker-compose up --build

# Stop the environment
down:
	docker-compose down

# Build the docker images using docker-compose
build:
	docker-compose build

logs:
	docker-compose logs -f

# Build local binaries with ldflags (size reduction, version injection) and pgo
build-local:
	go build -pgo=auto -ldflags="-s -w -X 'main.Version=$(shell git describe --tags --always HEAD || echo dev)'" -o bin/producer cmd/producer/main.go
	go build -pgo=auto -ldflags="-s -w -X 'main.Version=$(shell git describe --tags --always HEAD || echo dev)'" -o bin/consumer cmd/consumer/main.go

# Run unit tests
test:
	go test ./... -v -cover

# Generate protobuf code
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       proto/marbl.proto

# Generate database code with sqlc
sqlc:
	sqlc generate

# Clean up build artifacts and containers
clean:
	docker-compose down -v
	rm -rf bin/
