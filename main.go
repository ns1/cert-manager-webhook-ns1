package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	extapi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	//"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/jetstack/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/jetstack/cert-manager/pkg/acme/webhook/cmd"
	"github.com/jetstack/cert-manager/pkg/issuer/acme/dns/util"

	ns1API "gopkg.in/ns1/ns1-go.v2/rest"
	ns1DNS "gopkg.in/ns1/ns1-go.v2/rest/model/dns"
)

var GroupName = os.Getenv("GROUP_NAME")

func main() {
	if GroupName == "" {
		panic("GROUP_NAME must be specified")
	}

	// This will register our NS1 DNS provider with the webhook serving
	// library, making it available as an API under the provided GroupName.
	// You can register multiple DNS provider implementations with a single
	// webhook, where the Name() method will be used to disambiguate between
	// the different implementations.
	cmd.RunWebhookServer(GroupName,
		&ns1DNSProviderSolver{},
	)
}

// ns1DNSProviderSolver implements the logic needed to 'present' an ACME
// challenge TXT record. To do so, it implements the
// `github.com/jetstack/cert-manager/pkg/acme/webhook.Solver` interface.
type ns1DNSProviderSolver struct {
	// If a Kubernetes 'clientset' is needed, you must:
	// 1. uncomment the additional `client` field in this structure below
	// 2. uncomment the "k8s.io/client-go/kubernetes" import at the top of the file
	// 3. uncomment the relevant code in the Initialize method below
	// 4. ensure your webhook's service account has the required RBAC role
	//    assigned to it for interacting with the Kubernetes APIs you need.
	//client kubernetes.Clientset

	client *ns1API.Client
}

// ns1DNSProviderConfig is a structure that is used to decode into when
// solving a DNS01 challenge.
// This information is provided by cert-manager, and may be a reference to
// additional configuration that's needed to solve the challenge for this
// particular certificate or issuer.
// This typically includes references to Secret resources containing DNS
// provider credentials, in cases where a 'multi-tenant' DNS solver is being
// created.
// If you do *not* require per-issuer or per-certificate configuration to be
// provided to your webhook, you can skip decoding altogether in favour of
// using CLI flags or similar to provide configuration.
// You should not include sensitive information here. If credentials need to
// be used by your provider here, you should reference a Kubernetes Secret
// resource and fetch these credentials using a Kubernetes clientset.
type ns1DNSProviderConfig struct {
	// These fields will be set by users in the
	// `issuer.spec.acme.dns01.providers.webhook.config` field.

	//Email           string `json:"email"`
	//APIKeySecretRef v1alpha1.SecretKeySelector `json:"apiKeySecretRef"`

	APIKey    string `json:"apiKey"`
	Endpoint  string `json:"endpoint"`
	IgnoreSSL bool   `json:"ignoreSSL"`
	TTL       int    `json:"ttl"`
}

// Name is used as the name for this DNS solver when referencing it on the ACME
// Issuer resource.
// This should be unique **within the group name**, i.e. you can have two
// solvers configured with the same Name() **so long as they do not co-exist
// within a single webhook deployment**.
func (c *ns1DNSProviderSolver) Name() string {
	return "ns1"
}

// Present is responsible for actually presenting the DNS record with the
// DNS provider.
// This method should tolerate being called multiple times with the same value.
// cert-manager itself will later perform a self check to ensure that the
// solver has correctly configured the DNS provider.
func (c *ns1DNSProviderSolver) Present(ch *v1alpha1.ChallengeRequest) error {
	cfg, err := loadConfig(ch.Config)
	if err != nil {
		return err
	}
	fmt.Printf("Decoded configuration %v\n", cfg)

	zone, domain, err := c.parseChallenge(ch)
	if err != nil {
		return err
	}

	return c.createRecord(cfg, zone, domain, ch.Key)
}

// CleanUp should delete the relevant TXT record from the DNS provider console.
// If multiple TXT records exist with the same record name (e.g.
// _acme-challenge.example.com) then **only** the record with the same `key`
// value provided on the ChallengeRequest should be cleaned up.
// This is in order to facilitate multiple DNS validations for the same domain
// concurrently.
func (c *ns1DNSProviderSolver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	cfg, err := loadConfig(ch.Config)
	if err != nil {
		return err
	}
	fmt.Printf("Decoded configuration %v\n", cfg)

	zone, domain, err := c.parseChallenge(ch)
	if err != nil {
		return err
	}

	return c.deleteRecord(cfg, zone, domain, ch.Key)
}

// Initialize will be called when the webhook first starts.
// This method can be used to instantiate the webhook, i.e. initialising
// connections or warming up caches.
// Typically, the kubeClientConfig parameter is used to build a Kubernetes
// client that can be used to fetch resources from the Kubernetes API, e.g.
// Secret resources containing credentials used to authenticate with DNS
// provider accounts.
// The stopCh can be used to handle early termination of the webhook, in cases
// where a SIGTERM or similar signal is sent to the webhook process.
func (c *ns1DNSProviderSolver) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	///// UNCOMMENT THE BELOW CODE TO MAKE A KUBERNETES CLIENTSET AVAILABLE TO
	///// YOUR CUSTOM DNS PROVIDER

	//cl, err := kubernetes.NewForConfig(kubeClientConfig)
	//if err != nil {
	//	return err
	//}
	//
	//c.client = cl

	///// END OF CODE TO MAKE KUBERNETES CLIENTSET AVAILABLE
	return nil
}

// loadConfig is a small helper function that decodes JSON configuration into
// the typed config struct.
func loadConfig(cfgJSON *extapi.JSON) (ns1DNSProviderConfig, error) {
	cfg := ns1DNSProviderConfig{}
	// handle the 'base case' where no configuration has been provided
	if cfgJSON == nil {
		return cfg, nil
	}
	if err := json.Unmarshal(cfgJSON.Raw, &cfg); err != nil {
		return cfg, fmt.Errorf("error decoding solver config: %v", err)
	}

	return cfg, nil
}

func (c *ns1DNSProviderSolver) setClient(cfg ns1DNSProviderConfig) {
	httpClient := &http.Client{}
	if cfg.IgnoreSSL == true {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		httpClient.Transport = tr
	}
	client := ns1API.NewClient(
		httpClient,
		ns1API.SetAPIKey(cfg.APIKey),
		ns1API.SetEndpoint(cfg.Endpoint),
	)
	c.client = client
}

// Create a TXT Record for domain.zone with answer set to DNS challenge key
func (c *ns1DNSProviderSolver) createRecord(
	cfg ns1DNSProviderConfig, zone, domain, answer string,
) error {
	if c.client == nil {
		c.setClient(cfg)
	}
	fmt.Printf("Creating TXT Record for %s.%s\n", domain, zone)

	record := ns1DNS.NewRecord(zone, domain, "TXT")
	record.TTL = cfg.TTL
	record.AddAnswer(ns1DNS.NewTXTAnswer(answer))
	_, err := c.client.Records.Create(record)
	if err != nil {
	  if err != ns1API.ErrRecordExists {
			return err
		}
	}

	fmt.Printf("Created TXT Record: %v\n", record)
	return nil
}

// NS1 API disallows multiple TXT records for the same FQDN, so we don't need
// to check the answer value.
func (c *ns1DNSProviderSolver) deleteRecord(
	cfg ns1DNSProviderConfig, zone, domain, answer string,
) error {
	if c.client == nil {
		c.setClient(cfg)
	}
	fmt.Printf("Deleting TXT Record for %s.%s\n", domain, zone)

	_, err := c.client.Records.Delete(
		zone, fmt.Sprintf("%s.%s", domain, zone), "TXT",
	)
	if err != nil {
	  if err != ns1API.ErrRecordExists {
			return err
		}
	}

	fmt.Printf("Deleted TXT Record\n")
	return nil
}

// Get the zone and domain we are setting from the challenge request
func (c *ns1DNSProviderSolver) parseChallenge(ch *v1alpha1.ChallengeRequest) (
	zone string, domain string, err error,
) {

	zone, err = util.FindZoneByFqdn(ch.ResolvedFQDN, util.RecursiveNameservers)
	if err == nil {
		zone = util.UnFqdn(zone)
	} else {
		return "", "", err
	}

	if idx := strings.Index(ch.ResolvedFQDN, "." + ch.ResolvedZone); idx != -1 {
		domain = ch.ResolvedFQDN[:idx]
	} else {
		domain = util.UnFqdn(ch.ResolvedFQDN)
	}

	return zone, domain, nil
}
