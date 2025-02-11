# xmtpd infrastructure

This repository provides infrastructure-as-code examples and tooling to help node operators deploy and manage xmtpd nodes. xmtpd (XMTP daemon) is the node software that powers the testnet and will power the mainnet of the decentralized XMTP network.

## Minimum system requirements

Each node should be configured for high availability (HA) across all required components, including the database, xmtpd, and the MLS validation service.

Database:
- Postgres 16.0 or newer
- 20ms commit latency
- 250MB/s throughput
- 8GB RAM
- XXXXX CPUs

xmtpd:
- 2vCPU
- 2GiB memory
- 1GB/s network link

MLS validation service:
- 2vCPU
- 512MiB memory

## Available tooling

- [Helm charts](./helm/README.md) - Deploy xmtpd nodes on Kubernetes clusters
- [Terraform](./terraform/README.md) - Provision cloud infrastructure for xmtpd nodes

## Production deployment guide for GKE

See [Deploy xmtpd on Google Kubernetes Engine secured by SSL/TLS](./doc/nginx-cert-gke.md) for a detailed guide to creating a production-ready deployment of xmtpd on GKE using NGINX Ingress Controller and Let's Encrypt certificates.

## Get started

1. Choose your infrastructure approach:
   - Use Helm charts if you have an existing Kubernetes cluster or want to deploy on managed Kubernetes services
   - Use Terraform if you need to provision the underlying cloud infrastructure
2. Follow the respective README for your chosen tool
3. For GKE production deployments, refer to the SSL/TLS deployment guide for securing your node

## Learn more

- [Decentralizing XMTP](https://xmtp.org/decentralizing-xmtp)
- [XMTP documentation](https://docs.xmtp.org)

## Contribute

Contributions are welcome! See the [contributing guidelines](CONTRIBUTING.md) for details on how to get involved.
