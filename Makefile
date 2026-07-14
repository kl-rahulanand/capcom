.PHONY: test vet run migrate-up tidy

test:
	go test ./...

vet:
	go vet ./...

run:
	go run ./cmd/capcom-server

migrate-up:
	go run ./cmd/capcom migrate up

tidy:
	go mod tidy
