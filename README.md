# XMTP Infrastructure

- [XMTP Infrastructure](#xmtp-infrastructure)
  - [Minimum system requirements](#minimum-system-requirements)
  - [Get started](#get-started)
  - [Deploy xmtpd to AWS/ECS infrastructure with Terraform](#deploy-xmtpd-to-awsecs-infrastructure-with-terraform)
  - [Deploy xmtpd to your infrastructure using Helm charts](#deploy-xmtpd-to-your-infrastructure-using-helm-charts)
  - [Monitor xmtpd with Prometheus](#monitor-xmtpd-with-prometheus)
  - [Prune expired messages](#prune-expired-messages)
  - [Networking notes](#networking-notes)
  - [Learn more](#learn-more)
  - [Contribute](#contribute)

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

Choose your infrastructure approach:

- Use [Terraform](#deploy-xmtpd-to-awsecs-infrastructure-with-terraform) if you need to provision the underlying cloud infrastructure.

- Use [Helm charts](#deploy-xmtpd-to-your-infrastructure-using-helm-charts) if you have an existing Kubernetes cluster or want to deploy on managed Kubernetes services.

## Deploy xmtpd to AWS/ECS infrastructure with Terraform

You can use this [Terraform tooling](/terraform/) if you need to provision underlying cloud infrastructure on AWS/ECS.

[XMTPD Terraform Modules](/terraform/README.md) describes how to use Terraform modules to provision AWS/ECS infrastructure for xmtpd nodes.

## Deploy xmtpd to your infrastructure using Helm charts

You can use these [Helm charts](/helm/) to deploy xmtpd into an existing Kubernetes cluster or on managed Kubernetes services.

[Install xmtpd on Kubernetes using Helm charts](/helm/README.md) describes how to install xmtpd on Kubernetes using Helm charts.

Optionally, if you are using Google Kubernetes Engine, you can run xmtpd on GKE with Nginx ingress and Let's Encrypt.

[Deploy xmtpd on Google Kubernetes Engine secured by SSL/TLS](/doc/nginx-cert-gke.md) describes how to secure your deployment with HTTPS and ingress.

## Monitor xmtpd with Prometheus

Optionally, you can use Kubernetes and Prometheus to set up observability.

[Set up Prometheus service discovery for xmtpd in Kubernetes using Helm](/doc/k8s-prometheus-monitoring.md) describes how to automatically scrape metrics from xmtpd pods, visualize in the metrics in Grafana, and set alerts.

## Prune expired messages

To prevent data bloat and keep your node performant, be sure to [prune expired messages from your xmtpd database](/doc/db-pruning.md).

## Networking notes

Refer to the [networking](https://github.com/xmtp/xmtpd/blob/main/doc/networking.md) document to learn more about the networking architecture xmtpd uses.

Currently, xmtpd APIs are implemented using the [Connect-RPC](https://connectrpc.com/) library, which allows gRPC and gRPC-Web clients out of the box.

Because this library uses HTTP2 instead of gRPC, it also relies on requests headers to properly function, including CORS headers.

When using a load balancer in front of xmtpd, make sure it forwards correctly all headers, including CORS.

## Learn more

- [XMTP documentation](https://docs.xmtp.org)
- [Decentralizing XMTP](https://xmtp.org/decentralizing-xmtp)
- [Networking](https://github.com/xmtp/xmtpd/blob/main/doc/networking.md)

## Contribute

Contributions are welcome! See the [contributing guidelines](CONTRIBUTING.md) for details on how to get involved.
