lint-run:
	@echo "Lint Analysis Starting"
	golangci-lint run
	@echo "Lint Analysis Finished"

lint-fix:
	@echo "Lint Fix Starting"
	golangci-lint run --fix
	@echo "Lint Fix Finished"

run:
	@echo "Application Starting"
	go run cmd/main.go

all: lint-fix lint-run run
