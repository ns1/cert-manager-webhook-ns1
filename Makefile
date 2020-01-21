IMAGE_NAME := cert-manager-webhook-ns1
IMAGE_TAG := latest
REPO_NAME := ns1inc

OUT := $(shell pwd)/_out

$(shell mkdir -p "$(OUT)")

verify:
	go test -v .

build:
	docker build -t "$(IMAGE_NAME):$(IMAGE_TAG)" .

tag:
	docker tag "$(IMAGE_NAME):$(IMAGE_TAG)" "$(REPO_NAME)/$(IMAGE_NAME):$(IMAGE_TAG)"

push:
	docker push "$(REPO_NAME)/$(IMAGE_NAME):$(IMAGE_TAG)"

.PHONY: rendered-manifest.yaml
rendered-manifest.yaml:
	helm template \
		--name $(IMAGE_NAME) \
		--set image.repository=$(IMAGE_NAME) \
		--set image.tag=$(IMAGE_TAG) \
		deploy/$(IMAGE_NAME) > "$(OUT)/rendered-manifest.yaml"
