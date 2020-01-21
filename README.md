# NS1 Webhook for Cert Manager

This is a webhook solver for [NS1](http://ns1.com).

## Prerequisites

[certmanager](https://github.com/jetstack/cert-manager)

## Installation

Install with helm:

```bash
$ helm install --name cert-manager-webhook-ns1 ./deploy/cert-manager-ns1-webhook
```

Issuer

1. Populate secret with your NS1 API Key:

```bash
$ kubectl --namespace cert-manager create secret generic \
  ns1-credentials --from-literal=APIKey='Your NS1 API Key'
```

2. Grant permission for service-account to get the secret
```yaml
  apiVersion: rbac.authorization.k8s.io/v1
  kind: Role
  metadata:
    name: cert-manager-webhook-ns1:secret-reader
  rules:
  - apiGroups: [""]
    resources: ["secrets"]
    resourceNames: ["ns1-credentials"]
    verbs: ["get", "watch"]
  ---
  apiVersion: rbac.authorization.k8s.io/v1beta1
  kind: RoleBinding
  metadata:
    name: cert-manager-webhook-ns1:secret-reader
  roleRef:
    apiGroup: rbac.authorization.k8s.io
    kind: Role
    name: cert-manager-webhook-ns1:secret-reader
  subjects:
    - apiGroup: ""
      kind: ServiceAccount
      name: cert-manager-webhook-ns1
```

4. Create a staging issuer *Optional*
```yaml
apiVersion: certmanager.k8s.io/v1alpha1
kind: Issuer
metadata:
  name: letsencrypt-staging
spec:
  acme:
    # The ACME server URL
    server: https://acme-staging-v02.api.letsencrypt.org/directory

    # Email address used for ACME registration
    email: user@example.com # REPLACE THIS WITH YOUR EMAIL!!!

    # Name of a secret used to store the ACME account private key
    privateKeySecretRef:
      name: letsencrypt-staging

    solvers:
    - dns01:
        webhook:
          groupName: acme.nsone.net
          solverName: ns1
          config:
            apiKeySecretRef:
              key: apiKey
              name: ns1-credentials
            endpoint: "https://api.nsone.net/v1/"
            ignoreSSL: false
            ttl: 600
```

5. Create a production issuer
```yaml
apiVersion: certmanager.k8s.io/v1alpha1
kind: Issuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    # The ACME server URL
    server: https://acme-v02.api.letsencrypt.org/directory

    # Email address used for ACME registration
    email: user@example.com # REPLACE THIS WITH YOUR EMAIL!!!

    # Name of a secret used to store the ACME account private key
    privateKeySecretRef:
      name: letsencrypt-prod

    solvers:
    - dns01:
        webhook:
          groupName: acme.nsone.net
          solverName: ns1
          config:
            apiKeySecretRef:
              key: apiKey
              name: ns1-credentials
            endpoint: "https://api.nsone.net/v1/"
            ignoreSSL: false
            ttl: 600
```

## Certificate

1. Issue a certificate
```yaml
apiVersion: certmanager.k8s.io/v1alpha1
kind: Certificate
metadata:
  name: example-com
spec:
  commonName: example-com
  dnsNames:
  - example.com
  issuerRef:
    name: letsencrypt-staging
  secretName: example-com-tls
```

### Automatically creating Certificates for Ingress resources

See [this](https://docs.cert-manager.io/en/latest/tasks/issuing-certificates/ingress-shim.html).

## Development

### Running the test suite

All DNS providers **must** run the DNS01 provider conformance testing suite,
else they will have undetermined behaviour when used with cert-manager.

**It is essential that you configure and run the test suite when creating or
modifying a DNS01 webhook.**

The tests are "live" and require a functioning, DNS-accessible zone, as well as
credentials for the NS1 API. The tests will create (and remove) a TXT record
for the test zone.

1. Prepare testing environment by running the `fetch-test-binaries` script:

```bash
$ scripts/fetch-test-binaries.sh
```

2. See the `README` in `test_data/` and copy/edit the files as needed.

3. Run the tests with `TEST_ZONE_NAME` set to your live, NS1-controlled zone:

```bash
$ TEST_ZONE_NAME=example.com. go test .
```
