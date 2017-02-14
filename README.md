# kloader
Runs commands when a Kubernetes Configmap changes.

### ConfigMap
Many applications require configuration via some combination of config files. These configuration artifacts
should be decoupled from image content in order to keep containerized applications portable.
The ConfigMap API resource provides mechanisms to inject containers with configuration data while keeping
containers agnostic of Kubernetes. ConfigMap can be used to store fine-grained information like individual
properties or coarse-grained information like entire config files or JSON blobs.

[Read More about the use cases and usage of ConfigMap](https://kubernetes.io/docs/user-guide/configmap/).

## What Kloader does?
`Kloader` watches a specified ConfigMap, mount the ConfigMap data in specified directory as files. In case of
any update in ConfigMap data `Kloader` updates the mounted file and run an additional bash script.

## Configurations
`Kloader` has following configurations that should be provided as `flags`.

```
--config-map
ConfigMap name that will be mounted.


--mount-location
Location the ConfigMap Will be mounted inside


--boot-cmd
bash script or script file location that will be run after mounted file update.


--k8s-master
Kubernetes API server address. Default is InCluster address.


--k8s-config
Kubernetes API Configurations. Default is InCluster config.
```

## Building Kloader
```
./hack/make.py build kloader
```
