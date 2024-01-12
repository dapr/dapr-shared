package main

import (
	"context"
	"log"
	"os"

	"github.com/dapr/dapr/pkg/injector/sidecar"
	daprutils "github.com/dapr/dapr/utils"
	"github.com/spf13/cobra"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	daprTrustAnchorsConfigMapKey   string = "dapr-trust-anchors"
	daprTrustCertChainConfigMapKey string = "dapr-cert-chain"
	daprTrustCertKeyConfigMapKey   string = "dapr-cert-key"
	namespaceDefault               string = "default"
	DaprSystemNamespace            string = "dapr-system"
	DaprControlPlaneNamespace      string = "DAPR_CONTROL_PLANE_NAMESPACE"
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

	rootCert, certChain, certKey := c.Get(ctx, LookupEnvOrString(DaprControlPlaneNamespace, DaprSystemNamespace))

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: namespaceDefault,
		},
		Data: map[string]string{
			daprTrustAnchorsConfigMapKey:   rootCert,
			daprTrustCertChainConfigMapKey: certChain,
			daprTrustCertKeyConfigMapKey:   certKey,
		},
	}

	_, err := kubeClient.CoreV1().ConfigMaps(namespaceDefault).Get(ctx, configMapName, metav1.GetOptions{})
	if err == nil {
		err := kubeClient.CoreV1().ConfigMaps(namespaceDefault).Delete(ctx, configMapName, metav1.DeleteOptions{})
		if err != nil {
			panic(err)
		}
	}

	_, err = kubeClient.CoreV1().ConfigMaps(namespaceDefault).Create(ctx, configMap, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
}
