FROM quay.io/mocaccino/extra AS builder
RUN luet install -y lang/go
WORKDIR /scheduler
COPY . .
RUN ./build

FROM quay.io/mocaccino/extra
MAINTAINER Ettore Di Giacinto <mudler@mocaccino.org>
COPY --from=builder /scheduler/k8s-resource-scheduler /usr/bin/scheduler
RUN luet install -y container/kubectl
ENTRYPOINT ["/usr/bin/scheduler"]
