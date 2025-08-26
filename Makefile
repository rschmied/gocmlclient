NAME := coverage

.PHONY: cover clean

cover:
	# TEST_LIVE=1 go test -v -cover -coverprofile $(NAME).out -coverpkg=./internal/...,./pkg/... ./...
	go test -v -cover -coverprofile $(NAME).out -coverpkg=./internal/...,./pkg/... ./...
	go tool cover -html $(NAME).out -o $(NAME).html
	open $(NAME).html

clean:
	@bash -c "rm -f $(NAME).{html,out}"
