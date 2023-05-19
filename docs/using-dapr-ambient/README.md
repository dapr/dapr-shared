# Dapr Ambient with KinD
This tutorial provides step-by-step instructions for installing Dapr Ambient and configuring a set of applications to work with it.

## Prerequisites and Installation
Before proceeding, ensure that you have the necessary tools installed on your system. We will be creating a local KinD cluster and utilizing Dapr Ambient.

To get started, make sure you have the following CLIs installed:

- [Docker](https://www.docker.com/)

- [KinD (Kubernetes in Docker)](https://kind.sigs.k8s.io/docs/user/quick-start/)

- [kubectl](https://kubernetes.io/docs/tasks/tools/)

- [Helm](https://helm.sh/docs/intro/install/)

The installation of these CLIs is essential for the successful setup of Dapr Ambient with KinD.

## Creating a local Kubernetes cluster with KinD: 

Here, you will create a simple kubernetes cluster with KinD defaults running the following command:

```bash
  kind create cluster --name dapr-ambient
```

## Installing Redis into the KinD cluster:

On this step, you will use helm to install the redis into the kubernetes cluster:

```sh
  helm repo add bitnami https://charts.bitnami.com/bitnami
  helm repo update                            
  helm install redis bitnami/redis --set image.tag=6.2 --set architecture=standalone
```

Finally, let's install Dapr: 

```sh
  helm repo add dapr https://dapr.github.io/helm-charts/
  helm repo update
  helm upgrade --install dapr dapr/dapr \
  --version=1.10.4 \
  --namespace dapr-system \
  --create-namespace \
  --wait
```

Note that you create a new namespace calles `dapr-system`.

## Installing Dapr Components

In this section, we will be install two Dapr Building block: [Publish and Subscriber](https://docs.dapr.io/developing-applications/building-blocks/pubsub/pubsub-overview/) and [State Management](https://docs.dapr.io/developing-applications/building-blocks/state-management/state-management-overview/). All Building Blocks will use [Redis](https://redis.io/) for their purposes.


So before deploying our applications let's configure these components to connect the Redis instance that we created before. 

Create the StateStore component applying this resource to Kubernetes by running:

```sh
kubectl apply -f - <<EOF
apiVersion: dapr.io/v1alpha1
kind: Component
metadata:
  name: statestore
spec:
  type: state.redis
  version: v1
  metadata:
  - name: keyPrefix
    value: name
  - name: redisHost
    value: redis-master:6379
  - name: redisPassword
    secretKeyRef:
      name: redis
      key: redis-password
auth:
  secretStore: kubernetes
EOF
```

Create the PubSub component applying this resource to Kubernetes by running:

```sh
kubectl apply -f - <<EOF
apiVersion: dapr.io/v1alpha1
kind: Component
metadata:
  name: notifications-pubsub
spec:
  type: pubsub.redis
  version: v1
  metadata:
  - name: redisHost
    value: redis-master:6379
  - name: redisPassword
    secretKeyRef:
      name: redis
      key: redis-password
auth:
  secretStore: kubernetes
EOF
```

Once we have the PubSub component configured, we can register Subscritions to define who and where notifications will be sent when new messages arrive to a certain topic.

Create the Subscription component applying this resource to Kubernetes by running:

```sh
kubectl apply -f - <<EOF
apiVersion: dapr.io/v1alpha1
kind: Subscription
metadata:
  name: notifications-subscription
spec:
  topic: notifications 
  route: /notifications
  pubsubname: notifications-pubsub
EOF
```

## Installing Dapr Ambient and all applications

Finally, let's install Dapr Ambient and three applications that uses the Dapr StateStore and PubSub components.

Install Dapr Ambient running this:

```sh
  helm package  chart/dapr-ambient
  helm install my-ambient-dapr-ambient dapr-ambient-1.9.5.tgz --set ambient.appId=read-values --set ambient.proxy.remoteURL=read-values-svc:8080
```

Let's deploy the apps:

This are normal/regular Kubernetes applications, using Deployments and Services.
```sh
  kubectl apply -f https://raw.githubusercontent.com/salaboy/dapr-ambient-examples/main/apps.yaml
```

If you want to see the implementation's detail, you [can access this repository](https://github.com/salaboy/dapr-ambient-examples).

### Saving using the write-values application

Let's create a value on the store:

```sh
  kubectl port-forward svc/write-values-svc 8080:8080
```

Send a request to the application:

```sh
  curl --request POST \
  --url 'http://localhost:8080/?value=10'
``` 

You can see the log using `kubectl logs -f <pod>`

At this point the `subscriber` application has been received the notification from `dapr-ambient`. You can see this, with the same way, using `kubectl logs -f <pod>`.

### Getting the average fom read-values application

The `read-values` applications gets all values from StateStore and calculates the average.

```sh
  kubectl port-forward svc/read-values-svc 8888:8080
```

After, you can make a request to `read-values-svc`:

```sh
  curl http://localhost:8888
```

The response should looks like it:

```
10
```

## Thank you

In this tutorial you gets how to use Dapr Ambient with some Kubernetes applications.