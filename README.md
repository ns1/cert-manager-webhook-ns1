# NS1 Webhook for Cert Manager

This is an webhook solver for [NS1](http://ns1.com), for use with cert-manager,
to solve ACME DNS01 challenges.

Tested with kubernetes v16 and v17

## Prerequisites

[certmanager](https://cert-manager.io/docs/installation/kubernetes/)

Tested with cert-manager v0.13.0

## Installation

1. Install the chart with helm (tested vith v2.16.1):

Note: on kubernetes v17, the `extension-apiserver-authentication-reader` role
has the needed permissions out of the box. On earlier versions, the role may
not have sufficient permissions to manage `configmaps`. Before installing this
chart, edit the role and verify/ensure that the role has `get`, `list`, and
`watch` verbs set on the `configmaps` resource.

```bash
$ helm install --namespace cert-manager --name cert-manager-webhook-ns1 ./deploy/cert-manager-webhook-ns1
```

2. Populate a secret with your NS1 API Key and grant permission for service
account to get the secret. Note that it may make more sense in your setup to
use a `ClusterRole`.
```bash
$ kubectl --namespace cert-manager create secret generic \
  ns1-credentials --from-literal=apiKey='Your NS1 API Key'
```

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

## Configuration

You'll need to edit and apply some resources, with something like:
```bash
$ kubectl --namespace cert-manager apply -f my_resource.yaml
```
Note that we use the `cert-manager` namespace, but it may make more sense in
your setup to hame more nuanced namespace management.

2. Create Issuer(s), we'll use `letsencrypt` for example. We'll use
`ClusterIssuer` here, which will be available accross namespaces. You may
prefer to use `Issuer`.

Staging issuer (**optional**):
```yaml
apiVersion: cert-manager.io/v1alpha2
kind: ClusterIssuer
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
            # Replace this with your NS1 API endpoint if not using "managed"
            endpoint: "https://api.nsone.net/v1/"
            ignoreSSL: false
            ttl: 600
```

Production issuer:
```yaml
apiVersion: cert-manager.io/v1alpha2
kind: ClusterIssuer
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
            # Replace this with your NS1 API endpoint if not using "managed"
            endpoint: "https://api.nsone.net/v1/"
            ignoreSSL: false
            ttl: 600
```

2. Test things by issuing a certificate. This example requests a cert for
`example.com` from the staging issuer:
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

After a minute or two, the `Certificate` should show as `Ready`. If not, you
can follow the resource chain from `Certificate` to `CertificateRequest` and on
down until you see a useful error message.

### Automatically creating Certificates for Ingress resources

See [this](https://docs.cert-manager.io/en/latest/tasks/issuing-certificates/ingress-shim.html).

The gist of it is adding some annotations and a `tls` section to your ingress
definition, e.g:
```bash
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: my-ingress
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod" # sets the issuer to use
spec:
  rules:
  - host: mp-app.example.com
    http:
      paths:
      - backend:
          serviceName: my-app-service
          servicePort: 80
  tls:
  - hosts:
    - my-app.example.com   # domain name(s) for the certificate
    secretName: my-app-tls # where to store the secret
```

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

### Maintaining the Docker image

See `Makefile` for commands to build and push the Docker image.
