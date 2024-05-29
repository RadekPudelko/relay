all:
	go run cmd/api/main.go

client:
	go run cmd/client/main.go

int:
	go test test/integration_test.go test/test_utils.go -v

unit:
	go test test/unit_test.go -v

fmt:
	gofmt -s -w .

.PHONY: client

