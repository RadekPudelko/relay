all:
	go run cmd/api/main.go

client:
	go run cmd/client/main.go

int:
	go test test/integration_test.go test/test_utils.go -v | tee int.txt

unit:
	go test test/unit_test.go -v

cancel:
	go test test/cancellation_test.go test/test_utils.go -v

fmt:
	gofmt -s -w .

.PHONY: client

