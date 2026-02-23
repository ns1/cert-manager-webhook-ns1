# NS1 Webhook for Cert Manager

> This project is in [MAINTENANCE](https://github.com/ns1/community/blob/master/project_status/MAINTENANCE.md) mode

This is a webhook solver for [NS1](http://ns1.com), for use with cert-manager,
to solve ACME DNS01 challenges.

Tested with kubernetes v1.15.7, v1.16.2, and v1.17.0

## Prerequisites

[helm](https://helm.sh/), for installing charts.

Tested with helm v2.16.1 and v3.0.3

[certmanager](https://cert-manager.io/docs/installation/kubernetes/), the
underlying framework this project plugs into.

Tested with cert-manager v1.19.3

## Installation

1. Install the chart with helm

We have a [helm repo](https://ns1.github.io/cert-manager-webhook-ns1/) set up,
so you can use that, or you can install directly from source:

```bash
$ helm install --namespace cert-manager cert-manager-webhook-ns1 ./deploy/cert-manager-webhook-ns1
```

2. Populate a secret with your NS1 API Key

```bash
$ kubectl --namespace cert-manager create secret generic ns1-credentials --from-literal=apiKey='Your NS1 API Key'
```

3. We need to grant permission for service account to get the secret. Copy the
following and apply with something like:

```bash
$ kubectl --namespace cert-manager apply -f secret_reader.yaml
```

Note that it may make more sense in your setup to use a `ClusterRole` and
`ClusterRoleBinding` here.

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

1. Create Issuer(s), we'll use `letsencrypt` for example. We'll use
`ClusterIssuer` here, which will be available accross namespaces. You may
prefer to use `Issuer`. This is where `NS1` API options are set (`endpoint`,
`ignoreSSL`).

Staging issuer (**optional**):

```yaml
apiVersion: cert-manager.io/v1
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
```

Production issuer:

```yaml
apiVersion: cert-manager.io/v1
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
```

2. Test things by issuing a certificate. This example requests a cert for
`example.com` from the staging issuer, default namespace should be fine:

```yaml
apiVersion: cert-manager.io/v1
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

See cert-manager
[docs](https://docs.cert-manager.io/en/latest/tasks/issuing-certificates/ingress-shim.html)
on "ingress shims".

The gist of it is adding an annotation, and a `tls` section to your Ingress
definition. A simple ingress example is below with pertinent areas bolded. We
use the `ingress-nginx` ingress controller, but it should be the same idea for
any ingress.

You do of course, need to set up an `A` Record in `NS1` connecting the domain
to the external IP of the ingress controller's LoadBalancer service. In the
example below the domain would be `my-app.example.com`.

<pre>
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: my-ingress
  annotations:
    kubernetes.io/ingress.class: "nginx"
    <b>cert-manager.io/cluster-issuer: "letsencrypt-prod"</b> # sets the issuer to use
spec:
  rules:
  - host: my-app.example.com
    http:
      paths:
      - backend:
          serviceName: my-app-service
          servicePort: 80
  <b>tls:
  - hosts:
    - my-app.example.com</b>   # domain name(s) for the certificate
    <b>secretName: my-app-tls</b> # where to store the secret
</pre>

### Troubleshooting

If things aren't working, check the logs in the main `cert-manager` pod first,
they are pretty communicative. Check logs from the other `cert-manager-*` pods
and the `cert-manager-webhook-ns1` pod.

If you've generated a `Certificate` but no `CertificateRequest` is generated,
the main `cert-manager` pod logs should show why any action was skipped.

Since this project is essentially a plugin to `cert-manager`, detailed docs
mainly live in the `cert-manager` project. Here are some specific docs that may
be helpful:

* About [ACME](https://cert-manager.io/docs/configuration/acme/)
* About [DNS01 Challenges](https://cert-manager.io/docs/configuration/acme/dns01/)
* [Troubleshooting Issuing ACME Certificates](https://cert-manager.io/docs/faq/acme/)
* [Securing Ingress Resources](https://cert-manager.io/docs/usage/ingress/)

## Development

### Running the test suite

All DNS providers **must** run the DNS01 provider conformance testing suite,
else they will have undetermined behaviour when used with cert-manager.

**It is essential that you configure and run the test suite when creating or
modifying a DNS01 webhook.**

The tests are "live" and require a functioning, DNS-accessible zone, as well as
credentials for the NS1 API. The tests will create (and remove) a TXT record
for the test zone.

1. See the `README` in `testdata/ns1` and copy/edit the files as needed.

2. Run the tests with `TEST_ZONE_NAME` set to your live, NS1-controlled zone:

```bash
$ TEST_ZONE_NAME=example.com. make test
```

### Maintaining the Docker image and Helm repository

See `Makefile` for commands to build and push the Docker image, and to maintain
the Helm repo.

## Contributions

Pull Requests and issues are welcome. See the
[NS1 Contribution Guidelines](https://github.com/ns1/community/blob/master/Contributing.md)
for more information.
