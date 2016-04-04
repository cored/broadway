package deployment

import (
	"io/ioutil"
	"os"

	"k8s.io/kubernetes/pkg/client/restclient"
)

var (
	EnvServiceHost string = os.Getenv("KUBERNETES_SERVICE_HOST")
	EnvServicePort string = os.Getenv("KUBERNETES_PORT_443_TCP_PORT")
)

func init() {

}

// IsKubernetesEnv returns true if necessary Kubernetes environment variables
// and files are available
func IsKubernetesEnv() bool {
	files := []string{
		"/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
		"/var/run/secrets/kubernetes.io/serviceaccount/token",
	}

	for _, file := range files {
		fi, err := os.Stat(file)
		if err != nil || !fi.IsDir() {
			return false
		}
	}

	envs := []string{
		EnvServicePort,
		EnvServiceHost,
	}

	for _, env := range envs {
		if env == "" {
			return false
		}
	}

	return true
}

// Config returns a kubernetes configuration
func Config() (*restclient.Config, error) {
	config := LocalConfig()
	if IsKubernetesEnv() {
		var err error
		config, err = KubernetesConfig()
		if err != nil {
			return nil, err
		}
	}

	return config, nil
}

// KubernetesConfig returns Kubernetes configuration for native Kubernetes environment
func KubernetesConfig() (*restclient.Config, error) {
	token, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		return nil, err
	}
	return &restclient.Config{
		Host:        "http://" + EnvServiceHost + ":" + EnvServicePort,
		BearerToken: string(token),
		TLSClientConfig: restclient.TLSClientConfig{
			CAFile: "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
		},
	}, nil
}

// LocalConfig returns a configuration for local development
func LocalConfig() *restclient.Config {
	return &restclient.Config{
		Host:     "http://localhost:8080",
		Insecure: true,
	}
}
