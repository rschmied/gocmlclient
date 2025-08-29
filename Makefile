NAME := coverage

.PHONY: cover clean update

update:
	go get -u ./...
	go mod download && go mod verify && go mod tidy

cover:
	go test -v -coverprofile $(NAME).out ./*.go
	go tool cover -html $(NAME).out -o $(NAME).html
	open $(NAME).html

clean:
	@bash -c "rm -f $(NAME).{html,out}"
