.PHONY: build run test clean docker-build docker-up

build:
	cd services/core-sync && go build -o ../../bin/sync-service ./cmd/server

run:
	cd services/core-sync && go run ./cmd/server/main.go

test:
	cd services/core-sync && go test ./...

clean:
	rm -rf bin/

docker-build:
	docker-compose -f deployments/docker-compose.yml build

docker-up:
	docker-compose -f deployments/docker-compose.yml up -d
