retrieve:
	@cd retrieval; \
	go run . \
		--buildpack-toml-path=$(buildpackTomlPath) \
		--output=$(output)


test:
	@cd tests; \
		./run-test --tarballPath $(tarballPath) \
			--version $(version)
