.PHONY: test vet run migrate-up

test:
	go test ./...

vet:
	go vet ./...

run:
	go run ./cmd/capcom-server

migrate-up:
	go run ./cmd/capcom migrate up
