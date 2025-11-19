package testlib

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	corev1 "k8s.io/api/core/v1"
)

func installMLS(t *testing.T, options *helm.Options, helmChartReleaseName string) {
	if options.Version == "" {
		helm.Install(t, options, MlsHelmChartPath, helmChartReleaseName)
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
func StartMLS(
	t *testing.T,
	options *helm.Options,
	replicaCount int,
	namespace string,
) (string, string, *MLS) {
	return StartMLSTemplate(t, options, replicaCount, namespace, "", installMLS, true)
}

type MLS struct {
	Endpoint string
}

type MLSInstallationStep func(t *testing.T, options *helm.Options, helmChartReleaseName string)

func StartMLSTemplate(
	t *testing.T,
	options *helm.Options,
	replicaCount int,
	namespace string,
	releaseName string,
	installStep MLSInstallationStep,
	awaitRunning bool,
) (string, string, *MLS) {
	randomSuffix := strings.ToLower(random.UniqueId())

	helmChartReleaseName := releaseName
	if helmChartReleaseName == "" {
		helmChartReleaseName = fmt.Sprintf("mls-%s", randomSuffix)
	}

	namespaceName := namespace
	if namespace == "" {
		namespaceName = CreateRandomNamespace(t, 4)
	}

	kubectlOptions := k8s.NewKubectlOptions("", "", namespaceName)
	options.KubectlOptions = kubectlOptions
	options.KubectlOptions.Namespace = namespaceName

	installStep(t, options, helmChartReleaseName)

	t.Cleanup(func() {
		helm.Delete(t, options, helmChartReleaseName, true)
	})

	mlsDeployment := fmt.Sprintf("%s-%s", helmChartReleaseName, "mls-validation-service")

	mls := &MLS{
		Endpoint: fmt.Sprintf("http://%s.%s.svc.cluster.local:50051", mlsDeployment, namespaceName),
	}

	if !awaitRunning {
		return helmChartReleaseName, namespaceName, mls
	}

	defer func() {
		// collect some useful diagnostics
		if t.Failed() {
			// ignore any errors. This is already failed
			_ = k8s.RunKubectlE(t, kubectlOptions, "describe", "deployment", mlsDeployment)
		}
	}()

	AwaitNrReplicasScheduled(t, namespaceName, mlsDeployment, replicaCount)

	pods := FindPodsFromChart(t, namespaceName, mlsDeployment)

	for _, p := range pods {
		t.Cleanup(func() {
			if t.Failed() {
				// dump diagnostic info to test logs
				_ = k8s.RunKubectlE(t, kubectlOptions, "describe", "pod", p.Name)
			}
			for _, c := range p.Spec.Containers {
				cName := c.Name
				go GetAppLog(
					t,
					namespaceName,
					p.Name,
					cName,
					&corev1.PodLogOptions{Follow: true, Container: cName},
				)
			}

			for _, c := range p.Spec.InitContainers {
				cName := c.Name
				go GetAppLog(
					t,
					namespaceName,
					p.Name,
					cName,
					&corev1.PodLogOptions{Follow: true, Container: cName},
				)
			}
		})
	}

	AwaitNrReplicasReady(t, namespaceName, mlsDeployment, replicaCount)

	return helmChartReleaseName, namespaceName, mls
}
