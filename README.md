## WebDAV CSI Driver

This Container Storage Interface (CSI) Driver implements the [CSI Specification](https://github.com/container-storage-interface/spec/blob/master/spec.md) to provide WebDAV access within [Kubernetes](https://kubernetes.io/)) .

### CSI Specification Compatibility

WebDAV CSI Driver only supports CSI Specification Version v1.2.0 or higher.

### Volume Mount Parameters

Parameters specified in Persistent Volume (PV) and Storage Class (SC) are passed to the CSI Driver to mount a volume.

#### WebDAV Driver
| Field | Description | Example |
| --- | --- | --- |
| client | Driver type | "davfs2" |
| user | WebDAV user id | "webdav_user" |
| password | WebDAV user password | "password" in plain text |
| url | URL | "http://webdavserver.io" |

Mounts **url**

**user** and **password** can be supplied via secrets (nodeStageSecretRef).
Please check out `examples` for more information.

### Install & Uninstall

Be aware that the Master branch is not stable! Please use recently released version of code.

Installation can be done using [Helm Chart Repository](https://scodeman.github.io/webdav-csi-driver-helm/), [Helm Chart (manual)](https://github.com/scodeman/webdav-csi-driver/tree/master/helm) or by [Manual Deployment](https://github.com/scodeman/webdav-csi-driver/tree/master/kube).

Install using Helm Chart Repository with default configuration:
```shell script
helm repo add webdav-csi-driver-repo https://scodeman.github.io/webdav-csi-driver-helm/
helm install webdav-csi-driver webdav-csi-driver-repo/webdav-csi-driver
```

Install using Helm Chart with default configuration:
```shell script
helm install webdav-csi-driver helm
```

Install using Helm Chart with custom configuration:
Edit `helm/user_values.yaml` file. You can set global configuration using the file.

```shell script
helm install webdav-csi-driver -f helm/user_values.yaml helm
```

Uninstall using Helm Chart:
```shell script
helm delete webdav-csi-driver
```

### Example of Pre-previsioned Persistent Volume
Please check out [examples](https://github.com/scodeman/webdav-csi-driver/tree/master/examples).

Define Storage Class (SC):
```shell script
kubectl apply -f "examples/kubernetes/storageclass.yaml"
```

Define Persistent Volume (PV):
```shell script
kubectl apply -f "examples/kubernetes/pv.yaml"
```

Claim Persistent Volume (PVC):
```shell script
kubectl apply -f "examples/kubernetes/pvc.yaml"
```

Execute Application with Volume Mount:
```shell script
kubectl apply -f "examples/kubernetes/app.yaml"
```

To undeploy, use following command:
```shell script
kubectl delete -f "<YAML file>"
```

### References

Following CSI driver implementations were used as references:
- [AWS EFS CSI Driver](https://github.com/kubernetes-sigs/aws-efs-csi-driver)
- [AWS FSx CSI Driver](https://github.com/kubernetes-sigs/aws-fsx-csi-driver)
- [iRODS CSI Driver](https://github.com/scodeman/webdav-csi-driver)

Many code parts in the driver are from **AWS EFS CSI Driver** and **AWS FSx CSI Driver**.

Following resources are helpful to understand the CSI driver implementation:
- [CSI Specification](https://github.com/container-storage-interface/spec/blob/master/spec.md)
- [Kubernetes CSI Developer Documentation](https://kubernetes-csi.github.io/docs/)

Following resources are helpful to configure the CSI driver:
- [Kubernetes Secrets](https://kubernetes.io/docs/concepts/configuration/secret/)

####  Licenses

The driver contains open-source original code under CyVerse license
Please check [LICENSE](https://github.com/cyverse/irods-csi-driver/tree/master/LICENSE.CyVerse) file.

The driver contains open-source code parts under Apache License v2.0.
The code files containing the open-source code parts have the Apache license header in front and which parts are from which code.
Please check [LICENSE](https://github.com/scodeman/webdav-csi-driver/tree/master/LICENSE.APL) file for the details of Apache License v2.0.
