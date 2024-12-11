# XMTPD Helm Charts

**⚠️ Experimental:** This software is in early development. Expect frequent changes and unresolved issues.

`xmtpd` (XMTP daemon) is an experimental version of XMTP node software. It is **not** the node software that currently forms the XMTP network.

This repository includes Helm Charts that can be used to deploy XMTPD in Kubernetes. You may also contribute by submitting enhancement requests if you like. Charts are curated application definitions for Helm. For more information about installing and using Helm, see its
[README.md](https://github.com/helm/helm/tree/master/README.md). To get a quick introduction to Helm Charts see this [chart document](https://github.com/helm/helm/blob/master/docs/charts.md). For more information on using Helm, refer to the [Helm's documentation](https://github.com/kubernetes/helm#docs).

## XMTPD Helm Chart Installation

To add the XMTPD charts to your local helm repository, clone this repository locally.

Eventually XMTP will provide a public Helm Charts release.

## 1) Dependencies

XMTPD needs a Postgres database to be running accessible from the cluster.
For example, you can use [Bitnami PG Helm Charts](https://bitnami.com/stack/postgresql/helm), or any other tooling to provision a Postgres database.

## 2) Installing MLS Validation Service

The XMTP MLS Validation Service does not have any external dependencies.
It is a stateless horizontally-scalable service.

To install, run:
```bash
helm install mls-validation-service mls-validation-service/
```

## 3) Installing the XMTPD Node

The XMPT daemon depends on the following:
- a PG database running locally
- the MLS validation service running locally. XMTP Expects the format `http|https://<DNS|IP>[:port]`
- a blockchain with the [Nodes contract](https://github.com/xmtp/xmtpd)
- a private key
- the private key needs to be registered with the blockchain smart contract and a NodeID has to be issued. For more info see [Onboarding](https://github.com/xmtp/xmtpd/blob/main/doc/onboarding.md)

Create a `xmtpd.yaml` file and fill out all required variables:
```yaml
env:
  secret:
    XMTPD_DB_WRITER_CONNECTION_STRING: "<postgres://postgres:postgres@pg-postgresql.default.svc.cluster.local:5432/postgres?sslmode=disable>"
    XMTPD_SIGNER_PRIVATE_KEY: "<private-key>"
    XMTPD_PAYER_PRIVATE_KEY: "<private-key>"
    XMTPD_CONTRACTS_RPC_URL: "<https://rpc-testnet-staging-88dqtxdinc.t.conduit.xyz/>"
    XMTPD_MLS_VALIDATION_GRPC_ADDRESS: "<http://mls-validation-service.default.svc.cluster.local:50051>"
    XMTPD_CONTRACTS_CHAIN_ID: "34498"
    XMTPD_CONTRACTS_NODES_ADDRESS: "<nodes-address>"
    XMTPD_CONTRACTS_MESSAGES_ADDRESS: "<messages-address>"
    XMTPD_CONTRACTS_IDENTITY_UPDATES_ADDRESS: "<identity-address>"
    XMTPD_METRICS_ENABLE: "true"
    XMTPD_REFLECTION_ENABLE: "true"
    XMTPD_LOG_LEVEL: "debug"
```

Install the helm chart
```bash
helm install xmtpd xmtpd/ -f xmtpd.yaml
```

### Validating the installation

Once you have successfully installed all charts, including a DB, you should see 4 pods running.
You can confirm via:
```bash
$ kubectl get pods
NAME                                      READY   STATUS    RESTARTS   AGE
mls-validation-service-75b6b96f79-kp4jq   1/1     Running   0          6h30m
pg-postgresql-0                           1/1     Running   0          6h26m
xmtpd-api-7dd49f6b88-mfwls                1/1     Running   0          4h56m
xmtpd-sync-7dd49f6b88-dn28d               1/1     Running   0          4h56m
```


## XMTP Payer Helm Chart Installation

If you are a XMTPD Node operator, you will not need to run the Payer service.
This service might be required by application developers that work on top of the XMTP protocol.

The payer service does not depend on the XMTPD service, a database, or the MLS service.
It does require access to the public internet to read state from the blockchain.
It also requires write access to all known XMTP nodes in the cluster.

As of XMTPD docker image v0.1.1, all services require the same configuration options even though some might be unused.
As such, you can reuse the same config that we described in [Step 3](#3-Installing-the-XMTPD-Node).

Install the helm chart
```bash
helm install xmtp-payer xmtp-payer/ -f xmtpd.yaml
```
Once you have successfully installed the chart, you should see 2 pods running.
You can confirm via:
```bash
$ kubectl get pods
NAME                                      READY   STATUS    RESTARTS   AGE
xmtp-payer-7978dbcb8-mnvxx                1/1     Running   0          9m48s
xmtp-payer-7978dbcb8-vkwsv                1/1     Running   0          9m48s
```

## Kubernetes Ingress

We provide an [ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) for both XMTP services.
It's specific configuration will depend on the type of cloud being used.
It can:
- expose an external IP address
- handle TLS termination
- load balance and route

To enable the ingress, you have to set:
```yaml
# filename xmtpd.yaml
ingress:
  enable: true
  className: <your ingress controller>
```

The class name will depend on the type of [Ingress Controller](
https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/) that your environment supports.

The ingress will automatically route all traffic to the `xmtpd service`.

We recommend using the [`nginx ingress controller`](https://kubernetes.github.io/ingress-nginx/)

### TLS Termination

One of the easiest ways to terminate TLS is to use [`Let's Encrypt cert-manager`](https://cert-manager.io/docs/getting-started/).
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

For a comprehensive guide on how to terminate TLS in GKE, you can read our [GKE+Let'Encrypt+NGINX Howto](../doc/nginx-cert-gke.md)