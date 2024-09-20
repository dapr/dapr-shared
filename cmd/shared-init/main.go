package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/dapr/dapr/pkg/injector/sidecar"
	daprutils "github.com/dapr/dapr/utils"
	"github.com/spf13/cobra"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	daprTrustAnchorsConfigMapKey string = "dapr-trust-anchors"
	namespaceDefault             string = "default"
	DaprSystemNamespace          string = "dapr-system"
	DaprControlPlaneNamespace    string = "DAPR_CONTROL_PLANE_NAMESPACE"
	DaprSharedInstanceNamespace  string = "DAPR_SHARED_INSTANCE_NAMESPACE"
	namespaceFilePath                   = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
)

var configMapName string

func main() {
	log.Println("executing dapr-shared-init")
	rootCmd := NewRootCmd()
	rootCmd.Execute()
}

// LookupEnvOrString tries to look for an environment variable, if found, return it, otherwise find,
// return the default parameter.
func LookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

// DaprSidecarClient represents a client to get Dapr trust bundle.
type DaprSidecarClient interface {
	Get(ctx context.Context, ns string) (string, string, string)
}

type daprSidecarClient struct {
	Clientset *kubernetes.Clientset
}

// NewDaprSidecarClient creates a DaprSidecarClient instance.
func NewDaprSidecarClient(cs *kubernetes.Clientset) DaprSidecarClient {
	return &daprSidecarClient{
		Clientset: cs,
	}
}

// Get gets DaprTrustBundle struct.
func (s *daprSidecarClient) Get(ctx context.Context, ns string) (string, string, string) {
	daprCPNamespace := LookupEnvOrString(DaprControlPlaneNamespace, DaprSystemNamespace)
	return sidecar.GetTrustAnchorsAndCertChain(ctx, s.Clientset, daprCPNamespace)
}

// NewRootCmd
func NewRootCmd() *cobra.Command {

	rootCmd := &cobra.Command{
		Use: "shared-init",
	}
	rootCmd.AddCommand(NewInitCmd())
	return rootCmd
}

// NewInitCmd creates a new *cobra.Command for init command.
func NewInitCmd() *cobra.Command {
	initCmd := &cobra.Command{
		Use: "init",
		Run: func(cmd *cobra.Command, args []string) {
			InitHandler()
		},
	}

	initCmd.PersistentFlags().StringVar(&configMapName, "config-map", "dapr-shared-configmap", "--config-map=value")
	_ = initCmd.MarkPersistentFlagRequired("config-map")

	return initCmd
}

// InitHandler handles the init command.
func InitHandler() {
	ctx := context.Background()

	kubeClient := daprutils.GetKubeClient()

	c := NewDaprSidecarClient(kubeClient)

	rootCert, _, _ := c.Get(ctx, LookupEnvOrString(DaprControlPlaneNamespace, DaprSystemNamespace))

	namespace := getNamespace()
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: namespace,
		},
		Data: map[string]string{
			daprTrustAnchorsConfigMapKey: rootCert,
		},
	}

	_, err := kubeClient.CoreV1().ConfigMaps(namespace).Get(ctx, configMapName, metav1.GetOptions{})
	if err == nil {
		err := kubeClient.CoreV1().ConfigMaps(namespace).Delete(ctx, configMapName, metav1.DeleteOptions{})
		if err != nil {
			panic(err)
		}
	}

	_, err = kubeClient.CoreV1().ConfigMaps(namespace).Create(ctx, configMap, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
}

// getNamespace retrieves the Kubernetes namespace from the service account file or an environment variable.
func getNamespace() string {
	// Read the namespace file
	bytes, err := os.ReadFile(namespaceFilePath)
	if err == nil {
		namespace := string(bytes)
		// Trim any whitespace
		namespace = strings.TrimSpace(namespace)
		if namespace != "" {
			return namespace
		}
	}

	// Fall back to environment variable or default value
	namespace := LookupEnvOrString(DaprSharedInstanceNamespace, namespaceDefault)

	return namespace
}
