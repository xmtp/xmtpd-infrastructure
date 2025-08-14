# Install xmtpd on Kubernetes using Helm charts

**xmtpd (XMTP daemon)** is software that operators run to provide nodes in the decentralized XMTP network. The responsibilities of an xmtpd node include message storage and replication, censorship resistance, network resilience and recovery, and much more.

This repository includes Helm charts that you can use to deploy xmtpd on Kubernetes. Charts are curated application definitions for Helm. 

- To learn how to install Helm, see the Helm [README](https://github.com/helm/helm/tree/master/README.md). 

- To learn how to use Helm, see the Helm [Quickstart Guide](https://helm.sh/docs/intro/quickstart/).

- To learn more about Helm charts, see [Charts](https://helm.sh/docs/topics/charts/). 

## What does the xmtpd Helm chart contain?

xmtpd is composed of two key deployments: 

1. **Synchronization service**
    
    Handles data replication from other nodes to the local node, ensuring consistency across the network. 
    
2. **API service**
    
    Provides a client-facing gRPC endpoint for both nodes and client applications (like Converse) to interact with the system.
    
Both services support multiple replicas for high availability. However, adding replicas doesn't enhance the overall throughput of the service. As such, we recommend running exactly two replicas of each.

## Prerequisites

Before diving into the installation process, ensure you have the following:

1. **Kubernetes cluster**: A running Kubernetes cluster with sufficient resources. To learn more, see [Kubernetes documentation](https://kubernetes.io/docs/home/). If you want to test the charts locally, you can use [Docker Desktop](https://www.docker.com/get-started/)
2. **Helm**: A package manager for Kubernetes. To learn more, see [Helm documentation](https://helm.sh/docs/).
   ```bash
   brew install helm
   ```
3. **PostgreSQL database**: xmtpd relies on a PostgreSQL database. Tools like Bitnami’s [PostgreSQL Helm Charts](https://github.com/bitnami/charts/tree/main/bitnami/postgresql) can simplify database provisioning. xmtpd has been tested with Postgres versions 13 and newer. To learn more, see [PostgreSQL documentation](https://www.postgresql.org/docs/).
   ```bash
    helm repo add bitnami https://charts.bitnami.com/bitnami
    helm repo update
    ```

4. **XMTP Helm Repository**: Fetch the pre-packaged repository
    ```bash
    helm repo add xmtp https://xmtp.github.io/xmtpd-infrastructure
    helm repo update
    ```

## Step 1. Get an Alchemy account

To run xmtpd, you need an Alchemy account.

1. See [Create an Alchemy API Key](https://docs.alchemy.com/docs/alchemy-quickstart-guide#1key-create-an-alchemy-api-key) and log in to Alchemy.

2. Go to the [XMTP Ropsten Chain page](https://dashboard.alchemy.com/chains/xmtp?network=XMTP_ROPSTEN) and set the **Network Status** to ***Enabled***.

3. In the **API URL** column, click **Copy** to copy the XMTP app WebSockets endpoint along with its API key. It should use the format `wss://xmtp-ropsten.g.alchemy.com/v2/<apikey>`.

4. You will use this endpoint for the `XMTPD_APP_CHAIN_WSS_URL` config option.

5. Repeat for the HTTPs endpoint and set `XMTPD_APP_CHAIN_RPC_URL` to `https://xmtp-ropsten.g.alchemy.com/v2/<apikey>`

5. Repeat the steps for [Base Sepolia](https://dashboard.alchemy.com/chains/base?network=BASE_SEPOLIA) and set both `XMTPD_SETTLEMENT_CHAIN_WSS_URL` to `wss://base-sepolia.g.alchemy.com/v2/<apikey>` and `XMTPD_SETTLEMENT_CHAIN_RPC_URL` to `https://base-sepolia.g.alchemy.com/v2/<apikey>`.

## Step 2: Register your node

The xmtpd node software is open source, allowing anyone to run a node. However, only registered nodes can join the XMTP testnet.

To enable your node to join the XMTP testnet, register its public key and DNS address on the blockchain.

The node registration process is currently managed by [Ephemera](https://ephemerahq.com/), stewarding the development and adoption of XMTP. To learn more about the XMTP network node operator qualification criteria and selection process, see [XIP-54](https://community.xmtp.org/t/xip-54-xmtp-network-node-operator-qualification-criteria/868).

### **Step 2.1: Get all keys**

If you need a new private key, you can generate one using the xmtpd CLI Docker image. 

```bash
docker run ghcr.io/xmtp/xmtpd-cli:latest generate-key | jq
```

Example response:

```json
{
  "level": "INFO",
  "time": "2024-10-15T13:21:14.036-0400",
  "message": "generated private key",
  "private-key": "0x7d3cd4989b92c593db9a4b3ac1c2a5d542efad058b2a83e26c3467392b29c6f9",
  "public-key": "0x03da53968d81f4eb3c9dd8b96617575767ec0cccbd28103b2cfd7f1511bb282d30",
  "address": "0x9419db765e6b469edc028ffa72ba2944f2bad169"
}
```

If you already have a private key, you can extract relevant public details using the xmtpd CLI Docker image.

```bash
docker run ghcr.io/xmtp/xmtpd-cli:latest get-pub-key --private-key 0xa9b48d687f450ea99a5faaae1be096ddb49487cb28393d3906d7359ede6ea460 | jq
```

 Example response:

```json
{
  "level": "INFO",
  "time": "2024-10-15T13:21:51.276-0400",
  "message": "parsed private key",
  "pub-key": "0x027a64295b98e48682cb77be1b990d4ecf8f1a86badf051df0af123e6fe3790e3f",
  "address": "0x9419db765e6b469edc028ffa72ba2944f2bad169"
}
```

### **Step 2.2: Provide your node public key and address**

Provide your node public key and address to the XMTP Security Council team member working with you to register your node.

Ensure that public key and address values are correct because once registered, they are immutable and cannot be changed.

## Step 3: Set up dependencies

### Confirm you have the newest version of the charts
```bash
helm search repo xmtp
NAME                            CHART VERSION   APP VERSION     DESCRIPTION                                 
xmtp/xmtp-payer                 0.4.0           v0.3.0          A Helm chart for XMTP Payer                 
xmtp/xmtpd                      0.4.0           v0.3.0          A Helm chart for XMTPD                      
xmtp/mls-validation-service     0.1.0           v0.1.0          A Helm chart for XMTP MLS Validation Service
```

If your versions are outdated, run `helm repo update`.

### Install the PostgreSQL database

xmtpd requires a PostgreSQL database that's accessible from the Kubernetes cluster. Use a Helm chart or your preferred tool to set up the database.

For example, run:

```bash
helm install postgres bitnami/postgresql --set auth.postgresPassword=postgres
```

> [!IMPORTANT]
> For your convenience while testing this flow, this command sets the PostgresSQL database password to `postgres`. Be sure to set a secure password before going live on the XMTP testnet.

You’ll use values in the response to update the `XMTPD_DB_WRITER_CONNECTION_STRING` value in your `xmtpd.yaml` configuration.

### Install the MLS validation service

The MLS validation service is a stateless, horizontally scalable service that xmtpd depends on.

Install it using the Helm chart from the repository. To do this, run:

```bash
helm install mls-validation-service xmtp/mls-validation-service
```

## Step 4: Install the xmtpd node

### Configure the xmtpd node

Before deploying xmtpd, create a configuration file (`xmtpd.yaml`) to specify the required environment variables. 

Here is a sample configuration:

```yaml
env:
  secret:
    XMTPD_DB_WRITER_CONNECTION_STRING: "postgres://<username>:<password>@<host-service>:<port>/<database>?sslmode=disable"
    XMTPD_SIGNER_PRIVATE_KEY: "<private-key>"
    XMTPD_APP_CHAIN_WSS_URL: "wss://xmtp-ropsten.g.alchemy.com/v2/<apikey>"
    XMTPD_SETTLEMENT_CHAIN_WSS_URL: "wss://base-sepolia.g.alchemy.com/v2/<apikey>"
    XMTPD_APP_CHAIN_RPC_URL: "https://xmtp-ropsten.g.alchemy.com/v2/<apikey>"
    XMTPD_SETTLEMENT_CHAIN_RPC_URL: "https://base-sepolia.g.alchemy.com/v2/<apikey>"
    XMTPD_MLS_VALIDATION_GRPC_ADDRESS: "http://mls-validation-service.default.svc.cluster.local:50051"
    XMTPD_METRICS_ENABLE: "true"
    XMTPD_REFLECTION_ENABLE: "true"
    XMTPD_LOG_LEVEL: "debug"
```

As of helm chart release `0.3.0` you do not have to specify contract addresses.

Replace placeholder values with actual credentials and configurations:

- The `XMTPD_DB_WRITER_CONNECTION_STRING` is constructed as `postgres://<username>:<password>@<host-service>:<port>/<database>?sslmode=disable`
    - Replace `<username>` with the username for pg-postgresql. The default value is usually `postgres`.
    - Replace `<password>` with the password for pg-postgresql. If the password includes a special character, be sure to URL encode it.
    - Replace `<host-service>` with `<helm-chart-name>-postgresql.<namespace>.svc.cluster.local`. The default value is usually `postgres-postgresql.default.svc.cluster.local`.
    - Replace `<port>` with the port for pg-postgresql. The default value is usually `5432`.
    - Replace `<database>` with the database name for pg-postgresql. The default value is usually `postgres`.

    If you are following the setup steps in this document, the full connection string will be: `postgres://postgres:postgres@postgres-postgresql.default.svc.cluster.local:5432/postgres?sslmode=disable`
- Replace `<apikey>` with the key from your full Alchemy URL    
- Replace `<private-key>` with the private key for your registered node.

### Install xmtpd

Use Helm to deploy the xmtpd node from the repository. In the directory where you created your `xmtpd.yaml` configuration file, run:

```bash
helm install xmtpd xmtp/xmtpd -f xmtpd.yaml
```

## Step 5: Validate the installation

After installation, verify that all the required pods are running:

```bash
kubectl get pods
```

Expected output:

```bash
NAME                                      READY   STATUS    RESTARTS   AGE
mls-validation-service-75b6b96f79-kp4jq   1/1     Running   0          6h30m
pg-postgresql-0                           1/1     Running   0          6h26m
xmtpd-api-7dd49f6b88-mfwls                1/1     Running   0          4h56m
xmtpd-sync-7dd49f6b88-mfwls               1/1     Running   0          4h56m
```

> [!TIP]
> If you see a `no matching public key found in registry` error, you can resolve it by [registering your node](#step-2-register-your-node).

## Install the XMTP Payer Helm chart

> [!IMPORTANT]
> xmptd node operators do not need to run the Payer service. This service might be required by developers building apps on XMTP.

The payer service does not depend on the xmtpd service, a database, or the MLS service.

It does require access to the public internet to read state from the blockchain.

It also requires write access to all known XMTP nodes in the cluster.

You can use the config defined in [Step 4](#Step-4-Install-the-xmtpd-node) as a starting point.

Set `XMTPD_PAYER_PRIVATE_KEY` to a key to a wallet that has been funded and can be used to pay for blockchain messages and XMTP system messages.

### Install the Helm chart

```bash
helm install xmtp-payer xmtp/xmtp-payer -f xmtpd.yaml
```

Once you have successfully installed the chart, you should see 1 pod running.
You can confirm via:

```bash
kubectl get pods

NAME                                      READY   STATUS    RESTARTS   AGE
xmtp-payer-7978dbcb8-mnvxx                1/1     Running   0          9m48s
```

## Kubernetes ingress

We provide an [ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) for both XMTP services.

Its specific configuration will depend on the type of cloud being used.

The Kubernetes ingress can:
- Expose an external IP address
- Handle TLS termination
- Load balance and route

To enable the ingress, set the following:

```yaml
# filename xmtpd.yaml
ingress:
  enable: true
  className: <your ingress controller>
```

The class name will depend on the type of [Ingress Controller](
https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/) your environment supports.

The ingress will automatically route all traffic to the `xmtpd service`.

We recommend using the [Ingress-Nginx Controller](https://kubernetes.github.io/ingress-nginx/).

### Terminate TLS

One of the easiest ways to terminate TLS is to use [Let's Encrypt cert-manager](https://cert-manager.io/docs/getting-started/).

The configuration will depend on your cloud provider.

To configure the ingress to use `cert-manager`, set the following:

```yaml
ingress:
  enable: true
  className: <your ingress controller>
  host: <example.com>
  tls:
    certIssuer: <letsencrypt-production>
    secretName: <tls-certs>
```

For a comprehensive guide on how to terminate TLS in GKE, see [Deploy xmtpd on Google Kubernetes Engine secured by SSL/TLS](../doc/nginx-cert-gke.md)

## XMTPD Database Pruning

We offer a convenient Kubernetes CronJob that prunes the XMTPD database following the XMTP protocol data retention protocol.

To read more about the infrastructure agnostic database pruning policy, go to [XMTPD Database pruning](../doc/db-pruning.md).

To enable the pruning job, set the following:

```yaml
# filename xmtpd.yaml
prune:
  create: true
```

## What’s next

> [!TIP]
> Have feedback or ideas for improving xmtpd? [Open an issue](https://github.com/xmtp/xmtpd/issues) in the xmtpd GitHub repo.
