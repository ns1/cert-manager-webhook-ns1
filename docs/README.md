## cert-manager-webhook-ns1

This is a helm repository for [cert-manager-webhook-ns1](https://github.com/ns1/cert-manager-webhook-ns1).
See available versions [here](https://github.com/ns1/cert-manager-webhook-ns1/tree/master/docs)

    # add the repository
    helm repo add cert-manager-webhook-ns1 https://ns1.github.io/cert-manager-webhook-ns1
    # make sure we're up to date
    helm repo update
    # take a look at any configuration you might want to set
    helm show values cert-manager-webhook-ns1/cert-manager-webhook-ns1
    # install the chart
    helm install --namespace cert-manager cert-manager-webhook-ns1/cert-manager-webhook-ns1
