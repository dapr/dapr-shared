package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"github.com/dapr/dapr/pkg/injector/sidecar"
	sentryConsts "github.com/dapr/dapr/pkg/sentry/consts"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	AppRemoteUrl string = "http://remote-url:8080/"
	ProxyPort    string = "8080"
)

func setEnv() {
	daprCPNamespace := os.Getenv("DAPR_CONTROL_PLANE_NAMESPACE")
	if daprCPNamespace == "" {
		panic(fmt.Errorf("DAPR_CONTROL_PLANE_NAMESPACE env var not set"))
	}

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

	trustAnchors, certChain, certKey := sidecar.GetTrustAnchorsAndCertChain(context.TODO(), kubeClient, daprCPNamespace)
	os.Setenv(sentryConsts.TrustAnchorsEnvVar, trustAnchors)
	os.Setenv(sentryConsts.CertChainEnvVar, certChain)
	os.Setenv(sentryConsts.CertKeyEnvVar, certKey)

	trustAnchorsString := []byte(trustAnchors)
	certChainString := []byte(certChain)
	certKeyString := []byte(certKey)
	os.WriteFile("/shared/DAPR_TRUST_ANCHORS", trustAnchorsString, 0644)
	os.WriteFile("/shared/DAPR_CERT_CHAIN", certChainString, 0644)
	os.WriteFile("/shared/DAPR_CERT_KEY", certKeyString, 0644)

}

func main() {

	// add custom flags
	flag.StringVar(&AppRemoteUrl, "app-remote-url", LookupEnvOrString("AMBIENT_APP_REMOTE_URL", AppRemoteUrl), "The remote url to forward app requests to")
	flag.StringVar(&ProxyPort, "proxy-port", LookupEnvOrString("AMBIENT_PROXY_PORT", ProxyPort), "The port where the proxy will listen for requests")

	// Dapr subscription routes orders topic to this route
	http.HandleFunc("/", handleRequest)

	// Add handlers for readiness and liveness endpoints
	http.HandleFunc("/health/{endpoint:readiness|liveness}", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})

	setEnv()

	log.Printf("Starting Dapr Ambient Proxy in Port %s forwarding to remote URL %s", ProxyPort, AppRemoteUrl)
	// Start the server; this is a blocking call
	if err := http.ListenAndServe(fmt.Sprintf(":%s", ProxyPort), nil); err != nil {
		fmt.Errorf("failed to start HTTP proxy: %w", err)
	}
}

func LookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	r.Body = io.NopCloser(bytes.NewReader(body))

	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		fmt.Println(err)
	}
	log.Println(string(requestDump))

	url := fmt.Sprintf("%s%s", AppRemoteUrl, r.RequestURI)
	if !strings.HasPrefix(url, "http") {
		url = fmt.Sprintf("http://%s", url)
	}

	log.Printf("Proxying request to %s", url)

	proxyReq, err := http.NewRequest(r.Method, url, bytes.NewReader(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	proxyReq.Header = make(http.Header)
	for h, val := range r.Header {
		proxyReq.Header[h] = val
	}

	resp, err := http.DefaultClient.Do(proxyReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for h, val := range resp.Header {
		w.Header()[h] = val
	}

	w.WriteHeader(resp.StatusCode)

	log.Printf("Proxied request response code %s - %d", resp.Status, resp.StatusCode)

	_, err = w.Write(body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
