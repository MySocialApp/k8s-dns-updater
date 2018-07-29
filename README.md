# k8s-dns-updater [![Build Status](https://travis-ci.org/MySocialApp/k8s-dns-updater.svg?branch=master)](https://travis-ci.org/MySocialApp/k8s-dns-updater)

Kubernetes DNS updater is a tool to automatically update DNS entries on a round robin configuration when a node goes into maintenance (drain and uncordon).

# Usage

Simply copy the example configuration file config.yaml.example to config.yaml and update the configuration with your needs:

```yaml
GlobalConfig:
  # Use node name (node) or dns name (dns) IP to update DNS
  UpdateDnsType: node
  # Maximum entries in the Round Robin DNS
  MaxDnsEntries: 10

# DNS info
DnsInfos:
  Name: "my-round-robin.domain.com"
  Type: A
  Ttl: 120
  Proxied: false

# Credentials
CloudFlareApiInfos:
  Zoneid: ""
  Zonename: ""
  Email: ""
  Key: ""
```

Then launch the binary in the same folder than the configuration file.

# Requirements

* Kubeconfig must be setup to be able to connect to a Kubernetes cluster
* Support only CloudFlare provider
