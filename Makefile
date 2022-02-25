PKG=github.com/scodeman/webdav-csi-driver
CSI_DRIVER_IMAGE?=scodeman/webdav-csi-driver
CSI_DRIVER_DOCKERFILE=Dockerfile
VERSION=v0.0.1
GIT_COMMIT?=$(shell git rev-parse HEAD)
BUILD_DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS?="-X ${PKG}/pkg/driver.driverVersion=${VERSION} -X ${PKG}/pkg/driver.gitCommit=${GIT_COMMIT} -X ${PKG}/pkg/driver.buildDate=${BUILD_DATE}"
GO111MODULE=on
GOPROXY=direct
GOPATH=$(shell go env GOPATH)

.EXPORT_ALL_VARIABLES:

.PHONY: webdav-csi-driver
webdav-csi-driver:
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux go build -ldflags ${LDFLAGS} -o bin/webdav-csi-driver ./cmd/

.PHONY: image
image:
	docker build -t $(CSI_DRIVER_IMAGE):latest -f $(CSI_DRIVER_DOCKERFILE) .

.PHONY: image-clean
image-clean:
	docker rmi -f scodeman/webdav-csi-driver:latest scodeman/webdav-csi-driver:$(VERSION) webdav_csi_driver_build   -f

.PHONY: push
push: image
	docker push $(CSI_DRIVER_IMAGE):latest

.PHONY: image-release
image-release:
	docker build -t $(CSI_DRIVER_IMAGE):$(VERSION) -f $(CSI_DRIVER_DOCKERFILE) .

.PHONY: push-release
push-release:
	docker push $(CSI_DRIVER_IMAGE):$(VERSION)

.PHONY: helm
helm:
	helm lint helm && helm package helm
