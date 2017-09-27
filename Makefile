dirs = .

all: kylixor

kylixor: ## Default action. Builds Kylixor.
	@env GO15VENDOREXPERIMENT="1" go build .

.PHONY: test
test: ## Runs unit tests for Kylixor.
	@env GO15VENDOREXPERIMENT="1" go test $(dirs)

.PHONY: clean
clean: ## Removes compiled Kylixor binaries.
	@rm -f kylixor*

.PHONY: install
install: ## Copies kylixor binary to /usr/local/bin for easy execution.
	@cp -f kylixor* /usr/local/bin/kylixor

.PHONY: help
help: ## Shows this helptext.
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
