.PHONY: help version bump-patch bump-minor bump-major test test-race lint

SHELL := /bin/sh

help: ## Print available targets
	@echo "Available targets:"
	@echo "  version      Print the latest git tag"
	@echo "  bump-patch   Increment patch version and push tag (e.g. v0.1.0 -> v0.1.1)"
	@echo "  bump-minor   Increment minor version and push tag (e.g. v0.1.1 -> v0.2.0)"
	@echo "  bump-major   Increment major version and push tag (e.g. v0.2.0 -> v1.0.0)"
	@echo "  test         Run go test ./..."
	@echo "  test-race    Run go test -race ./..."
	@echo "  lint         Run go vet ./..."

version: ## Print the latest git tag
	@git describe --tags --abbrev=0 2>/dev/null || echo "no tags"

bump-patch: ## Increment patch version and push annotated tag
	@CURRENT=$$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"); \
	MAJOR=$$(echo "$$CURRENT" | sed 's/^v//' | cut -d. -f1); \
	MINOR=$$(echo "$$CURRENT" | sed 's/^v//' | cut -d. -f2); \
	PATCH=$$(echo "$$CURRENT" | sed 's/^v//' | cut -d. -f3); \
	NEW_PATCH=$$(expr $$PATCH + 1); \
	NEW_TAG="v$${MAJOR}.$${MINOR}.$${NEW_PATCH}"; \
	echo "Bumping patch: $$CURRENT -> $$NEW_TAG"; \
	git tag -a "$$NEW_TAG" -m "$$NEW_TAG"; \
	git push origin "$$NEW_TAG"; \
	echo "Tagged and pushed: $$NEW_TAG"

bump-minor: ## Increment minor version, reset patch to 0, and push annotated tag
	@CURRENT=$$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"); \
	MAJOR=$$(echo "$$CURRENT" | sed 's/^v//' | cut -d. -f1); \
	MINOR=$$(echo "$$CURRENT" | sed 's/^v//' | cut -d. -f2); \
	NEW_MINOR=$$(expr $$MINOR + 1); \
	NEW_TAG="v$${MAJOR}.$${NEW_MINOR}.0"; \
	echo "Bumping minor: $$CURRENT -> $$NEW_TAG"; \
	git tag -a "$$NEW_TAG" -m "$$NEW_TAG"; \
	git push origin "$$NEW_TAG"; \
	echo "Tagged and pushed: $$NEW_TAG"

bump-major: ## Increment major version, reset minor and patch to 0, and push annotated tag
	@CURRENT=$$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0"); \
	MAJOR=$$(echo "$$CURRENT" | sed 's/^v//' | cut -d. -f1); \
	NEW_MAJOR=$$(expr $$MAJOR + 1); \
	NEW_TAG="v$${NEW_MAJOR}.0.0"; \
	echo "Bumping major: $$CURRENT -> $$NEW_TAG"; \
	git tag -a "$$NEW_TAG" -m "$$NEW_TAG"; \
	git push origin "$$NEW_TAG"; \
	echo "Tagged and pushed: $$NEW_TAG"

test: ## Run go test ./...
	go test ./...

test-race: ## Run go test -race ./...
	go test -race ./...

lint: ## Run go vet ./...
	go vet ./...
