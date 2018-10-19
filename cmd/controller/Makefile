.PHONY: image

PREFIX = welmoki
NAME = bakku-app-controller
TAG = 0.0.1
DEV_TAG = $(TAG)-dev

image: ## Build image
	docker build -t "$(PREFIX)/$(NAME):$(TAG)" -f ./Dockerfile ../..

push: image ## Push image (should be used only by ci)
	docker push "$(PREFIX)/$(NAME):$(TAG)"

image-dev: ## Make development image
	docker build -t "$(PREFIX)/$(NAME):$(DEV_TAG)" -f ./Dockerfile ../..

push-dev: image-dev ## Push development image
	docker push "$(PREFIX)/$(NAME):$(DEV_TAG)"

.PHONY: help
help:  ## Show help messages for make targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[32m%-30s\033[0m %s\n", $$1, $$2}'
