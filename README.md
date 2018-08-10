# k8s-dns-updater [![Build Status](https://travis-ci.org/MySocialApp/k8s-dns-updater.svg?branch=master)](https://travis-ci.org/MySocialApp/k8s-dns-updater) [![Docker Repository on Quay](https://quay.io/repository/mysocialapp/k8s-dns-updater/status "Docker Repository on Quay")](https://quay.io/repository/mysocialapp/k8s-dns-updater)

Kubernetes DNS updater is a tool to automatically update DNS entries on a round robin (RR) configuration when a node goes into maintenance.

![test](img/kdu_main.png)

We've made this application at [MySocialApp](https://mysocialapp.io) in order to have automatic changes to:

* Add a node in the round robin DNS when a node is uncordon
* Remove a node from the round robin DNS when a node is drain

# Usage

Simply copy the binary and the example configuration file [config.yaml.example](config.yaml.example) to config.yaml. Then update the configuration with your information:

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

# Kubernetes (HELM)

You can deploy it with the provided HELM chart. First update the [values.yaml](kubernetes/values.yaml) file:

```yaml
KduImageVersion: v0.2
#KduNodeSelector:
#  node-role.kubernetes.io/node: "true"
KduRbacEnabled: true

# Global Config
KduGlobalUpdateType: node
KduGlobalMaxDnsEntries: 10

# DNS Info
KduInfosName: "your_rr_record"
KduInfosType: A
KduInfosTtl: 120
KduInfosProxied: false

# Cloudflare API
KduCfZoneId: "your_id"
KduCfZoneName: "your-domain.com"
KduCfEmail: "your_mail"
KduCfKey: "your_key"
```

Then deploy it into your cluster:

```bash
helm install --values kubernetes/values.yaml kubernetes
```

# Limitations

* Support only CloudFlare provider

# Todo

* When starting, validate the current status and update accordingly the DNS
* Add Ingress support and detect ingress readiness before adding back in RR
* Add Ingress support and detect if an ingress readiness is failing to remove from RR
* Support a limit of the number of DNS entries in RR
* Add prometheus metrics
