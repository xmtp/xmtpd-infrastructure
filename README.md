# XMTP Infrastructure

- [XMTP Infrastructure](#xmtp-infrastructure)
  - [Minimum system requirements](#minimum-system-requirements)
  - [Get started](#get-started)
  - [Deploy xmtpd to AWS/ECS infrastructure with Terraform](#deploy-xmtpd-to-awsecs-infrastructure-with-terraform)
  - [Deploy xmtpd to your infrastructure using Helm charts](#deploy-xmtpd-to-your-infrastructure-using-helm-charts)
  - [Monitor xmtpd with Prometheus](#monitor-xmtpd-with-prometheus)
  - [Prune expired messages](#prune-expired-messages)
  - [Security protocols](#security-protocols)
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

## Security protocols

Node operators must implement robust security protocols across all layers of their deployment. The following practices are recommended for production environments:

- **Transport Layer Security (TLS)**: Enforce TLS for all external communications with automated certificate management. See [TLS configuration for Kubernetes](/doc/nginx-cert-gke.md) or [AWS load balancer TLS configuration](/terraform/aws/xmtpd-api/load-balancer.tf).

- **Secrets management**: Use platform-native secret stores with appropriate access controls. See examples: [Kubernetes Secrets](/helm/xmtpd/templates/secret.yaml) for database credentials and signing keys, or [AWS Secrets Manager integration](/terraform/aws/fargate-task-definition/main.tf) for ECS deployments.

- **Network segmentation**: Deploy databases in private subnets with security group rules restricting access to VPC CIDR blocks only. See [AWS RDS security group configuration](/terraform/examples/aws-complete/rds.tf) and [API security groups](/terraform/aws/xmtpd-api/security-groups.tf).

- **Access control**: Implement least-privilege IAM roles and Kubernetes service accounts with granular permissions. See [ECS task execution role](/terraform/aws/fargate-task-definition/main.tf) and [Kubernetes service accounts](/helm/xmtpd/templates/serviceaccount.yaml).

- **Private key protection**: Securely generate and store node signing keys in secrets management systems and register keys with blockchain before deployment. Keys stored as `XMTPD_SIGNER_PRIVATE_KEY` in [Kubernetes Secrets](/helm/xmtpd/templates/secret.yaml) or AWS Secrets Manager.

- **Database security**: Use managed PostgreSQL with SSL connections, automated backups, and deletion protection. See [Aurora PostgreSQL configuration](/terraform/examples/aws-complete/rds.tf).

- **Pod security standards**: Configure security contexts with read-only root filesystems, drop unnecessary Linux capabilities, and run as non-root users. See security context examples in [xmtpd values](/helm/xmtpd/values.yaml).

- **Health monitoring and alerting**: Deploy Prometheus-based monitoring with service discovery and health checks. See [Prometheus setup guide](/doc/k8s-prometheus-monitoring.md) to learn how to automatically scrape metrics from xmtpd pods using PodMonitor, visualize metrics in Grafana, and set alerts.

- **High availability and updates**: Deploy services across multiple availability zones. See [AWS multi-AZ Aurora](/terraform/examples/aws-complete/rds.tf) and [Kubernetes deployment strategies](/helm/xmtp-gateway/templates/deployment.yaml).

## Performance and operations

- **Data retention and pruning**: Implement automated database pruning to prevent data bloat and maintain node performance. See [database pruning guide](/doc/db-pruning.md) and [prune CronJob configuration](/helm/xmtpd/templates/prune-cronjob.yaml) for details.

## Networking notes

To learn more about the networking architecture xmtpd uses, see [XMTP Node Communication APIs](https://github.com/xmtp/xmtpd/blob/main/doc/networking.md) in the xmtpd repo.

Currently, xmtpd APIs are implemented using the [Connect-RPC](https://connectrpc.com/) library, which allows gRPC and gRPC-Web clients out of the box.

Because this library uses HTTP2 instead of gRPC, it also relies on requests headers to properly function, including CORS headers.

When using a load balancer in front of xmtpd, make sure it correctly forwards all headers, including CORS.

## Learn more

- [XMTP documentation](https://docs.xmtp.org)
- [Decentralizing XMTP](https://xmtp.org/decentralizing-xmtp)
- [Networking](https://github.com/xmtp/xmtpd/blob/main/doc/networking.md)

## Contribute

Contributions are welcome! See the [contributing guidelines](CONTRIBUTING.md) for details on how to get involved.
