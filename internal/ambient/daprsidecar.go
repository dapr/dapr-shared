package ambient

import (
	"context"

	"github.com/dapr/dapr/pkg/injector/sidecar"
	"k8s.io/client-go/kubernetes"
)

// DaprSidecarClient represents a client to get Dapr trust bundle.
type DaprSidecarClient interface {
	Get(ctx context.Context, ns string) DaprTrustBundle
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
func (s *daprSidecarClient) Get(ctx context.Context, ns string) DaprTrustBundle {
	daprCPNamespace := LookupEnvOrString(DaprControlPlaneNamespace, DaprControlPlaneDefaultNamespace)
	trustAnchors, certChain, certKey := sidecar.GetTrustAnchorsAndCertChain(ctx, s.Clientset, daprCPNamespace)
	return DaprTrustBundle{
		TrustAnchors: trustAnchors,
		CertChain:    certChain,
		CertKey:      certKey,
	}
}
