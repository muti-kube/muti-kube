FROM --platform=${TARGETPLATFORM} golang:alpine AS builder
ARG TARGETPLATFORM
RUN echo "building for ${TARGETPLATFORM}"
LABEL stage=gobuilder
ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOPROXY https://goproxy.cn,direct
ARG WORKDIR="/go/src/github.com/muti-kube/muti-kube"
WORKDIR ${WORKDIR}
COPY . ${WORKDIR}/
RUN go mod download
RUN go build cmd/kube.go
RUN mv -f kube /
RUN mv -f config/ /config/
WORKDIR /

FROM ubuntu
ENV TZ=Asia/Shanghai
RUN DEBIAN_FRONTEND=noninteractive apt-get update \
    && apt-get install -y tzdata \
    && apt-get clean \
    && apt-get autoclean \
    && rm -rf /var/lib/apt/lists/* \
EXPOSE 9000
RUN mkdir /root/.kube/
RUN mkdir /opt/config/
COPY --from=builder /kube /opt/muti-kube
COPY --from=builder /config/kube/config  /root/.kube/config
COPY --from=builder /config/config.yml  /opt/config/config.yml
ENTRYPOINT ["/opt/muti-kube","server","-c","/opt/config/config.yml"]