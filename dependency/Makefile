.PHONY: test

retrieve:
	@cd retrieval; \
	go run . \
		--buildpack-toml-path=$(buildpackTomlPath) \
		--output=$(output)

test:
	@cd test; \
		./run-test --tarballPath $(tarballPath) \
			--version $(version)
