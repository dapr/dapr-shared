## Dapr Ambient :: Step by step Tutorial

In this short tutorial we will install Dapr Ambient and configure a set of applications to work with it.

We will be using Knative Serving and Knative Functions to demonstrate where Dapr Ambient can add value for serverless scenarios. 

## Prerequisites and installation 

We will be creating a local KinD cluster where we will install Knative Serving and Dapr.

For this you will need to install the following CLIs:

- [Install `kubectl`](https://kubernetes.io/docs/tasks/tools/)
- [Install `kind`](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
- [Install `helm`](https://helm.sh/docs/intro/install/) 
- [Install `docker`](https://docs.docker.com/engine/install/)
- [Install the Knative Functions `func` CLI](https://knative.dev/docs/functions/install-func/)
- [Install the `dapr` CLI](https://docs.dapr.io/getting-started/install-dapr-cli/)

Create a local Kubernetes cluster with: 

```
cat <<EOF | kind create cluster --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 31080 # expose port 31380 of the node to port 80 on the host, later to be use by kourier or contour ingress
    listenAddress: 127.0.0.1
    hostPort: 80
EOF
```

Let's now install Knative Serving into the cluster: 

[Check this link for full instructions from the official docs](https://knative.dev/docs/install/yaml-install/serving/install-serving-with-yaml/#prerequisites)

```
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.8.0/serving-crds.yaml
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.8.0/serving-core.yaml

```

Installing the networking stack to support advanced traffic management: 

```
kubectl apply -f https://github.com/knative/net-kourier/releases/download/knative-v1.8.0/kourier.yaml

```

```
kubectl patch configmap/config-network \
  --namespace knative-serving \
  --type merge \
  --patch '{"data":{"ingress-class":"kourier.ingress.networking.knative.dev"}}'

```

Configuring domain mappings: 

```
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.8.0/serving-default-domain.yaml

```

**Only for Knative on KinD** 

For Knative Magic DNS to work in KinD you need to patch the following ConfigMap:

```
kubectl patch configmap -n knative-serving config-domain -p "{\"data\": {\"127.0.0.1.sslip.io\": \"\"}}"
```

and if you installed the `kourier` networking layer you need to create an ingress:

```
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  name: kourier-ingress
  namespace: kourier-system
  labels:
    networking.knative.dev/ingress-provider: kourier
spec:
  type: NodePort
  selector:
    app: 3scale-kourier-gateway
  ports:
    - name: http2
      nodePort: 31080
      port: 80
      targetPort: 8080
EOF
```


Let's create a Redis instance that our applications can use to store state or exchange messages: 

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