.PHONY: help

HELP_FUN = %help; while (<>) { /^([A-Za-z0-9_-]+)\s*:.*\#\#(?:@([A-Za-z0-9_-]+))?\s(.*)$$/ or next; push @{$$help{$$2 || "other"}}, [$$1, $$3]; $$width = length($$1) if length($$1) > $$width } print "\n"; for $$category (sort keys %help) { print "\e[37m$$category\e[0m\n"; for $$entry (@{$$help{$$category}}) { printf "  \e[33m%-*s\e[0m  \e[32m%s\e[0m\n", $$width, $$entry->[0], $$entry->[1] } }

help: ##@other Show this help.
	@perl -e '$(HELP_FUN)' $(MAKEFILE_LIST)

##@quality
test: ##@quality Run the test suite.
	go test ./...

vet: ##@quality Run Go vet.
	go vet ./...

MODERNIZE_CMD = go run golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@v0.18.1

##@modernization
modernize: modernize-fix ##@modernization Apply Go modernization fixes.

modernize-fix: ##@modernization Apply Go modernization fixes.
	@echo "Running gopls modernize with -fix..."
	$(MODERNIZE_CMD) -test -fix ./...

modernize-check: ##@modernization Check for needed Go modernizations.
	@echo "Checking if code needs modernization..."
	$(MODERNIZE_CMD) -test ./...
