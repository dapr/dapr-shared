# Dapr Ambient 

Dapr Ambient allows you to create Dapr Applications using the `daprd` Sidecar as a Kubernetes `DeamonSet`. This enables other use cases where Sidecars are not the best option. 

By running `daprd` as a Kubernetes `DaemonSet` the `daprd` container will be present in each Kubernetes Node, reducing the network hops between the applications and Dapr. 


If you need multiple Dapr Applications you can deploy this chart multiple times using different `ambient.appId`s. 


To deploy this chart you can run from inside the `chart/dapr-ambient` directory: 

```
helm install my-ambient . --set ambient.appId=<DAPR_APP_ID> --set ambient.proxy.remoteURL=<REMOTE_URL>  

```

Where `<DAPR_APP_ID>` is the Dapr App Id that you can use in your components (for example for scopes) and `<REMOTE_URL>` is a reachable URL where `dapr-ambient` will forward notifications received by the Dapr sidecar. 



Future versions might include forwarding notifications to multiple remote URLs.

## Building from source

I've used the [CNCF `ko` project](https://ko.build/) to build multiplatform images for the proxy. 
You can run the following command to build containers for the `dapr-ambient` proxy: 

```
ko build --platform=linux/amd64,linux/arm64
```

