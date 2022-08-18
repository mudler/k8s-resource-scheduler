FROM golang:1.18 AS builder
RUN curl -fsSLo /usr/share/keyrings/kubernetes-archive-keyring.gpg https://packages.cloud.google.com/apt/doc/apt-key.gpg
RUN echo "deb [signed-by=/usr/share/keyrings/kubernetes-archive-keyring.gpg] https://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee /etc/apt/sources.list.d/kubernetes.list
RUN apt-get update
RUN apt-get install -y kubectl
WORKDIR /scheduler
COPY . .
RUN ./build

FROM scratch
MAINTAINER Ettore Di Giacinto <mudler@mocaccino.org>
COPY --from=builder /scheduler/k8s-resource-scheduler /usr/bin/scheduler
COPY --from=builder /kubernetes/client/bin/kubectl /usr/bin/kubectl
ENTRYPOINT ["/usr/bin/scheduler"]
