GREEN  := $(shell tput -Txterm setaf 2)
WHITE  := $(shell tput -Txterm setaf 7)
YELLOW := $(shell tput -Txterm setaf 3)
RESET  := $(shell tput -Txterm sgr0)

.PHONY: help

HELP_FUN = %help; while(<>) { if (/^([A-Za-z0-9_-]+)\s*:.*\#\#(?:@([A-Za-z0-9_-]+))?\s(.*)$$/) { push @{$$help{$$2 || "other"}}, [$$1, $$3]; $$width = length($$1) if length($$1) > $$width } } print "\n"; for $$category (sort keys %help) { print "${WHITE}$$category${RESET}\n"; for $$entry (@{$$help{$$category}}) { printf "  ${YELLOW}%-*s${RESET}  ${GREEN}%s${RESET}\n", $$width, $$entry->[0], $$entry->[1] } }

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
