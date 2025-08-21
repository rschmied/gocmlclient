NAME := coverage

.PHONY: cover clean

cover:
	go test -v -coverprofile $(NAME).out ./...
	go tool cover -html $(NAME).out -o $(NAME).html
	open $(NAME).html

clean:
	@bash -c "rm -f $(NAME).{html,out}"
