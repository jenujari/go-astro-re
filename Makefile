APP_NAME=go-astro-re

.PHONY: test fmt tidy run-srv run-srv-direct

test:
	go test ./...

fmt:
	go fmt ./...

tidy:
	go mod tidy

run-srv:
	go tool air -c .air.toml

run-srv-direct:
	go run ./entry/srv/main.go
