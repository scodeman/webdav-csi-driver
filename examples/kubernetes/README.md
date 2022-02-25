## Kubernetes Volume Provisioning

In Kubernetes volume provisioning, persistent volumes must be pre-provisioned before they are claimed. Static volume provisioning examples includes "pv.yaml" files to pre-provision the volumes.

## WebDaV Configuration

The "pv.yaml" files contain DavFS2 client required information including username, password and url information.

#### WebDAV Client
| Field | Description | Example |
| --- | --- | --- |
| client (or driver) | Client type | "webdav" |
| user | WebDAV user id | "webdav_user" or leave empty for anonymous access |
| password | WebDAV user password | "password" in plane text or leave empty for anonymous access |
| url | URL | "https://webdavserver.io" |

Mounts **url**

### Kubernetes Secrets

Optionally, Kubernetes Secrets can be used to pass sensitive informations such as username and password.
DavFSs host information also can be passed in this way.
Kubernetes Secrets can be supplied via **nodeStageSecretRef**.

### Execute examples in following order

Define Storage Class (SC):
```shell script
kubectl apply -f "storageclass.yaml"
```

Define Persistent Volume (PV):
```shell script
kubectl apply -f "pv.yaml"
```

Claim Persistent Volume (PVC):
```shell script
kubectl apply -f "pvc.yaml"
```

Execute Application with Volume Mount:
```shell script
kubectl apply -f "app.yaml"
```

Undeployment must be done in reverse order.
```shell script
kubectl delete -f "<YAML file>"
```
