# NS1 Webhook for Cert Manager

This is a webhook solver for [NS1](http://ns1.com).

## Prerequisites

[certmanager](https://github.com/jetstack/cert-manager)

tested with v0.13.0

## Installation

Install with helm:

```bash
$ helm install --namespace cert-manager --name cert-manager-webhook-ns1 \
  ./deploy/cert-manager-webhook-ns1
```

Issuer

1. Populate a secret with your NS1 API Key:
```bash
$ kubectl --namespace cert-manager create secret generic \
  ns1-credentials --from-literal=apiKey='Your NS1 API Key'
```

2. Add the following resources, something like:
```bash
  kubectl --namespace cert-manager apply -f my_resource.yaml
```

Note that it may make more sense in your setup to use e.g. `ClusterRole` or
`ClusterIssuer` and/or to have more nuanced namespace management.

3. Grants permission for service-account to get the secret:
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
  apiVersion: rbac.authorization.k8s.io/v1
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

4. Creates a staging issuer *Optional*:
```yaml
apiVersion: cert-manager.io/v1alpha2
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

5. Creates a production issuer:
```yaml
apiVersion: cert-manager.io/v1alpha2
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

1. Issues a certificate
```yaml
apiVersion: cert-manager.io/v1alpha2
kind: Certificate
metadata:
  name: example.com
spec:
  commonName: example.com
  dnsNames:
  - example.com
  issuerRef:
    name: letsencrypt-staging
  secretName: example-com-tls
```

### Automatically creating Certificates for Ingress resources

See [this](https://docs.cert-manager.io/en/latest/tasks/issuing-certificates/ingress-shim.html).

### Troubleshooting

Some cert-manager docs that may be helpful:

* About [ACME](https://cert-manager.io/docs/configuration/acme/)
* About [DNS01 Challenges](https://cert-manager.io/docs/configuration/acme/dns01/)
* [Troubleshooting Issuing ACME Certificates](https://cert-manager.io/docs/faq/acme/)

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
