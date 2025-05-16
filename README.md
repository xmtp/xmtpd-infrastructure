# xmtpd infrastructure

This repository provides infrastructure-as-code examples and tooling to help node operators deploy and manage xmtpd nodes. xmtpd (XMTP daemon) is the node software that powers the testnet and will power the mainnet of the decentralized XMTP network.

## Minimum system requirements

Each node should be configured for high availability (HA) across all required components, including the database, xmtpd, and the MLS validation service.

Database:
- 2vCPU
- 8GB RAM
- Postgres 16.0 or newer
- 20ms commit latency
- 250MB/s throughput

xmtpd:
- 2vCPU
- 2GiB memory
- 1GB/s network link

MLS validation service:
- 2vCPU
- 512MiB memory

## Get started

1. Provision your infrastructure.

   You can use this [Terraform tooling](/terraform/) if you need to provision underlying cloud infrastructure.
   
   [XMTPD Terraform Modules](/terraform/README.md) describes how to use Terraform modules to provision cloud infrastructure for xmtpd nodes (AWS support).

2. Deploy xmtpd to your infrastructure.
   
   - You can use these [Helm charts](/helm/) to deploy xmtpd into an existing Kubernetes cluster or on managed Kubernetes services.

     [Install xmtpd on Kubernetes using Helm charts](/helm/README.md) describes how to install xmtpd on Kubernetes using Helm charts.

   - Optionally, if you are using Google Kubernetes Engine, you can run xmtpd on GKE with Nginx ingress and Let’s Encrypt.

     [Deploy xmtpd on Google Kubernetes Engine secured by SSL/TLS](/doc/nginx-cert-gke.md) describes how to secure your deployment with HTTPS and ingress.

3. Monitor xmtpd with Prometheus.
   
   Optionally, you can use Kubernetes and Prometheus to set up observability. 

   [Set up Prometheus service discovery for xmtpd in Kubernetes using Helm](/doc/k8s-prometheus-monitoring.md) describes how to automatically scrape metrics from xmtpd pods, visualize in the metrics in Grafana, and set alerts.

## Learn more

- [XMTP documentation](https://docs.xmtp.org)
- [Decentralizing XMTP](https://xmtp.org/decentralizing-xmtp)

## Contribute

Contributions are welcome! See the [contributing guidelines](CONTRIBUTING.md) for details on how to get involved.
