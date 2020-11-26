# :japanese_castle: k8s-resource-scheduler - Simple Kubernetes Resource scheduler

Simple (experimental) CPU/Memory pod scheduler based from https://github.com/kelseyhightower/scheduler.

The scheduler reads kubernetes metrics to determine nodes current CPU and Memory usage, and it tries to assign pods to nodes which are less busy.

## Prerequisite

- Kubernetes cluster
- Metric server running

## Run the Scheduler on Kubernetes


```bash

$ kubectl apply -f https://raw.githubusercontent.com/mudler/k8s-resource-scheduler/master/deployments/scheduler.yaml

```

## Usage

Add `schedulerName` to your pods definition;

```yaml
...
    spec:
      schedulerName: k8s-resource-scheduler
```


### CPU and Memory Bound workloads

There are occasions where you want to weight scheduling based on cpu or memory, or both.


### Privileging CPU bound applications

If applications that you are going to deploy are likely CPU intensive, you might want to privilege scheduling based on cpu load.

You can put annotation in a pod, or in a node:

```yaml
k8s-resource-scheduler/cpu-bound=true
```

### Privileging Memory bound applications

If applications that you are going to deploy are likely Memory intensive, you might want to privilege scheduling based on memory load.

You can put annotation in a pod, or in a node:

```yaml
k8s-resource-scheduler/memory-bound=true
```

By default, the scheduler will try to assign the pod to the node which has less cpu/memory current usage.

