# ACME webhook for NS1

## Installation

## Issuer

```bash
$ kubectl -n cert-manager create secret generic ns1-credentials --from-literal=APIKey='Your NS1 API Key'
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
