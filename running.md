# Installing and running the tool(s)

We assume some level of familiarity with the following tools and technologies:

* Kafka
* uReplicator
* Brooklin
* AWS
* Kubernetes

And some nice to have and useful skills:

* Prometheus
* Grafana
* Golang

## Prerequisite and setup

The following tools are required:
* `make` (already installed on most systems)
* `AWS CLI`. And setup AWS keys
* `kops`. The current tested version is 1.10.0 (with brew it's `brew install kops@1.10.0` or `brew upgrade kops@1.10.0` or `brew switch kops 1.10.0`)
* `kubectl` - the kubernetes CLI
* `kafka client tools`, in particular: `zookeeper-shell` and `kafka-console-consumer`

# Running it

NOTICE: This will incur costs from AWS. We setup up hefty clusters and drive traffic between them and this is costs $$$

```
make k8s-all # Wait for all resources to be created. This could take up to 40min, depending on the cluster size.
```

# Destroying it

```
make k8s-delete-all # And wait for all resources to get deleted. This can take a few minutes
```
