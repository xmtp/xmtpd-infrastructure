# Tutorial: Set up Prometheus service discovery for xmtpd in Kubernetes using Helm

## Introduction

This tutorial walks you through setting up **Prometheus service discovery** for **xmtpd** in a Kubernetes cluster using **Helm**. By the end of this guide, Prometheus will automatically discover xmtpd pods and scrape their metrics.

## Prerequisites

Before proceeding, ensure you have the following:
- A running **Kubernetes cluster**
- `kubectl` installed and configured for your cluster
- **Helm** installed (Kubernetes package manager)

## Step 1. Install Helm

Helm simplifies the deployment of Kubernetes applications, including Prometheus.

First, install the Helm client. Follow the [official Helm installation instructions](https://helm.sh/docs/intro/install/), or use a package manager like Homebrew:

```bash
brew install kubernetes-helm
```

## Step 2. Deploy a Prometheus stack

Prometheus is an open-source monitoring and alerting toolkit. To set it up in Kubernetes, we will use the kube-prometheus-stack Helm chart.

### Add the Prometheus Helm repository

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
```

### Install Prometheus

```bash
helm install prometheus prometheus-community/kube-prometheus-stack --namespace default
```

This will install Prometheus along with necessary exporters and Grafana.
- **Prometheus** (for monitoring and alerting)
- **Grafana** (for visualization)
- **Alertmanager** (for alert processing)
- **Various Kubernetes exporters** for collecting cluster metrics

Check if Prometheus pods are running:

```bash
kubectl get pods -l app.kubernetes.io/name=prometheus -n default
```

## Step 3. Deploy xmtpd

xmtpd is the XMTP network's distributed messaging daemon.

Follow [Install xmtpd on Kubernetes using Helm charts](../helm/README.md) to deploy it.

Once installed, verify the xmtpd pods are running:

```bash
kubectl get pods -l app.kubernetes.io/name=xmtpd -n default
```

## Step 4. Configure Prometheus to discover xmtpd

To enable Prometheus to **automatically discover xmtpd pods** and scrape metrics, we will use a **PodMonitor**.

### Create a _PodMonitor_

A PodMonitor instructs Prometheus to scrape metrics **directly from pods**, rather than through a service.

Create a file named `xmtpd-podmonitor.yaml` in your local directory with the following content:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: xmtpd-podmonitor
  labels:
    release: prometheus  # Make sure it matches your Prometheus deployment
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: xmtpd  # This should match the pod labels
  namespaceSelector:
    matchNames:
      - default  # Set this to the namespace where your pods are running
  podMetricsEndpoints:
    - port: metrics
      path: /metrics
      interval: 15s
      relabelings:
        - sourceLabels: [ __meta_kubernetes_pod_label_app ]
          targetLabel: app
        - sourceLabels: [ __meta_kubernetes_pod_label_app_kubernetes_io_instance ]
          targetLabel: kubernetes_instance
        - sourceLabels: [ __meta_kubernetes_pod_label_app_kubernetes_io_name ]
          targetLabel: kubernetes_name
        - sourceLabels: [ __meta_kubernetes_pod_label_app_kubernetes_io_role ]
          targetLabel: service
```

### Apply the _PodMonitor_

```bash
kubectl apply -f xmtpd-podmonitor.yaml
```

## Step 5. Validate the installation

After setting up **Prometheus and PodMonitor**, verify that Prometheus is discovering xmtpd.

### Set up Prometheus UI

Port-forward the Prometheus service to access the UI:

```bash
kubectl port-forward svc/prometheus-kube-prometheus-prometheus 9090:9090
```

Now, open http://localhost:9090 in your browser.

### Verify the Service Discovery

1. In the Prometheus UI, navigate to **Status > Target health**.
2. Find your `xmtpd-podmonitor` target.
3. Ensure it is **marked as "UP"** which means Prometheus is successfully scraping metrics.

## Conclusion

You have successfully set up **Prometheus to monitor xmtpd** in your Kubernetes cluster. Prometheus now automatically discovers xmtpd pods and collects metrics.

### Next steps

- Use **Grafana** to visualize xmtpd metrics.
- Set up **alerts** in Prometheus to monitor for critical events.
- Explore **Prometheus queries (PromQL)** to analyze collected metrics.
