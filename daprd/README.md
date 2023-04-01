# Custom `daprd` version

This Dockerfile produces a custom `daprd` version that also contains `sh` coming from `alpine` to be able to export environment variables before launching the `darpd` binary. 

Check the `daprd` version that you want to use inside the [`Dockerfile`](Dockerfile). 

To build this container run (from within this directory): 

```
docker build -t <user>/daprd:<version> .
```

Then to push: 

```
docker push <user>/daprd:<version>
```

When running `dapr-ambient` you can specify to use this container instead of `daprd` by setting up this value: 

```
--set dapr.ambient.daprd.image.user=<user> --set dapr.ambient.daprd.image.name=daprd --set dapr.ambient.daprd.image.version=<version>
```
