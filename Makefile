lint:
	@echo "Begin to run lint check"
	@golangci-lint run
	@echo "Success to pass lint check"
.PHONY: lint
