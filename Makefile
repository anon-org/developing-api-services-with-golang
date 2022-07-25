fmt:
	@go vet ./...
	@gofmt -w -s .

test:
	@go test -count=1 -v -json -coverprofile=coverage.out -covermode=count ./... > result.json.out

test/coverage: test
	@go tool cover -html=coverage.out