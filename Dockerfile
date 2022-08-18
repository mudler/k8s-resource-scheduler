FROM quay.io/mocaccino/extra AS builder
RUN luet install -y lang/go
ARG KUBECTL_VERSION=v1.18.2
ARG KUBECTL_ARCH=linux-$(uname -m)
RUN wget -O kubectl.tar.gz https://dl.k8s.io/$KUBECTL_VERSION/kubernetes-client-$KUBECTL_ARCH.tar.gz && \
    echo "$KUBECTL_CHECKSUM kubectl.tar.gz"  && \
    tar xvf kubectl.tar.gz -C /
WORKDIR /scheduler
COPY . .
RUN ./build

FROM scratch
MAINTAINER Ettore Di Giacinto <mudler@mocaccino.org>
COPY --from=builder /scheduler/k8s-resource-scheduler /usr/bin/scheduler
COPY --from=builder /kubernetes/client/bin/kubectl /usr/bin/kubectl
ENTRYPOINT ["/usr/bin/scheduler"]
