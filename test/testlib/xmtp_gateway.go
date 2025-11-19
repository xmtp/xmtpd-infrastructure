package testlib

import (
	"fmt"
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/random"
	corev1 "k8s.io/api/core/v1"
	"strings"
	"testing"
)

func installXMTPGateway(t *testing.T, options *helm.Options, helmChartReleaseName string) {
	if options.Version == "" {
		helm.Install(t, options, XmtpGatewayHelmChartPath, helmChartReleaseName)
	} else {
		helm.Install(t, options, "xmtp/xmtp-gateway", helmChartReleaseName)
	}
}

// StartGateway
/**
 * StartGateway starts a XMTP Gateway Service using the specified Helm options and namespace.
 *
 * @param t *testing.T - The testing context.
 * @param options *helm.Options - The Helm options for the installation.
 * @param namespace string - The namespace for the service.
 *
 * @return (string, string) - Returns the Helm chart release name and namespace.
 */
func StartGateway(t *testing.T, options *helm.Options, replicaCount int, namespace string) (string, string) {
	return startGatewayTemplate(t, options, replicaCount, namespace, "", installXMTPGateway, true)
}

type GatewayInstallationStep func(t *testing.T, options *helm.Options, helmChartReleaseName string)

func startGatewayTemplate(t *testing.T, options *helm.Options, replicaCount int, namespace string, releaseName string, installStep GatewayInstallationStep, awaitRunning bool) (helmChartReleaseName string, namespaceName string) {
	randomSuffix := strings.ToLower(random.UniqueId())

	helmChartReleaseName = releaseName
	if helmChartReleaseName == "" {
		helmChartReleaseName = fmt.Sprintf("xmtp-gateway-%s", randomSuffix)
	}

	if namespace == "" {
		namespaceName = CreateRandomNamespace(t, 4)
	} else {
		namespaceName = namespace
	}

	kubectlOptions := k8s.NewKubectlOptions("", "", namespaceName)
	options.KubectlOptions = kubectlOptions
	options.KubectlOptions.Namespace = namespaceName
	options.Logger = logger.Discard

	installStep(t, options, helmChartReleaseName)

	t.Cleanup(func() {
		helm.Delete(t, options, helmChartReleaseName, true)
	})

	if !awaitRunning {
		return
	}

	gatewayDeployment := helmChartReleaseName

	defer func() {
		// collect some useful diagnostics
		if t.Failed() {
			// ignore any errors. This is already failed
			_ = k8s.RunKubectlE(t, kubectlOptions, "describe", "deployment", gatewayDeployment)
		}
	}()

	AwaitNrReplicasScheduled(t, namespaceName, gatewayDeployment, replicaCount)

	pods := FindPodsFromChart(t, namespaceName, gatewayDeployment)

	for _, p := range pods {
		t.Cleanup(func() {
			if t.Failed() {
				// dump diagnostic info to test logs
				_ = k8s.RunKubectlE(t, kubectlOptions, "describe", "pod", p.Name)
			}

			for _, c := range p.Spec.Containers {
				cName := c.Name
				go GetAppLog(t, namespaceName, p.Name, cName, &corev1.PodLogOptions{Follow: true, Container: cName})
			}

			for _, c := range p.Spec.InitContainers {
				cName := c.Name
				go GetAppLog(t, namespaceName, p.Name, cName, &corev1.PodLogOptions{Follow: true, Container: cName})
			}

		})
	}

	AwaitNrReplicasReady(t, namespaceName, gatewayDeployment, replicaCount)

	return
}
