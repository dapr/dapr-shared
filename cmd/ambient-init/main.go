package main

import (
	"context"
	"log"

	"github.com/salaboy/dapr-ambient/internal/ambient"
)

func main() {
	log.Println("executing dapr-ambient-init")

	ctx := context.Background()

	cs := ambient.NewClientset()

	c := ambient.NewDaprSidecarClient(cs)

	trustBundle := c.Get(ctx, ambient.LookupEnvOrString(ambient.DaprControlPlaneNamespace, ambient.DaprControlPlaneDefaultNamespace))

	k := ambient.NewK8SClient(cs)

	k.CreateConfigMap(ctx, ambient.DefaultNamespace, trustBundle.ToMap())
}
