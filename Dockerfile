ARG SRC_DIR="/go/src/github.com/scodeman/webdav-csi-driver/"
FROM golang:1.16.8-stretch AS webdav_csi_driver_builder
WORKDIR ${SRC_DIR}
ENV GOPROXY=direct
COPY go.mod .
COPY go.sum .
RUN go mod download
ADD . .
RUN make webdav-csi-driver

FROM ubuntu:20.04
ARG GOROOT=/usr/local/go
ARG SRC_DIR="/go/src/github.com/scodeman/webdav-csi-driver"
ARG DEBIAN_FRONTEND=noninteractive
RUN apt-get update && \
    apt-get install -y \
      davfs2=1.5.5-1 \
      dumb-init=1.2.2-1.2
COPY --from=webdav_csi_driver_builder \
  ${SRC_DIR}/bin/webdav-csi-driver \
  /usr/bin/webdav-csi-driver
ENTRYPOINT ["/usr/bin/dumb-init", "--", "/usr/bin/webdav-csi-driver"]
