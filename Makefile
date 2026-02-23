IMAGE_NAME := cert-manager-webhook-ns1
IMAGE_TAG := latest
REPO_NAME := ns1inc

OUT := $(shell pwd)/_out

$(shell mkdir -p "$(OUT)")

.PHONY: all build tag push helm rendered-manifest.yaml

all: ;

# When Go code changes, we need to update the Docker image
build:
	docker build -t "$(IMAGE_NAME):$(IMAGE_TAG)" .

tag:
	docker tag "$(IMAGE_NAME):$(IMAGE_TAG)" "$(REPO_NAME)/$(IMAGE_NAME):$(IMAGE_TAG)"

push:
	docker push "$(REPO_NAME)/$(IMAGE_NAME):$(IMAGE_TAG)"


# When helm chart changes, we need to publish to the repo (/docs/):
#
# Ensure version is updated in Chart.yaml
# Run `make helm`
# Check and commit the resuls, including the tar.gz
helm:
	helm package deploy/$(IMAGE_NAME)/ -d docs/
	helm repo index docs --url https://ns1.github.io/cert-manager-webhook-ns1 --merge docs/index.yaml

rendered-manifest.yaml:
	helm template \
		$(IMAGE_NAME) \
		--set image.repository=$(IMAGE_NAME) \
		--set image.tag=$(IMAGE_TAG) \
		deploy/$(IMAGE_NAME) > "$(OUT)/rendered-manifest.yaml"


GO ?= $(shell which go)
OS ?= $(shell $(GO) env GOOS)
ARCH ?= $(shell $(GO) env GOARCH)
ENVTEST_K8S_VERSION=1.35.0
OUT := $(shell pwd)/_out


test: setup-envtest
	TEST_ASSET_ETCD=$(LOCALBIN)/k8s/$(ENVTEST_K8S_VERSION)-$(OS)-$(ARCH)/etcd \
	TEST_ASSET_KUBE_APISERVER=$(LOCALBIN)/k8s/$(ENVTEST_K8S_VERSION)-$(OS)-$(ARCH)/kube-apiserver \
	TEST_ASSET_KUBECTL=$(LOCALBIN)/k8s/$(ENVTEST_K8S_VERSION)-$(OS)-$(ARCH)/kubectl \
	$(GO) test -v .

.PHONY: clean
clean:
	chmod -R u+w $(LOCALBIN) $(OUT) 2>/dev/null || true
	rm -rf $(LOCALBIN) $(OUT)

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p "$(LOCALBIN)"

## Tool Binaries

ENVTEST ?= $(LOCALBIN)/setup-envtest

#ENVTEST_VERSION is the version of controller-runtime release branch to fetch the envtest setup script (i.e. release-0.20)
ENVTEST_VERSION ?= $(shell v='$(call gomodver,sigs.k8s.io/controller-runtime)'; \
  [ -n "$$v" ] || { echo "Set ENVTEST_VERSION manually (controller-runtime replace has no tag)" >&2; exit 1; }; \
  printf '%s\n' "$$v" | sed -E 's/^v?([0-9]+)\.([0-9]+).*/release-\1.\2/')

#ENVTEST_K8S_VERSION is the version of Kubernetes to use for setting up ENVTEST binaries (i.e. 1.31)
ENVTEST_K8S_VERSION ?= $(shell v='$(call gomodver,k8s.io/api)'; \
  [ -n "$$v" ] || { echo "Set ENVTEST_K8S_VERSION manually (k8s.io/api replace has no tag)" >&2; exit 1; }; \
  printf '%s\n' "$$v" | sed -E 's/^v?[0-9]+\.([0-9]+).*/1.\1/')

.PHONY: setup-envtest
setup-envtest: envtest ## Download the binaries required for ENVTEST in the local bin directory.
	@echo "Setting up envtest binaries for Kubernetes version $(ENVTEST_K8S_VERSION)..."
	@"$(ENVTEST)" use $(ENVTEST_K8S_VERSION) --bin-dir "$(LOCALBIN)" -p path || { \
		echo "Error: Failed to set up envtest binaries for version $(ENVTEST_K8S_VERSION)."; \
		exit 1; \
	}

.PHONY: envtest
envtest: $(ENVTEST) ## Download setup-envtest locally if necessary.
$(ENVTEST): $(LOCALBIN)
	$(call go-install-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest,$(ENVTEST_VERSION))

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f "$(1)-$(3)" ] && [ "$$(readlink -- "$(1)" 2>/dev/null)" = "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
rm -f "$(1)" ;\
GOBIN="$(LOCALBIN)" go install $${package} ;\
mv "$(LOCALBIN)/$$(basename "$(1)")" "$(1)-$(3)" ;\
} ;\
ln -sf "$$(realpath "$(1)-$(3)")" "$(1)"
endef

define gomodver
$(shell go list -m -f '{{if .Replace}}{{.Replace.Version}}{{else}}{{.Version}}{{end}}' $(1) 2>/dev/null)
endef