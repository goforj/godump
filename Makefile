.PHONY: help

HELP_FUN = %help; while (<>) { /^([A-Za-z0-9_-]+)\s*:.*\#\#(?:@([A-Za-z0-9_-]+))?\s(.*)$$/ or next; push @{$$help{$$2 || "other"}}, [$$1, $$3]; $$width = length($$1) if length($$1) > $$width } print "\e[1;97m$(or $(HELP_NAME),$(notdir $(CURDIR)))\e[0m\n\n"; for $$category (sort keys %help) { print "\e[1;97m$$category\e[0m\n"; for $$entry (@{$$help{$$category}}) { printf "  \e[1;32m%-*s\e[0m  \e[90m%s\e[0m\n", $$width, $$entry->[0], $$entry->[1] } }

help: ##@other Show this help.
	@perl -e '$(HELP_FUN)' $(MAKEFILE_LIST)

##@tests
test: ##@tests Run the test suite.
	go test ./...

##@analysis
vet: ##@analysis Run Go vet.
	go vet ./...

MODERNIZE_CMD = go run golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@v0.18.1

##@modernization
modernize: ##@modernization Apply Go modernization fixes.
	@echo "Running gopls modernize with -fix..."
	$(MODERNIZE_CMD) -test -fix ./...

modernize-check: ##@analysis Check for needed Go modernizations.
	@echo "Checking if code needs modernization..."
	$(MODERNIZE_CMD) -test ./...
