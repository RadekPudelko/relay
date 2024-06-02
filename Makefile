all:
	go run cmd/api/main.go

client:
	go run examples/client/main.go

int:
	go test test/integration_test.go test/test_utils.go -v | tee int.txt

unit: relay cancellation client_test

relay:
	go test test/relay_test.go test/test_utils.go -v

cancellation:
	go test test/cancellation_test.go test/test_utils.go -v

client_test:
	go test test/client_test.go test/test_utils.go -v

fmt:
	gofmt -s -w .

.PHONY: client

