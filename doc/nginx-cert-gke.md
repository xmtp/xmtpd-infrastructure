# How to Set Up SSL/TLS Certificates from Let’s Encrypt with NGINX Ingress Controller and Cert-Manager on Google GKE

In this howto we will create a Nginx ingress controller on Google GKE and secure it with TLS certificates issued by Let's Encrypt.
We will use this strategy to expose the XMTPD services to the outside world.

Normally, GKE comes with the [GCE ingress controller](https://github.com/kubernetes/ingress-gce), but we had issues with gRPC.

## Step 1) Install Helm

The easiest way to install `cert-manager` and `nginx ingress` is to use Helm.

First, ensure the Helm client is installed following the Helm installation instructions:
```bash
brew install kubernetes-helm
```

## Step 2) Deploy NGINX Ingress Controller

A kubernetes ingress controller is designed to be the access point for HTTP and HTTPS traffic to the software running within your cluster. 
The ingress-nginx-controller does this by providing an HTTP proxy service supported by your cloud provider's load balancer.

### How NGINX Ingress Controller Works

Here’s a simplified flow of how the NGINX Ingress Controller works:
1) An Ingress resource is created, updated, or deleted in your Kubernetes cluster.
2) The Ingress Controller pod detects the change and dynamically updates the NGINX configuration.
3) NGINX processes incoming traffic based on the rules defined in the Ingress resource, such as routing based on hostnames or paths and handling SSL/TLS termination.
4) NGINX forwards the traffic to the appropriate backend services based on the Ingress rules.
5) The backend services handle the requests and provide responses.

### Add NGINX Helm Repo

Add the latest helm repository for the ingress-nginx
```bash
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
```

Update the helm repository with the latest charts:
```bash
$ helm repo update
Hang tight while we grab the latest from your chart repositories...
...Successfully got an update from the "ingress-nginx" chart repository
Update Complete. ⎈ Happy Helming!⎈
```

### Install The Chart

Use helm to install an NGINX Ingress controller:
```bash
helm upgrade --install --namespace ingress-nginx --create-namespace ingress-nginx ingress-nginx/ingress-nginx

Release "ingress-nginx" has been upgraded. Happy Helming!
NAME: ingress-nginx
... lots of output ...
```

It can take a minute or two for the cloud provider to provide and link a public IP address.
When it is complete, you can see the external IP address using the kubectl command:
```bash
$ kubectl get svc -n ingress-nginx
NAME                                 TYPE           CLUSTER-IP       EXTERNAL-IP      PORT(S)                      AGE
ingress-nginx-controller             LoadBalancer   34.118.230.234   35.202.107.198   80:30925/TCP,443:32526/TCP   17h
ingress-nginx-controller-admission   ClusterIP      34.118.237.149   <none>           443/TCP                      17h
```

In our example, the ingress has been allocated the public IP `35.202.107.198`.

## Step 3) Assign a DNS name

Before we can install TLS certificates, we have to have a DNS entry.
The external IP that is allocated to the ingress-controller is the IP to which all incoming traffic should be routed.
To enable this, add it to a DNS zone you control, for example as `grpc.xmtp-partners.xyz`

This howto assumes you know how to assign a DNS entry to an IP address and will do so.

## Step 4) Deploy cert-manager

We need to install cert-manager to do the work with Kubernetes to request a certificate and respond to the challenge to validate it.
We can use Helm or plain Kubernetes manifests to install cert-manager.

cert-manager mainly uses two different custom Kubernetes resources - known as CRDs - to configure and control how it operates, as well as to store state.
These resources are Issuers and Certificates.
There are other resources that are out of scope of this document.

### Cluster Issuers
An Issuer defines how cert-manager will request TLS certificates.
Issuers are specific to a single namespace in Kubernetes, but the ClusterIssuer is a cluster-wide version.
For convenience, we are using `ClusterIssuers`.

### Certificates
Certificates resources allow you to specify the details of the certificate you want to request.
They reference an issuer to define how they'll be issued.


### Install Cert Manager
```bash
helm install \
  cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --version v1.16.2 \
  --set crds.enabled=true
 ```

In some cases the CRDs will not get installed properly.
Alternatively you can install them direcly from static files:
```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.16.2/cert-manager.crds.yaml
```

### Verify Installation

First, verify that all CRDs have been installed successfully:
```bash
kubectl get crds | grep cert-manager.io 
certificaterequests.cert-manager.io                    2024-12-10T21:38:49Z
certificates.cert-manager.io                           2024-12-10T21:38:49Z
challenges.acme.cert-manager.io                        2024-12-10T21:43:45Z
clusterissuers.cert-manager.io                         2024-12-10T21:38:49Z
issuers.cert-manager.io                                2024-12-10T21:38:50Z
orders.acme.cert-manager.io                            2024-12-10T21:38:51Z
```

Second, check that all cert-manager pods are running:
```bash
kubectl get all -n cert-manager
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

When in doubt, check the pods for any error logs.

## Step 5) Configure a Let's Encrypt Issuer

We'll set up two issuers for Let's Encrypt in this example: staging and production.

The Let's Encrypt production issuer has very [strict rate limits](https://letsencrypt.org/docs/rate-limits/).
When you're experimenting and learning, it can be very easy to hit those limits.
Because of that risk, we'll start with the Let's Encrypt staging issuer, and once we're happy that it's working we'll switch to the production issuer.

Note that you'll see a warning about untrusted certificates from the staging issuer, but that's totally expected.

Create this definition locally and update the email address to your own.
This email is required by Let's Encrypt and used to notify you of certificate expiration and updates.

Create a file named `issuers.yaml` such as:
```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
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
kind: Issuer
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

Once created, apply the custom resource:
```bash
kubectl apply -f issuers.yaml
```

## Step 6) Deploy XMTPD with an Ingress

These helm charts provide convenient templates for cert-manager and nginx.

Fetch the repo:
```bash
git clone git@github.com:xmtp/xmtpd-infrastructure.git
cd xmtpd-infrastructure/helm
```

Create a `xmtpd.yaml` file which sets all relevant `env` variables as described in the [Helm Readme](../helm/README.md)

Additionally, you will need this section:
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

To install the helm chart, run
```bash
helm install xmtpd ./xmtpd -f xmtpd.yaml
```

Among other things, this will create an Ingress such as:
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

To validate, check for the existence of the Ingress:
```bash
kubectl get ing
NAME               CLASS   HOSTS                          ADDRESS          PORTS     AGE
xmtpd              nginx   grpc.xmtp-partners.xyz         35.202.107.198   80, 443   17h
```

It might take a few minutes for the Ingress to populate.

In some cases neither the NGINX nor the cert-manager will create the secret.
To work around this chicken-and-egg problem, you can create an empty secret such as:
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
```bash
kubectl apply -f secret.yaml
```

## Step 7) Validate the Connectivity

You can verify that the certificate has been mounted by the Ingress:
```bash
kubectl describe ing xmtpd 
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

OpenSSL should be able to connect to your DNS entry and show a certificate chain:
```bash
openssl s_client -showcerts -connect grpc.xmtp-partners.xyz:443 < /dev/null
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

As a final check, validate that the XMTPD service is reachable:
```bash
grpcurl -insecure grpc.xmtp-partners.xyz:443 list 
grpc.health.v1.Health
grpc.reflection.v1.ServerReflection
grpc.reflection.v1alpha.ServerReflection
xmtp.xmtpv4.message_api.ReplicationApi
```

Once you have confirmed that DNS, TLS and Ingress work, you can use the production certificates.

## Step 8) Upgrade to Production Cert Manager Cluster Issuer

Update the yaml file to use the production certIssuer:
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

Re-install the chart with the new setting:
```bash
helm upgrade xmtpd ./xmtpd -f xmtpd.yaml
```

To validate:
```bash
grpcurl grpc.xmtp-partners.xyz:443 list 
grpc.health.v1.Health
grpc.reflection.v1.ServerReflection
grpc.reflection.v1alpha.ServerReflection
xmtp.xmtpv4.message_api.ReplicationApi

grpc-health-probe -tls -addr=grpc.xmtp-partners.xyz:443
status: SERVING
```

## Acknowledgements
This howto was based on the following articles:
- https://cert-manager.io/docs/tutorials/acme/nginx-ingress/
- https://cert-manager.io/docs/tutorials/getting-started-with-cert-manager-on-google-kubernetes-engine-using-lets-encrypt-for-ingress-ssl/
- https://medium.com/@yakuphanbilgic3/how-to-set-up-ssl-tls-certificates-from-lets-encrypt-with-nginx-ingress-controller-and-9593b0eb8f23
- https://kubernetes.github.io/ingress-nginx/deploy/#gce-gke
- https://kubernetes.github.io/ingress-nginx/examples/grpc/