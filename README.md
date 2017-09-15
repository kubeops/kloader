[![Go Report Card](https://goreportcard.com/badge/github.com/appscode/kloader)](https://goreportcard.com/report/github.com/appscode/kloader)

[Website](https://appscode.com) • [Slack](https://slack.appscode.com) • [Forum](https://discuss.appscode.com) • [Twitter](https://twitter.com/AppsCodeHQ)

# kloader
Runs commands when a Kubernetes Configmap/Secret changes.

## Why Kloader?
Many applications require configuration via some combination of config files. These configuration artifacts
should be decoupled from image content in order to keep containerized applications portable.
The ConfigMap API resource provides mechanisms to inject containers with configuration data while keeping
containers agnostic of Kubernetes. ConfigMap can be used to store fine-grained information like individual
properties or coarse-grained information like entire config files or JSON blobs. [Read More about the use cases and usage of ConfigMap](https://kubernetes.io/docs/user-guide/configmap/).

`Kloader` watches a specified ConfigMap, mount the ConfigMap data in specified directory as files. In case of
any update in ConfigMap data `Kloader` updates the mounted file and run an additional bash script.

## Configuration
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

## Release Kloader
```sh
./hack/make.py build; env APPSCODE_ENV=prod ./hack/make.py push; ./hack/make.py push
```

## Versioning Policy
Kloader __does not follow semver__, rather the _major_ version of operator points to the
Kubernetes [client-go](https://github.com/kubernetes/client-go#branches-and-tags) version. You can verify this
from the `glide.yaml` file. This means there might be breaking changes between point releases of the kloader. Please always check the release notes for upgrade instructions.

---

**Kloader collects anonymous usage statistics to help us learn how the software is being used and how we can improve it.
To disable stats collection, run the binary with the flag** `--analytics=false`.

---
