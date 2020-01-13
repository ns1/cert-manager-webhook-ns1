#!/usr/bin/env bash

IMAGE_NAME="ns1/cert-manager-webhook-ns1"
IMAGE_TAG="latest"

OUT=`pwd`/deploy

helm template \
        --name cert-manager-webhook-ns1 \
        --set image.repository=$IMAGE_NAME \
        --set image.tag=$IMAGE_TAG \
        deploy/ns1-webhook > "$OUT/rendered-manifest.yaml"
