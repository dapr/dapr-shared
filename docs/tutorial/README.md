# Dapr Shared Getting Started Tutorial

This tutorial follows up on the Dapr Hello Kubernetes tutorial that can be found here in the Dapr Quickstart repo: 

https://github.com/dapr/quickstarts/tree/master/tutorials/hello-kubernetes

Instead of deploying Dapr as a sidecar we are going to use Dapr Shared instances. 


## Prerequisites and Installation

Before proceeding make sure that you have the necessary tools installed on your system. We will create a local KinD cluster to install Dapr, some applications, and an instance of Dapr Shared.

To get started, make sure you have the following CLIs installed:

- [Docker](https://www.docker.com/)

- [KinD (Kubernetes in Docker)](https://kind.sigs.k8s.io/docs/user/quick-start/)

- [kubectl](https://kubernetes.io/docs/tasks/tools/)

- [Helm](https://helm.sh/docs/intro/install/)


## Creating a local Kubernetes cluster with KinD: 

Create a Kubernetes cluster with KinD defaults by running the following command:

```bash
kind create cluster
```

Next install a version of the Dapr control plane into the cluster:

```
helm repo add dapr https://dapr.github.io/helm-charts/
helm repo update
helm upgrade --install dapr dapr/dapr \
--version=1.13.2 \
--namespace dapr-system \
--create-namespace \
--wait
```
Or use the Dapr CLI to install the latest version with the following command

`dapr init -k`
## Running the Hello example

We will be using the Dapr Statestore API and for that we will install a Redis instance into our cluster using Helm: 

```shell
helm install redis oci://registry-1.docker.io/bitnamicharts/redis --version 17.11.3 --set "architecture=standalone" --set "master.persistence.size=1Gi"
```

Once Redis is installed we can deploy our application workloads, including the Statestore component by running: 

```shell
kubectl apply -f deploy/
```

This creates a Statestore Dapr Component, a Node application and a Python application. 

If you inspect the `deploy/node.yaml` and `deploy/python.yaml` files you see that both are define two environment variables: 

```yaml
        - name: DAPR_HTTP_ENDPOINT
          value: http://nodeapp-dapr.default.svc.cluster.local:3500
        - name: DAPR_GRPC_ENDPOINT
          value: http://nodeapp-dapr.default.svc.cluster.local:50001
```

These two environment variables let the Dapr SDK know where the Dapr endpoints are hosted (usually for the sidecar these are located on `localhost`).

Because these workloads are not annotated with Dapr annotations, the Dapr Control Plane will not inject the Dapr Sidecar, instead we will create two instances of the Dapr Shared Helm Chart for our services to use.


## Creating two Dapr Shared instances for our services

For each application service that needs to talk to the Dapr APIs we need to deploy a new Dapr Shared instance. Each instance have a one to one relationship with Dapr Application Ids. 

Let's create a new Dapr Shared instance for the Node (`nodeapp`) application: 

```sh
helm install nodeapp-shared oci://docker.io/daprio/dapr-shared-chart --set shared.appId=nodeapp
```

Notice that the `shared.appId` is the name used for the endpoint variables defined in the previous section. 

Let's do the same for the Python application: 

```sh
helm install pythonapp-shared oci://docker.io/daprio/dapr-shared-chart --set shared.appId=pythonapp
```

Once both Dapr Shared instances are up, the application should be able to connect to the Dapr APIs. You can validate this by interacting with the `nodeapp` by running: 

```
kubectl port-forward service/nodeapp 8080:80
```

And then sending a request to place a new order: 

```shell
curl --request POST --data "@sample.json" --header Content-Type:application/json http://localhost:8080/neworder
```

Validate the order has been persisted: 

```shell
curl http://localhost:8080/order
```

Expected Output: 
```
{ "orderId": "42" }
```

## Get involved

If you want to contribute to Dapr Shared, please get in touch, create an issue, or submit a Pull Request. 
You can also check the Project Roadmap to see what is coming or to find out how you can help us get the next version done. 
