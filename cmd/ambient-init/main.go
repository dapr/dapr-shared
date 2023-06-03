package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/dapr/dapr/pkg/injector/sidecar"
	daprutils "github.com/dapr/dapr/utils"

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

var configMapName *string

func main() {
	log.Println("executing dapr-ambient-init")

	configMapName := flag.String("config-map", "dapr-ambient-configmap", "--config-map=value")
	flag.Parse()
	log.Println("config map name: ", *configMapName)

	ctx := context.Background()

	kubeClient := daprutils.GetKubeClient()

	c := NewDaprSidecarClient(kubeClient)

	rootCert, certChain, certKey := c.Get(ctx, LookupEnvOrString(DaprControlPlaneNamespace, DaprSystemNamespace))

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      *configMapName,
			Namespace: namespaceDefault,
		},
		Data: map[string]string{
			daprTrustAnchorsConfigMapKey:   rootCert,
			daprTrustCertChainConfigMapKey: certChain,
			daprTrustCertKeyConfigMapKey:   certKey,
		},
	}

	_, err := kubeClient.CoreV1().ConfigMaps(namespaceDefault).Get(ctx, *configMapName, metav1.GetOptions{})
	if err == nil {
		panic(fmt.Errorf("configmap %s already exists", *configMapName))
	}

	_, err = kubeClient.CoreV1().ConfigMaps(namespaceDefault).Create(ctx, configMap, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
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
