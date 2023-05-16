package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	chi "github.com/go-chi/chi"

	dapr "github.com/dapr/go-sdk/client"
)

var (
	port            string
	daprAmbientPort string
	daprAmbientHost string
	daprClient      dapr.Client
	stateStoreName  string
)

func init() {
	// init variables
	port = GetenvOrDefault("APP_PORT", "8080")
	stateStoreName = GetenvOrDefault("STATE_STORE_NAME", "sample-state")
	daprAmbientPort = GetenvOrDefault("DAPR_AMBIENT_PORT", "50001")
	daprAmbientHost = GetenvOrDefault("DAPR_AMBIENT_HOST", "127.0.0.1")

}

func TryConnect() {
	var err error
	daprClient, err = dapr.NewClientWithAddress(net.JoinHostPort(daprAmbientHost, daprAmbientPort))
	if err != nil {
		panic(err)
	}
}

func main() {
	// create chi router
	r := chi.NewRouter()

	// create handler for /states
	r.Get("/state", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		result, err := daprClient.GetState(ctx, stateStoreName, "values", nil)
		if err != nil {
			panic(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response, err := json.Marshal(result)
		if err != nil {
			panic(err)
		}
		w.Write(response)
	})

	r.HandleFunc("/health/{endpoint:readiness|liveness}", func(w http.ResponseWriter, r *http.Request) {
		log.Println("receiving request from dapr-ambient proxy")
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})

	// try connect to daprd daemonset
	go connectWithRetry(10, 5*time.Second)

	http.ListenAndServe(fmt.Sprintf(":%s", port), r)
}

func GetenvOrDefault(name, defaultValue string) string {
	envValue := os.Getenv(name)
	if envValue != "" {
		return envValue
	}
	return defaultValue
}

func connectWithRetry(maxRetries int, retryInterval time.Duration) error {
	var err error
	for i := 0; i < maxRetries; i++ {

		daprClient, err = dapr.NewClientWithAddress(net.JoinHostPort(daprAmbientHost, daprAmbientPort))
		if err == nil {
			return nil
		}
		time.Sleep(retryInterval)
	}

	return fmt.Errorf("was not possible to connect with daprd, after %d tries", maxRetries)
}
