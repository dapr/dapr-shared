# Dapr Ambient with KinD
This tutorial provides step-by-step instructions for installing Dapr Ambient and configuring a set of applications to work with it.

## Prerequisites and Installation
Before proceeding, ensure that you have the necessary tools installed on your system. We will be creating a local KinD cluster and utilizing dapr-ambiente.

To get started, make sure you have the following CLIs installed:

- Docker: The Docker software is required and can be downloaded and installed from the official website (https://www.docker.com/).

- KinD (Kubernetes in Docker): KinD is a tool that enables running Kubernetes clusters using Docker. Follow the instructions in the KinD documentation (https://kind.sigs.k8s.io/docs/user/quick-start/) to install it on your machine.

- kubectl: kubectl is the command-line tool used to interact with Kubernetes clusters. It is required for managing the deployment and configuration of applications. You can install kubectl by following the instructions provided in the Kubernetes documentation (https://kubernetes.io/docs/tasks/tools/).

- Helm: Helm is a package manager for Kubernetes that simplifies the deployment and management of applications. Install Helm by following the instructions in the Helm documentation (https://helm.sh/docs/intro/install/).

The installation of these CLIs is essential for the successful setup of Dapr Ambient with KinD.

## Create a local Kubernetes cluster with: 

```bash
  kind create cluster --name dapr-ambient
```

Installing Redis:

```
helm repo add bitnami https://charts.bitnami.com/bitnami
helm repo update                            
helm install redis bitnami/redis --set image.tag=6.2 --set architecture=standalone
```

Finally, let's install Dapr into the Cluster: 

```
helm repo add dapr https://dapr.github.io/helm-charts/
helm repo update
helm upgrade --install dapr dapr/dapr \
--version=1.10.4 \
--namespace dapr-system \
--create-namespace \
--wait
```

Let's now deploy and configure some Dapr apps! 


## Deploying the applications and wiring things together

In this section, we will be deploying three applications that want to store and read data from a state store and publish and consume messages. 
To achieve this we will use the Dapr StateStore and PubSub components. So before deploying our applications let's configure these components to connect the Redis instance that we created before. 

The Dapr Statestore configuration looks like this: 
```
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
```

We can apply this resource to Kubernetes by running: 
```
kubectl apply -f resources/statestore.yaml
```

The PubSub Component looks like this: 
```
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
```

We can apply this resource to Kubernetes by running: 
```
kubectl apply -f resources/pubsub.yaml
```

Once we have the PubSub component configured, we can register Subscritions to define who and where notifications will be sent when new messages arrive to a certain topic. A Subscription resource look like this: 

```
apiVersion: dapr.io/v1alpha1
kind: Subscription
metadata:
  name: notifications-subscritpion
spec:
  topic: notifications 
  route: /notifications
  pubsubname: notifications-pubsub
```

Finally, let's deploy three applications that uses the Dapr StateStore and PubSub components. This are normal/regular Kubernetes applications, using Deployments and Services. To make these apps dapr-aware we just need to add some Dapr annotations:


```

```

Let's deploy the apps with: 
```
kubectl apply -f apps.yaml
```