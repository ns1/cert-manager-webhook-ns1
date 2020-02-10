IMAGE_NAME := cert-manager-webhook-ns1
IMAGE_TAG := latest
REPO_NAME := ns1inc

OUT := $(shell pwd)/_out

$(shell mkdir -p "$(OUT)")

.PHONY: all build tag push helm rendered-manifest.yaml

all: ;

build:
	docker build -t "$(IMAGE_NAME):$(IMAGE_TAG)" .

tag:
	docker tag "$(IMAGE_NAME):$(IMAGE_TAG)" "$(REPO_NAME)/$(IMAGE_NAME):$(IMAGE_TAG)"

push:
	docker push "$(REPO_NAME)/$(IMAGE_NAME):$(IMAGE_TAG)"

helm:
	helm package deploy/$(IMAGE_NAME)/ -d docs/
	helm repo index docs --url https://ns1.github.io/cert-manager-webhook-ns1 --merge docs/index.yaml

rendered-manifest.yaml:
	helm template \
		--name $(IMAGE_NAME) \
		--set image.repository=$(IMAGE_NAME) \
		--set image.tag=$(IMAGE_TAG) \
		deploy/$(IMAGE_NAME) > "$(OUT)/rendered-manifest.yaml"
