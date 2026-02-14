.PHONY: release tag check-clean check-upstream check-tag-absent check-version-newer print-latest

# Override on CLI: make release VERSION=v0.1.1
VERSION ?= v0.1.0

# -------- helpers --------

check-clean:
	@git diff --quiet || (echo "ERROR: working tree has uncommitted changes"; exit 1)
	@git diff --cached --quiet || (echo "ERROR: index has staged but uncommitted changes"; exit 1)

check-upstream:
	@branch=$$(git rev-parse --abbrev-ref HEAD); \
	upstream=$$(git rev-parse --abbrev-ref --symbolic-full-name @{u} 2>/dev/null || true); \
	if [ -z "$$upstream" ]; then echo "ERROR: no upstream tracking branch set"; exit 1; fi; \
	ahead=$$(git rev-list --count $$upstream..HEAD); \
	if [ "$$ahead" -ne 0 ]; then echo "ERROR: you have $$ahead unpushed commit(s)"; exit 1; fi

check-tag-absent:
	@# fail if tag exists locally
	@if git rev-parse -q --verify "refs/tags/$(VERSION)" >/dev/null; then \
		echo "ERROR: tag $(VERSION) already exists locally"; exit 1; \
	fi
	@# fail if tag exists on origin
	@if git ls-remote --tags origin "refs/tags/$(VERSION)" | grep -q "$(VERSION)"; then \
		echo "ERROR: tag $(VERSION) already exists on origin"; exit 1; \
	fi

print-latest:
	@latest=$$(git tag -l 'v[0-9]*.[0-9]*.[0-9]*' --sort=-v:refname | head -n1); \
	echo "$${latest:-<none>}"

check-version-newer:
	@# Ensure VERSION is a valid semver tag of form vMAJOR.MINOR.PATCH
	@echo "$(VERSION)" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+$$' || \
		(echo "ERROR: VERSION must look like vMAJOR.MINOR.PATCH (e.g. v0.1.0)"; exit 1)

	@latest=$$(git tag -l 'v[0-9]*.[0-9]*.[0-9]*' --sort=-v:refname | head -n1); \
	if [ -z "$$latest" ]; then \
		echo "No previous semver tags found; allowing $(VERSION)"; exit 0; \
	fi; \
	if [ "$$latest" = "$(VERSION)" ]; then \
		echo "ERROR: VERSION equals latest tag ($$latest)"; exit 1; \
	fi; \
	highest=$$(printf "%s\n%s\n" "$$latest" "$(VERSION)" | sort -V | tail -n1); \
	if [ "$$highest" != "$(VERSION)" ]; then \
		echo "ERROR: VERSION $(VERSION) is not newer than latest tag $$latest"; exit 1; \
	fi; \
	echo "Latest tag: $$latest -> new tag OK: $(VERSION)"

# -------- targets --------

tag: check-clean check-upstream check-version-newer check-tag-absent
	@echo ">>> Creating git tag $(VERSION)"
	@git tag "$(VERSION)"
	@echo ">>> Pushing tag $(VERSION)"
	@git push origin "$(VERSION)"

release: tag
	@echo ">>> Release tag pushed: $(VERSION)"

