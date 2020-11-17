## Build the MCH functional test image
tests/docker-build:
	@echo "Building $(REGISTRY)/$(IMG)-tests:$(VERSION)"
	docker build . -f integration_tests/build/Dockerfile -t $(REGISTRY)/$(IMG)-tests:$(VERSION)

## Run the installer downstream functional tests
tests/docker-run:
	docker run --network host \
		--volume ~/.kube/config:/opt/.kube/config \
		$(REGISTRY)/$(IMG)-tests:$(VERSION)