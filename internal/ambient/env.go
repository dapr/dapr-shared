package ambient

import (
	"os"
)

var (
	// Environment variable to identify the dapr control plane namespace.
	DaprControlPlaneNamespace = "DAPR_CONTROL_PLANE_NAMESPACE"

	// Default dapr namespace.
	DaprControlPlaneDefaultNamespace = "dapr-system"

	// Default Kubernetes namespace.
	DefaultNamespace = "default"
)

// LookupEnvOrString tries to look for an environment variable, if found, return it, otherwise find,
// return the default parameter.
func LookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}
