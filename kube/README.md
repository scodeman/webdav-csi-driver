## WebDAV CSI Driver Installation on Kubernetes
This directory contains YAML files to install WebDAV CSI Driver on Kubernetes.

## Install the driver
Install the stable driver:

```shell script
kubectl apply -k "overlays/stable"
```

Install the development driver:
```shell script
kubectl apply -k "overlays/dev"
```

Verify the driver installation:
```shell script
kubectl get csinodes -o jsonpath='{range .items[*]} {.metadata.name}{": "} {range .spec.drivers[*]} {.name}{"\n"} {end}{end}'
```

## Uninstall the driver
Uninstall the stable driver:
```shell script
kubectl delete -k "overlays/stable"
```

Uninstall the development driver:
```shell script
kubectl delete -k "overlays/dev"
```
