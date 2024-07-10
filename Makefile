TOOLS_SHELL="./deploy/scripts/tools.sh"
# golangci-lint
LINTER := bin/golangci-lint

$(LINTER):
	curl -SL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s latest

.PHONY: tidy
# tidy all mod
tidy:
	@${TOOLS_SHELL} tidy $(dir)
	@echo "tidy finished"

.PHONY: fix
# lint fix
fix: $(LINTER)
	@${TOOLS_SHELL} fix $(dir)
	@echo "lint fix finished"

.PHONY: lint
# lint check
lint: $(LINTER)
	@${TOOLS_SHELL} lint $(dir)
	@echo "lint check finished"

.PHONY: test
# test module
test:
	@${TOOLS_SHELL} test $(dir)
	@echo "go test finished"

.PHONY: test-coverage
# test module with coverage
test-coverage:
	@${TOOLS_SHELL} test_coverage $(dir)
	@echo "go test with coverage finished"

# show help
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-_0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help