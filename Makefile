NAME := coverage
VERSION ?=

.PHONY: cover clean covero coversum

cover:
	# TEST_LIVE=1 go test -v -cover -coverprofile $(NAME).out -coverpkg=./internal/...,./pkg/... ./...
	# go test -v -cover -coverprofile $(NAME).out -covermode=atomic -coverpkg=./internal/...,./pkg/... ./...
	go test -cover -coverprofile $(NAME).out -covermode=atomic $(shell go list ./... | grep -v '/cmd/')
	# go tool cover -func $(NAME).out

covero: cover
	go tool cover -html $(NAME).out -o $(NAME).html
	open $(NAME).html

coversum: cover
	@echo "=== Package Coverage Summary ==="
	@go tool cover -func $(NAME).out | awk '/^total:/ {print $0}'

clean:
	@bash -c "rm -f $(NAME).{html,out}"

.PHONY: build
build:
	go build -ldflags "-X github.com/rschmied/gocmlclient/internal/version.Version=$(VERSION)" ./...

.PHONY: update
update:
	go get -u ./...
	go mod verify
	go mod download
	go mod tidy

.PHONY: integration
integration:
	go test -count=1 -v -tags=integration ./integration # -run TestIntegration_LabImportFromFiles

.PHONY: loc
loc:
	@echo "=== Lines of Code ==="
	@find . cmd pkg internal -name "*.go" -not -path "*/vendor/*" -exec wc -l {} + | awk '{total += $$1} END {print "Total:", total}'
