retrieve:
	@cd retrieval; \
	go run . \
		--buildpack-toml-path=$(buildpackTomlPath) \
		--output=$(output)


test:
	@cd actions/test; \
		./run-test --tarballPath $(tarballPath) \
			--version $(version)
