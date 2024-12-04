package testlib

import (
	"fmt"
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	corev1 "k8s.io/api/core/v1"
	"strings"
	"testing"
)

func installXMTPPayer(t *testing.T, options *helm.Options, helmChartReleaseName string) {
	if options.Version == "" {
		helm.Install(t, options, XMTP_PAYER_HELM_CHART_PATH, helmChartReleaseName)
	} else {
		helm.Install(t, options, "xmtp/xmtpd", helmChartReleaseName)
	}
}

// StartPayer
/**
 * StartPayer starts a XMTP Payer Service using the specified Helm options and namespace.
 *
 * @param t *testing.T - The testing context.
 * @param options *helm.Options - The Helm options for the installation.
 * @param namespace string - The namespace for the service.
 *
 * @return (string, string) - Returns the Helm chart release name and namespace.
 */
func StartPayer(t *testing.T, options *helm.Options, replicaCount int, namespace string) (string, string) {
	return startPayerTemplate(t, options, replicaCount, namespace, "", installXMTPPayer, true)
}

type PayerInstallationStep func(t *testing.T, options *helm.Options, helmChartReleaseName string)

func startPayerTemplate(t *testing.T, options *helm.Options, replicaCount int, namespace string, releaseName string, installStep PayerInstallationStep, awaitRunning bool) (helmChartReleaseName string, namespaceName string) {
	randomSuffix := strings.ToLower(random.UniqueId())

	helmChartReleaseName = releaseName
	if helmChartReleaseName == "" {
		helmChartReleaseName = fmt.Sprintf("xmtp-payer-%s", randomSuffix)
	}

	if namespace == "" {
		namespaceName = CreateRandomNamespace(t, 4)
	} else {
		namespaceName = namespace
	}

	kubectlOptions := k8s.NewKubectlOptions("", "", namespaceName)
	options.KubectlOptions = kubectlOptions
	options.KubectlOptions.Namespace = namespaceName

	installStep(t, options, helmChartReleaseName)

	AddTeardown(TEARDOWN_PAYER, func() {
		helm.Delete(t, options, helmChartReleaseName, true)
	})

	if !awaitRunning {
		return
	}

	payerDeployment := helmChartReleaseName

	defer func() {
		// collect some useful diagnostics
		if t.Failed() {
			// ignore any errors. This is already failed
			_ = k8s.RunKubectlE(t, kubectlOptions, "describe", "deployment", payerDeployment)
		}
	}()

	AwaitNrReplicasScheduled(t, namespaceName, payerDeployment, replicaCount)

	pods := FindPodsFromChart(t, namespaceName, payerDeployment)

	for _, pod := range pods {
		AddTeardown(TEARDOWN_PAYER, func() {
			if t.Failed() {
				// dump diagnostic info to test logs
				_ = k8s.RunKubectlE(t, kubectlOptions, "describe", "pod", pod.Name)
			}
			// collect logs
			go GetAppLog(t, namespaceName, pod.Name, "", &corev1.PodLogOptions{Follow: true})

		})
	}

	AwaitNrReplicasReady(t, namespaceName, payerDeployment, replicaCount)

	return
}
