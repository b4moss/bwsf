# GitHub repository ruleset helpers.
# Real work lives in scripts/; this Makefile is a thin wrapper.

SHELL := /bin/bash
.DEFAULT_GOAL := help

ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
SCRIPTS := $(ROOT_DIR)/scripts

REPO ?=
BRANCH ?= main
VISIBILITY ?= public
CREATE_FLAGS ?=

.PHONY: help create apply check

help:
	@printf '%s\n' \
		'Targets:' \
		'' \
		'  make create REPO=OWNER/NAME [VISIBILITY=public] [CREATE_FLAGS="--clone"]' \
		'      Create a GitHub repo and apply rulesets.' \
		'' \
		'  make apply REPO=OWNER/NAME' \
		'      Apply/update rulesets on an existing repo.' \
		'' \
		'  make check REPO=OWNER/NAME [BRANCH=main]' \
		'      List rulesets and check which rules apply to BRANCH.' \
		'' \
		'Notes:' \
		'  - GitHub Free (org): rulesets work on public repos only.' \
		'  - Requires gh (repo admin) and jq.'

create:
	@if [[ -z "$(REPO)" ]]; then \
		echo "error: REPO=OWNER/NAME is required" >&2; \
		echo "example: make create REPO=my-org/new-app" >&2; \
		exit 1; \
	fi
	@$(SCRIPTS)/create-repo-with-rulesets.sh "$(REPO)" "--$(VISIBILITY)" $(CREATE_FLAGS)

apply:
	@if [[ -z "$(REPO)" ]]; then \
		echo "error: REPO=OWNER/NAME is required" >&2; \
		echo "example: make apply REPO=my-org/existing-app" >&2; \
		exit 1; \
	fi
	@$(SCRIPTS)/apply-rulesets.sh --repo "$(REPO)"

check:
	@if [[ -z "$(REPO)" ]]; then \
		echo "error: REPO=OWNER/NAME is required" >&2; \
		echo "example: make check REPO=my-org/existing-app BRANCH=main" >&2; \
		exit 1; \
	fi
	@echo "== ruleset list =="
	@gh ruleset list -R "$(REPO)"
	@echo
	@echo "== ruleset check $(BRANCH) =="
	@gh ruleset check "$(BRANCH)" -R "$(REPO)"
