PKG_NAME := ghcr.io/qmaru
IMG_NAME := nxtrace
PLATFORM := linux/amd64

help: # Show help
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk -F':' '/^[0-9a-zA-Z\-\_]+:.*#/ {cmd=$$1;msg=split($$2,m,"#");msg=split($$2,m,"#");printf "%-10s%-10s\n",cmd,m[2];}' $(MAKEFILE_LIST)

test: check # Build a local tests
	docker buildx build --load -t test .

build: check # Build and Push
	docker buildx build --push \
		-t $(PKG_NAME)/$(IMG_NAME):go \
		--platform=$(PLATFORM) \
		.

output_api: check # Build api bin
	docker build --target=build-api --output=. .

output_core: check # Build core bin
	docker build --target=build-core --output=. .

.PHONY: check
check:
	/etc/init.d/docker status
