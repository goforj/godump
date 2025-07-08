#----------------------
# Parse makefile arguments
#----------------------
RUN_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
$(eval $(RUN_ARGS):;@:)

#----------------------
# Silence GNU Make
#----------------------
ifndef VERBOSE
MAKEFLAGS += --no-print-directory
endif

#----------------------
# Terminal
#----------------------

GREEN  := $(shell tput -Txterm setaf 2)
WHITE  := $(shell tput -Txterm setaf 7)
YELLOW := $(shell tput -Txterm setaf 3)
RESET  := $(shell tput -Txterm sgr0)

#------------------------------------------------------------------
# - Add the following 'help' target to your Makefile
# - Add help text after each target name starting with '\#\#'
# - A category can be added with @category
#------------------------------------------------------------------

HELP_FUN = \
	%help; \
	while(<>) { \
		push @{$$help{$$2 // 'options'}}, [$$1, $$3] if /^([a-zA-Z\-]+)\s*:.*\#\#(?:@([a-zA-Z\-]+))?\s(.*)$$/ }; \
		print "\n"; \
		for (sort keys %help) { \
			print "${WHITE}$$_${RESET \
		}\n"; \
		for (@{$$help{$$_}}) { \
			$$sep = " " x (32 - length $$_->[0]); \
			print "  ${YELLOW}$$_->[0]${RESET}$$sep${GREEN}$$_->[1]${RESET}\n"; \
		}; \
		print ""; \
	}

help: ##@other Show this help.
	@perl -e '$(HELP_FUN)' $(MAKEFILE_LIST)

#----------------------
# tool
#----------------------

MODERNIZE_CMD = go run golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@v0.18.1

modernize-fix: ##@tool Run modernize fix
	@echo "Running gopls modernize with -fix..."
	$(MODERNIZE_CMD) -test -fix ./...

modernize-check: ##@tool Run modernize check
	@echo "Checking if code needs modernization..."
	$(MODERNIZE_CMD) -test ./...

linter-run: ##@tool Run Go linter
	@echo "Running linter..."
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1 run -v
	@echo "Linter run complete."

run-all: ##@tool Run all tools
	@echo "Running all tools..."
	make modernize-check
	make linter-run
	@echo "All tools run complete."