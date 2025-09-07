.PHONY: install
install: 
	@uvx pre-commit install
	@uvx pre-commit install --hook-type commit-msg

.PHONY: test
test:
	@echo "run tests"
	# @go test -v -json ./... | tparse -all
	@go test $(go list ./... | grep -v /cmd/) -v -json | tparse -all

.PHONY: lint
lint:
	@echo "run lint"
	@golangci-lint run
