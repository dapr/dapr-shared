package ambient

import (
	"context"
	"fmt"
	"log"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// K8SClient represents a simple client for Kubernetes API.
type K8SClient interface {
	CreateConfigMap(ctx context.Context, ns string, data map[string]string) error
}

type k8SClient struct {
	Clientset *kubernetes.Clientset
}

// NewK8SClient creates a K8SClient instance.
func NewK8SClient(k *kubernetes.Clientset) K8SClient {
	return &k8SClient{
		Clientset: k,
	}
}

// CreateConfigMap creates a new ConfigMap resource on Kubernetes cluster.
func (k *k8SClient) CreateConfigMap(ctx context.Context, ns string, data map[string]string) error {

	var configMapName string = "dapr-ambient-configmap"
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: ns,
		},
		Data: data,
	}

	_, err := k.Clientset.CoreV1().ConfigMaps(ns).Get(ctx, configMapName, metav1.GetOptions{})
	if err == nil {
		log.Println(fmt.Printf("configmap %s already exists", configMapName))
		return err
	}

	_, err = k.Clientset.CoreV1().ConfigMaps(ns).Create(ctx, configMap, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}

	return nil
}

// NewClientset creates a new kubernetes.Clientset instance.
func NewClientset() *kubernetes.Clientset {
	kubeConfigLoader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{})

	restConfig, err := kubeConfigLoader.ClientConfig()
	if err != nil {
		panic(err)
	}

	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		panic(err)
	}

	return kubeClient
}
