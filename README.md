# ACME webhook for NS1

## Prerequisites

[certmanager](https://github.com/jetstack/cert-manager)

## Installation

Have an NS1 API Key and:

```bash
$ helm install --name cert-manager-webhook-ns1 ./deploy/ns1-webhook \
  --set groupName=<GROUP_NAME> \
  --set clusterIssuer.enabled=true,clusterIssuer.email=<EMAIL_ADDRESS>
```

## Issuer

secret

```bash
$ kubectl -n cert-manager create secret generic ns1-credentials --from-literal=APIKey='Your NS1 API Key'
```

RBAC

```yaml
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRole
metadata:
  name: cert-manager-webhook-ns1:secret-reader
rules:
  - apiGroups:
      - ''
    resources:
      - 'secrets'
    resourceNames:
      - 'ns1-credentials'
    verbs:
      - 'get'
      - 'watch'
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: cert-manager-webhook-ns1:secret-reader
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cert-manager-webhook-ns1:secret-reader
subjects:
  - apiGroup: ""
    kind: ServiceAccount
    name: cert-manager-webhook-ns1
    namespace: default
```

ClusterIssuer

```yaml
apiVersion: certmanager.k8s.io/v1alpha1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: <your email>
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - selector:
        dnsNames:
        - '*.example.com'
      dns01:
        webhook:
          config:
            apiKeyRef:
              key: apiKey
              name: ns1-credentials
            endpoint: 'https://api.nsone.net/v1/'
            ignoreSSL: false
            ttl: 600
          groupName: acme.ns1.net
          solverName: ns1
```

Certificate

```yaml
apiVersion: certmanager.k8s.io/v1alpha1
kind: Certificate
metadata:
  name: wildcard-example-com
spec:
  secretName: wildcard-example-com-tls
  renewBefore: 240h
  dnsNames:
  - '*.example.com'
  issuerRef:
    name: letsencrypt-prod
    kind: ClusterIssuer
```

Ingress

```yaml
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: example-ingress
  namespace: default
  annotations:
    certmanager.k8s.io/cluster-issuer: "letsencrypt-prod"
spec:
  tls:
  - hosts:
    - '*.example.com'
    secretName: wildcard-example-com-tls
  rules:
  - host: demo.example.com
    http:
      paths:
      - path: /
        backend:
          serviceName: backend-service
          servicePort: 80
```

## Development

### Running the test suite

All DNS providers **must** run the DNS01 provider conformance testing suite,
else they will have undetermined behaviour when used with cert-manager.

**It is essential that you configure and run the test suite when creating or
modifying a DNS01 webhook.**

The tests are "live" and require a functioning, DNS-accessible zone, as well as
credentials for the NS1 API. The tests will create (and remove) a TXT record
for the test zone.

Prepare testing environment by editing `test_data/ns1/config.json` and running
the `fetch-test-binaries` script:

```bash
$ scripts/fetch-test-binaries.sh
```

You can run the test suite with:

```bash
$ TEST_ZONE_NAME=example.com. go test .
```
