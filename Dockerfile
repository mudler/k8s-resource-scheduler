FROM quay.io/mocaccino/extra AS builder
RUN luet install -y lang/go container/kubectl
WORKDIR /scheduler
COPY . .
RUN ./build

FROM scratch
MAINTAINER Ettore Di Giacinto <mudler@mocaccino.org>
COPY --from=builder /scheduler/k8s-resource-scheduler /usr/bin/scheduler
COPY --from=builder /usr/bin/kubectl /usr/bin/kubectl
ENTRYPOINT ["/usr/bin/scheduler"]
