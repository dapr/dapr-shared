# Dapr Shared with KinD
This tutorial provides step-by-step instructions for installing Dapr Shared and configuring a set of applications to work with it.

## Prerequisites and Installation

Before proceeding, please make sure that you have the necessary tools installed on your system. We will create a local KinD cluster to install Dapr, some applications, and an instance of Dapr Shared.

To get started, make sure you have the following CLIs installed:

- [Docker](https://www.docker.com/)

- [KinD (Kubernetes in Docker)](https://kind.sigs.k8s.io/docs/user/quick-start/)

- [kubectl](https://kubernetes.io/docs/tasks/tools/)

- [Helm](https://helm.sh/docs/intro/install/)



## Creating a local Kubernetes cluster with KinD: 

Here, you will create a simple Kubernetes cluster with KinD defaults running the following command:

```bash
  kind create cluster --name dapr-shared
```

## Installing Redis into the KinD cluster:

Let's create a new Redis Instance for our application's services to use: 

```sh
  helm repo add bitnami https://charts.bitnami.com/bitnami
  helm repo update                            
  helm install redis bitnami/redis --set image.tag=6.2 --set architecture=standalone
```

Finally, let's install the Dapr Control Plane: 

```sh
  helm repo add dapr https://dapr.github.io/helm-charts/
  helm repo update
  helm upgrade --install dapr dapr/dapr \
  --version=1.12.3 \
  --namespace dapr-system \
  --create-namespace \
  --wait
```

Note that you create a new namespace called `dapr-system`.

## Installing Dapr Components

In this section, we will be configure two Dapr Components: [PubSub](https://docs.dapr.io/developing-applications/building-blocks/pubsub/pubsub-overview/) and [StateStore](https://docs.dapr.io/developing-applications/building-blocks/state-management/state-management-overview/). 
Both of these components will use Redis as their implementation. 

So before deploying our applications, let's configure these components to connect the Redis instance we created. 

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

Once configured the PubSub component, we can register Subscriptions to define who and where notifications will be sent when new messages arrive to the `notification` topic.

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

## Installing Dapr Shared and all applications

Finally, let's install Dapr Shared and three applications that uses the Dapr StateStore and PubSub components.

Install Dapr Shared Helm Chart running this:

```sh
  helm install my-dapr-shared oci://docker.io/daprio/dapr-shared-chart --set shared.appId=my-dapr-app --set shared.remoteURL=subscriber-svc --set shared.remotePort=80
```

Now that we have the Dapr control plane, Redis, the PubSub and StateStore component, and our Dapr Shared instance, let's deploy the example apps:

These are standard Kubernetes applications, using `Deployments` and `Services`.
```sh
  kubectl apply -f https://raw.githubusercontent.com/salaboy/dapr-shared-examples/main/apps.yaml
```

If you want to see the implementation's detail, you [can access this repository](https://github.com/salaboy/dapr-shared-examples).

### Storing data using the write-values application

Let's submit a value to the `write-values` service, but first, let's use `kubectl port-forward` to be able to reach the service which is running inside our cluster:

```sh
  kubectl port-forward svc/write-values-svc 8080:80
```

Now you can send an HTTP request to the application:

```sh
  curl --request POST \
  --url 'http://localhost:8080/?value=10'
``` 

You can see the log using `kubectl logs -f <pod>`

At this point the `subscriber` application has received the notification from `dapr-shared`. You can see this, with the same way, using `kubectl logs -f <pod>`.

### Getting the message on the subscriber application

When the application `write-values` save a value on Redis, after it is published, an event to topic `notifications`.

You can see the logs following these steps:

Execute the following command:
```sh
  kubectl get pods
```

Select the subscriber pod, and execute it:

```sh
  kubectl logs -f <subscribe-pod-name-here>
```

The logs should look like it:

```
2023/05/19 14:55:57 Starting Subscriber in Port: 8080
2023/05/19 14:57:02 POST /notifications HTTP/1.1
Host: subscriber-svc:8080
Accept-Encoding: gzip
Content-Length: 406
Content-Type: application/cloudevents+json
Pubsubname: notifications-pubsub
Traceparent: 00-00000000000000000000000000000000-0000000000000000-00
User-Agent: fasthttp

{"data":"10","datacontenttype":"text/plain","id":"7447314d-89a8-4144-a9b8-6be357aee618","pubsubname":"notifications-pubsub","source":"my-dapr-app","specversion":"1.0","time":"2023-05-19T14:57:02Z","topic":"notifications","traceid":"00-00000000000000000000000000000000-0000000000000000-00","traceparent":"00-00000000000000000000000000000000-0000000000000000-00","tracestate":"","type":"com.dapr.event.sent"}
Subscriber received on /notifications: 10
```

### Getting the average fom read-values application

The `read-values` application gets all values from StateStore and calculates the average.

```sh
  kubectl port-forward svc/read-values-svc 8888:80
```

After, you can make a request to `read-values-svc`:

```sh
  curl http://localhost:8888
```

The response should look like it:

```
10
```

## Get involved

If you want to contribute to Dapr Shared please get in touch, create an issue, or submit a Pull Request. 
You can also check the Project Roadmap to see what is coming or to find out how you can help us to get the next version done. 
