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

func installMLS(t *testing.T, options *helm.Options, helmChartReleaseName string) {
	if options.Version == "" {
		helm.Install(t, options, MLS_HELM_CHART_PATH, helmChartReleaseName)
	} else {
		helm.Install(t, options, "xmtp/mls-validation-service ", helmChartReleaseName)
	}
}

// StartMLS
/**
 * StartMLS starts a MLS Validation Service using the specified Helm options and namespace.
 *
 * @param t *testing.T - The testing context.
 * @param options *helm.Options - The Helm options for the installation.
 * @param namespace string - The namespace for the MLS Validation Service.
 *
 * @return (string, string, MLS) - Returns the Helm chart release name, namespace, and MLS connection information.
 */
func StartMLS(t *testing.T, options *helm.Options, replicaCount int, namespace string) (string, string, MLS) {
	return StartMLSTemplate(t, options, replicaCount, namespace, "", installMLS, true)
}

type MLS struct {
	Endpoint string
}

type MLSInstallationStep func(t *testing.T, options *helm.Options, helmChartReleaseName string)

func StartMLSTemplate(t *testing.T, options *helm.Options, replicaCount int, namespace string, releaseName string, installStep MLSInstallationStep, awaitRunning bool) (helmChartReleaseName string, namespaceName string, mls MLS) {
	randomSuffix := strings.ToLower(random.UniqueId())

	helmChartReleaseName = releaseName
	if helmChartReleaseName == "" {
		helmChartReleaseName = fmt.Sprintf("mls-%s", randomSuffix)
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

	AddTeardown(TEARDOWN_MLS, func() {
		helm.Delete(t, options, helmChartReleaseName, true)
	})

	if !awaitRunning {
		return
	}

	mlsDeployment := fmt.Sprintf("%s-%s", helmChartReleaseName, "mls-validation-service")

	defer func() {
		// collect some useful diagnostics
		if t.Failed() {
			// ignore any errors. This is already failed
			_ = k8s.RunKubectlE(t, kubectlOptions, "describe", "deployment", mlsDeployment)
		}
	}()

	AwaitNrReplicasScheduled(t, namespaceName, mlsDeployment, replicaCount)

	pods := FindPodsFromChart(t, namespaceName, mlsDeployment)

	for _, pod := range pods {
		AddTeardown(TEARDOWN_MLS, func() {
			if t.Failed() {
				// dump diagnostic info to test logs
				_ = k8s.RunKubectlE(t, kubectlOptions, "describe", "pod", pod.Name)
			}
			// collect logs
			go GetAppLog(t, namespaceName, pod.Name, "", &corev1.PodLogOptions{Follow: true})

		})
	}

	mls = MLS{
		Endpoint: fmt.Sprintf("http://%s.%s.svc.cluster.local:50051", mlsDeployment, namespaceName),
	}

	AwaitNrReplicasReady(t, namespaceName, mlsDeployment, replicaCount)

	return
}
