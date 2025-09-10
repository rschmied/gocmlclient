NAME := coverage

.PHONY: cover clean

cover:
	# TEST_LIVE=1 go test -v -cover -coverprofile $(NAME).out -coverpkg=./internal/...,./pkg/... ./...
	go test -v -cover -coverprofile $(NAME).out -coverpkg=./internal/...,./pkg/... ./...
	go tool cover -html $(NAME).out -o $(NAME).html

covero: cover
	open $(NAME).html

clean:
	@bash -c "rm -f $(NAME).{html,out}"

.PHONY: update
update:
	go get -u ./...
	go mod verify
	go mod download
	go mod tidy
