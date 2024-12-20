# Deploy xmtpd on Google Kubernetes Engine secured by SSL/TLS

[xmtpd](https://github.com/xmtp/xmtpd) (XMTP daemon) is the software that powers an XMTP node, enabling it to participate in the XMTP decentralized network.

Running xmtpd requires a secure and reliable infrastructure. In this tutorial, you'll learn how to deploy xmtpd on Google Kubernetes Engine (GKE) and secure it with SSL/TLS certificates issued by Let's Encrypt. You’ll use NGINX Ingress Controller to handle traffic routing and cert-manager to automate certificate management, ensuring your xmtpd services are safely exposed to the outside world.

> [!NOTE]
> While GKE comes tightly integrated with the [Google Compute Engine Ingress Controller](https://github.com/kubernetes/ingress-gce), which is efficient for many HTTP/HTTPS use cases, we opted for NGINX Ingress Controller for its flexibility and better handling of gRPC traffic, which is crucial for xmtpd services.

## Step 1. Install Helm

To install `cert-manager` and `nginx ingress` easily, you’ll use Helm, a package manager for Kubernetes.

First, install the Helm client. Follow the [official Helm installation instructions](https://helm.sh/docs/intro/install/), or use a package manager like Homebrew:

```bash
brew install kubernetes-helm
```

## Step 2. Deploy an NGINX Ingress Controller

A Kubernetes ingress controller is designed to be the access point for HTTP and HTTPS traffic to the software running within your cluster. The ingress-nginx-controller does this by providing an HTTP proxy service supported by your cloud provider's load balancer.

### How NGINX Ingress Controller works

Here’s a simplified flow of how the NGINX Ingress Controller works:

1. An Ingress resource is created, updated, or deleted in your Kubernetes cluster.
2. The Ingress Controller pod detects the change and dynamically updates the NGINX configuration.
3. NGINX processes incoming traffic based on the rules defined in the Ingress resource, such as routing based on hostnames or paths and handling SSL/TLS termination.
4. NGINX forwards the traffic to the appropriate backend services based on the Ingress rules.
5. The backend services handle the requests and provide responses.

### Add the NGINX Helm repo

1. Add the latest Helm repository for ingress-nginx:

    ```bash
    helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
    ```

2. Update the Helm repository with the latest charts:

    ```bash
    helm repo update
    ```

    Example output:

    ```bash
    Hang tight while we grab the latest from your chart repositories...
    ...Successfully got an update from the "ingress-nginx" chart repository
    Update Complete. ⎈ Happy Helming!⎈
    ```

### Install an NGINX Ingress Controller

1. Install an NGINX Ingress Controller using Helm:

    ```bash
    helm upgrade --install --namespace ingress-nginx --create-namespace ingress-nginx ingress-nginx/ingress-nginx
    ```

    Example output:

    ```bash
    Release "ingress-nginx" has been upgraded. Happy Helming!
    NAME: ingress-nginx
    ... lots of output ...
    ```

2. It can take a few minutes for the cloud provider to allocate and link a public IP address for the NGINX Ingress Controller. Once complete, you can get the external IP address by running:

    ```bash
    kubectl get svc -n ingress-nginx
    ```

    Example output:

    ```bash
    NAME                                 TYPE           CLUSTER-IP       EXTERNAL-IP      PORT(S)                      AGE
    ingress-nginx-controller             LoadBalancer   34.118.230.234   35.202.107.198   80:30925/TCP,443:32526/TCP   17h
    ingress-nginx-controller-admission   ClusterIP      34.118.237.149   <none>           443/TCP                      17h
    ```

    In this example, the NGINX Ingress Controller has been allocated the external IP `35.202.107.198`.

## Step 3. Assign a DNS name to your NGINX Ingress Controller external IP

Before you can install TLS certificates, your NGINX Ingress Controller needs a DNS entry.

The `EXTERNAL-IP` allocated to `ingress-nginx-controller` is where all incoming traffic will be routed. To assign a DNS name to your NGINX Ingress Controller, add this IP to a DNS zone you control. For example, `grpc.xmtp-partners.xyz`.

This tutorial assumes you are familiar with configuring DNS records and can complete this step based on your DNS provider’s instructions

## Step 4. Install cert-manager

Install cert-manager to work with Kubernetes to request a certificate and respond to the challenge to validate it. You can use Helm or plain Kubernetes manifests to install cert-manager.

cert-manager uses two primary types of custom Kubernetes resources, known as Custom Resource Definitions (CRDs), to configure and control how it operates and stores state: Issuers and Certificates. While Cert-Manager supports other resources, they're outside the scope of this tutorial.

### Cluster Issuers

An Issuer defines how cert-manager requests TLS certificates.

Issuers are specific to a single namespace in Kubernetes, but ClusterIssuer is a cluster-wide version. For convenience, this tutorial uses `ClusterIssuers`.

### Certificates

A Certificate resource enables you to specify the details of the certificate you want to request.

The resource references an issuer to define how the certificate is issued.

### Install cert-manager

To install cert-manager, run:

```bash
helm repo add jetstack https://charts.jetstack.io --force-update
```

```bash
helm install \
  cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --version v1.16.2 \
  --set crds.enabled=true \
  --set cainjector.enabled=true \
  --set global.leaderElection.namespace="cert-manager"
```

In some cases, the CRDs won't be installed properly. In this case, you can install them directly from static files:

```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.16.2/cert-manager.crds.yaml
```

### Verify your cert-manager installation

1. Verify that all CRDs have been installed successfully:

    ```bash
    kubectl get crds | grep cert-manager.io
    ```

    Example output:

    ```bash
    certificaterequests.cert-manager.io                    2024-12-10T21:38:49Z
    certificates.cert-manager.io                           2024-12-10T21:38:49Z
    challenges.acme.cert-manager.io                        2024-12-10T21:43:45Z
    clusterissuers.cert-manager.io                         2024-12-10T21:38:49Z
    issuers.cert-manager.io                                2024-12-10T21:38:50Z
    orders.acme.cert-manager.io                            2024-12-10T21:38:51Z
    ```

2. Verify that all cert-manager pods are running. If you have any doubts, check the pods for any error logs.

    ```bash
    kubectl get all -n cert-manager
    ```

    Example output:

    ```bash
    NAME                                           READY   STATUS      RESTARTS      AGE
    pod/cert-manager-c6b6b7554-7pq9s               1/1     Running     0             16h
    pod/cert-manager-cainjector-7bbbcf9b64-hdq9q   1/1     Running     1 (17h ago)   17h
    pod/cert-manager-startupapicheck-q4z6p         0/1     Completed   1             17h
    pod/cert-manager-webhook-569d869944-gnh2h      1/1     Running     0             17h

    NAME                              TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)            AGE
    service/cert-manager              ClusterIP   34.118.228.233   <none>        9402/TCP           17h
    service/cert-manager-cainjector   ClusterIP   34.118.231.242   <none>        9402/TCP           17h
    service/cert-manager-webhook      ClusterIP   34.118.236.190   <none>        443/TCP,9402/TCP   17h

    NAME                                      READY   UP-TO-DATE   AVAILABLE   AGE
    deployment.apps/cert-manager              1/1     1            1           17h
    deployment.apps/cert-manager-cainjector   1/1     1            1           17h
    deployment.apps/cert-manager-webhook      1/1     1            1           17h

    NAME                                                 DESIRED   CURRENT   READY   AGE
    replicaset.apps/cert-manager-c6b6b7554               1         1         1       17h
    replicaset.apps/cert-manager-cainjector-7bbbcf9b64   1         1         1       17h
    replicaset.apps/cert-manager-webhook-569d869944      1         1         1       17h

    NAME                                     STATUS     COMPLETIONS   DURATION   AGE
    job.batch/cert-manager-startupapicheck   Complete   1/1           3m33s      17h
    ```

## Step 5. Configure a Let's Encrypt issuer

Let’s Encrypt is a free, automated, and open Certificate Authority (CA) that provides SSL/TLS certificates. It simplifies obtaining and renewing certificates through automation.

In this tutorial, you'll set up two issuers for Let's Encrypt: Staging and production.

The Let's Encrypt production issuer has [very strict rate limits](https://letsencrypt.org/docs/rate-limits/). When you're experimenting and learning, you can easily hit those limits.

For this reason, you'll start by working with the Let's Encrypt staging issuer. Once you're happy that it's working correctly, yo can switch to using the production issuer.

> [!NOTE]
> You'll see a warning about untrusted certificates from the staging issuer, but this is expected.

1. Create a local file named `issuers.yaml` using the following example. Update the email address to your own. This email is required by Let's Encrypt to notify you of certificate expiration and updates.

    ```yaml
    apiVersion: cert-manager.io/v1
    kind: ClusterIssuer
    metadata:
      name: letsencrypt-staging
    spec:
      acme:
        # The ACME server URL
        server: https://acme-staging-v02.api.letsencrypt.org/directory
        # Email address used for ACME registration
        email: user@example.com
        # Name of a secret used to store the ACME account private key
        privateKeySecretRef:
          name: letsencrypt-staging
        # Enable the HTTP-01 challenge provider
        solvers:
          - http01:
              ingress:
                class: nginx
    ---
    apiVersion: cert-manager.io/v1
    kind: ClusterIssuer
    metadata:
      name: letsencrypt-prod
    spec:
      acme:
        # The ACME server URL
        server: https://acme-v02.api.letsencrypt.org/directory
        # Email address used for ACME registration
        email: user@example.com
        # Name of a secret used to store the ACME account private key
        privateKeySecretRef:
          name: letsencrypt-prod
        # Enable the HTTP-01 challenge provider
        solvers:
          - http01:
              ingress:
                class: nginx
    ```

2. Once created, apply the custom resource:

    ```bash
    kubectl apply -f issuers.yaml
    ```

## Step 6. Deploy xmtpd with an NGINX Ingress Controller

To deploy xmtpd and configure it with an NGINX Ingress Controller, follow these steps.

1. Fetch the Helm charts by cloning the [xmtpd-infrastructure](https://github.com/xmtp/xmtpd-infrastructure) repository and navigating to the Helm charts directory:

    ```bash
    git clone git@github.com:xmtp/xmtpd-infrastructure.git
    cd xmtpd-infrastructure/helm
    ```

2. Configure the deployment by creating an `xmtpd.yaml` file that sets all relevant `env` deployment variables as described in the [Helm README](../helm/README.md). Also, add this section:

    ```yaml
    # filename xmtpd.yaml
    # env: {}
    ingress:
      create: true
      host: <grpc.xmtp-partners.xyz>
      tls:
        certIssuer: letsencrypt-staging
        secretName: xmtp-tls-cert
    ```

3. To deploy xmtpd, install the Helm chart:

    ```bash
    helm install xmtpd ./xmtpd -f xmtpd.yaml
    ```

4. The following is an example of the Ingress resource created by the Helm chart:

    ```yaml
    # Source: xmtpd/templates/ingress.yaml
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: xmtpd
      labels:
        helm.sh/chart: xmtpd-0.1.0
        app.kubernetes.io/name: xmtpd
        app.kubernetes.io/instance: xmtpd
        app.kubernetes.io/version: "v0.1.1"
        app.kubernetes.io/managed-by: Helm
      annotations:
        kubernetes.io/ingress.class: nginx
        cert-manager.io/cluster-issuer: letsencrypt-staging
        nginx.ingress.kubernetes.io/ssl-redirect: "true"
        nginx.ingress.kubernetes.io/backend-protocol: "GRPC"
    spec:
      ingressClassName: nginx
      rules:
        - host: grpc.xmtp-partners.xyz
          http:
            paths:
              - path: /
                pathType: Prefix
                backend:
                  service:
                    name:  xmtpd
                    port:
                      number: 80
      tls:
      - secretName: xmtp-tls-cert
        hosts:
          - grpc.xmtp-partners.xyz
    ```

5. Validate the deployment by checking for the existence of the Ingress. It might take a few minutes for the Ingress to populate.

    ```bash
    kubectl get ing
    ```

    Example output:

    ```bash
    NAME               CLASS   HOSTS                          ADDRESS          PORTS     AGE
    xmtpd              nginx   grpc.xmtp-partners.xyz         35.202.107.198   80, 443   17h
    ```

6. In some cases, neither the NGINX nor the cert-manager will create the secret. To work around this chicken-and-egg problem, you can create an empty secret. For example:

    ```yaml
    # filename secret.yaml
    apiVersion: v1
    kind: Secret
    metadata:
      name: xmtp-tls-cert
    type: kubernetes.io/tls
    stringData:
      tls.key: ""
      tls.crt: ""
    ```

    Apply the secret:

    ```bash
    kubectl apply -f secret.yaml
    ```

## Step 7. Validate connectivity

In this step, you’ll verify that the DNS, TLS certificates, and Ingress are properly configured and that the xmtpd service is reachable.

1. Verify that the certificate has been mounted by the Ingress:

    ```bash
    kubectl describe ing xmtpd
    ```

    Example output:

    ```bash
    Namespace:        default
    Address:          35.202.107.198
    Ingress Class:    nginx
    Default backend:  <default>
    TLS:
      xmtp-tls-cert terminates grpc.xmtp-partners.xyz
    Rules:
      Host                         Path  Backends
      ----                         ----  --------
      grpc.xmtp-partners.xyz  
                                  /   xmtpd:80 (10.54.128.130:5050)
    ```

2. Verify that OpenSSL is able to connect to your DNS entry and show a certificate chain:

    ```bash
    openssl s_client -showcerts -connect grpc.xmtp-partners.xyz:443 < /dev/null
    ```

    Example output:

    ```bash
    Connecting to 35.202.107.198
    CONNECTED(00000005)
    ... lots of output ...
    Server certificate
    subject=CN=grpc.xmtp-partners.xyz
    issuer=C=US, O=Let's Encrypt (STAGING), CN=R11
    ---

    Verify return code: 0 (ok)
    ---
    DONE
    ```

3. As a final check, validate that the xmtpd service is reachable:

    ```bash
    grpcurl -insecure grpc.xmtp-partners.xyz:443 list 
    ```

    Example output:

    ```bash
    grpc.health.v1.Health
    grpc.reflection.v1.ServerReflection
    grpc.reflection.v1alpha.ServerReflection
    xmtp.xmtpv4.message_api.ReplicationApi
    ```

4. Once you've confirmed that DNS, TLS, and Ingress work, you can proceed to use the Let's Encrypt production certificates.

## Step 8. Upgrade to the Production cert-manager Cluster Issuer

Now that you’ve validated your DNS, TLS, and Ingress configurations using the Let’s Encrypt staging certificates, you’re ready to upgrade to production-grade certificates.

1. Update the `xmtpd.yaml` file to use the production `certIssuer`:

    ```yaml
    # filename xmtpd.yaml
    # env: {}
    ingress:
      create: true
      host: <grpc.xmtp-partners.xyz>
      tls:
        certIssuer: letsencrypt-prod
        secretName: xmtp-tls-cert
    ```

2. Apply the updated configuration by upgrading the Helm deployment:

    ```bash
    helm upgrade xmtpd ./xmtpd -f xmtpd.yaml
    ```

3. Validate the upgrade.

    To check gRPC endpoints, run:

    ```bash
    grpcurl grpc.xmtp-partners.xyz:443 list 
    ```

    Example output:

    ```
    grpc.health.v1.Health
    grpc.reflection.v1.ServerReflection
    grpc.reflection.v1alpha.ServerReflection
    xmtp.xmtpv4.message_api.ReplicationApi
    ```

    To check the health status of the xmtpd service, run:

    ```bash
    grpc-health-probe -tls -addr=grpc.xmtp-partners.xyz:443
    ```

    Expected output:

    ```bash
    status: SERVING
    ```

## Acknowledgements

This tutorial is based on the following articles:
- [Securing NGINX-ingress](https://cert-manager.io/docs/tutorials/acme/nginx-ingress/)
- [Deploy cert-manager on Google Kubernetes Engine (GKE) and create SSL certificates for Ingress using Let's Encrypt](https://cert-manager.io/docs/tutorials/getting-started-with-cert-manager-on-google-kubernetes-engine-using-lets-encrypt-for-ingress-ssl/)
- [How to Set Up SSL/TLS Certificates from Let’s Encrypt with NGINX Ingress Controller and Cert-Manager on AWS EKS](https://medium.com/@yakuphanbilgic3/how-to-set-up-ssl-tls-certificates-from-lets-encrypt-with-nginx-ingress-controller-and-9593b0eb8f23)
- [Ingress-Nginx Controller Installation Guide - GCE-GKE](https://kubernetes.github.io/ingress-nginx/deploy/#gce-gke)
- [Ingress-Nginx Controller Examples - Customization](https://kubernetes.github.io/ingress-nginx/examples/grpc/)