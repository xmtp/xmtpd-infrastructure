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

func installXMTPD(t *testing.T, options *helm.Options, helmChartReleaseName string) {
	if options.Version == "" {
		helm.Install(t, options, XmtpdHelmChartPath, helmChartReleaseName)
	} else {
		helm.Install(t, options, "xmtp/xmtpd", helmChartReleaseName)
	}
}

// StartXMTPD
/**
 * StartXMTPD starts a XMTPD Node using the specified Helm options and namespace.
 *
 * @param t *testing.T - The testing context.
 * @param options *helm.Options - The Helm options for the installation.
 * @param namespace string - The namespace for the node.
 *
 * @return (string, string) - Returns the Helm chart release name and namespace.
 */
func StartXMTPD(t *testing.T, options *helm.Options, replicaCount int, namespace string) (string, string) {
	return startXMTPDTemplate(t, options, replicaCount, namespace, "", installXMTPD, true)
}

type XMTPDInstallationStep func(t *testing.T, options *helm.Options, helmChartReleaseName string)

func startXMTPDTemplate(t *testing.T, options *helm.Options, replicaCount int, namespace string, releaseName string, installStep XMTPDInstallationStep, awaitRunning bool) (helmChartReleaseName string, namespaceName string) {
	randomSuffix := strings.ToLower(random.UniqueId())

	helmChartReleaseName = releaseName
	if helmChartReleaseName == "" {
		helmChartReleaseName = fmt.Sprintf("xmtpd-%s", randomSuffix)
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

	xmtpdDeploymentSync := fmt.Sprintf("%s-sync", helmChartReleaseName)
	xmtpdDeploymentApi := fmt.Sprintf("%s-api", helmChartReleaseName)
	xmtpdDeploymentIndexer := fmt.Sprintf("%s-indexer", helmChartReleaseName)
	xmtpdDeploymentReporting := fmt.Sprintf("%s-reporting", helmChartReleaseName)

	defer func() {
		// collect some useful diagnostics
		if t.Failed() {
			// ignore any errors. This is already failed
			_ = k8s.RunKubectlE(t, kubectlOptions, "describe", "deployment", xmtpdDeploymentSync)
			_ = k8s.RunKubectlE(t, kubectlOptions, "describe", "deployment", xmtpdDeploymentApi)
			_ = k8s.RunKubectlE(t, kubectlOptions, "describe", "deployment", xmtpdDeploymentIndexer)
			_ = k8s.RunKubectlE(t, kubectlOptions, "describe", "deployment", xmtpdDeploymentReporting)
		}
	}()

	AwaitNrReplicasScheduled(t, namespaceName, xmtpdDeploymentSync, replicaCount)
	AwaitNrReplicasScheduled(t, namespaceName, xmtpdDeploymentApi, replicaCount)
	AwaitNrReplicasScheduled(t, namespaceName, xmtpdDeploymentIndexer, replicaCount)
	AwaitNrReplicasScheduled(t, namespaceName, xmtpdDeploymentReporting, replicaCount)

	podsSync := FindPodsFromChart(t, namespaceName, xmtpdDeploymentSync)
	podsApi := FindPodsFromChart(t, namespaceName, xmtpdDeploymentApi)
	podsIndexer := FindPodsFromChart(t, namespaceName, xmtpdDeploymentIndexer)
	podsReporting := FindPodsFromChart(t, namespaceName, xmtpdDeploymentReporting)

	allPods := append(podsSync, podsApi...)
	allPods = append(allPods, podsIndexer...)
	allPods = append(allPods, podsReporting...)
	for _, pod := range allPods {
		t.Cleanup(func() {
			if t.Failed() {
				// dump diagnostic info to test logs
				_ = k8s.RunKubectlE(t, kubectlOptions, "describe", "pod", pod.Name)
			}
			// collect logs
			go GetAppLog(t, namespaceName, pod.Name, "", &corev1.PodLogOptions{Follow: true})

		})
	}

	AwaitNrReplicasReady(t, namespaceName, xmtpdDeploymentSync, replicaCount)
	AwaitNrReplicasReady(t, namespaceName, xmtpdDeploymentApi, replicaCount)
	AwaitNrReplicasReady(t, namespaceName, xmtpdDeploymentIndexer, replicaCount)
	AwaitNrReplicasReady(t, namespaceName, xmtpdDeploymentReporting, replicaCount)

	return
}
