package env

import (
	"log"
	"os"
)

// SlackToken contains the expected Slack custom command token.
var SlackToken string

// ServerHost is passed to gin and configures the listen address of the server
var ServerHost string

// K8sServiceHost is the Kubernetes host
var K8sServiceHost string

// K8sServicePort is the Kubernetes port
var K8sServicePort string

// K8sNamespace is the namespace used by Broadway's deployments
var K8sNamespace string

// EtcdHost is the Etcd host
var EtcdHost string

// LoadEnvs sets env.* variables to their OS-provided value
func LoadEnvs() {
	SlackToken = loadAndWarn("SLACK_VERIFICATION_TOKEN")
	ServerHost = loadAndWarn("HOST")
	K8sServiceHost = loadAndWarn("KUBERNETES_SERVICE_HOST")
	K8sServicePort = loadAndWarn("KUBERNETES_PORT_443_TCP_PORT")
	K8sNamespace = loadAndWarn("KUBERNETES_NAMESPACE")
	EtcdHost = loadAndWarn("ETCD_HOST")
}

func loadAndWarn(key string) (val string) {
	val, ok := os.LookupEnv(key)
	if !ok {
		log.Printf("Environment variable %s unset; defaulting to empty string\n", key)
	}
	return val
}

func init() {
	LoadEnvs()
}
