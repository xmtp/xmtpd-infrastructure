package testlib

import (
	"fmt"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	corev1 "k8s.io/api/core/v1"
)

func installAnvil(t *testing.T, options *k8s.KubectlOptions) {
	k8s.KubectlApply(t, options, AnvilDeploymentFile)
}

func deleteAnvil(t *testing.T, options *k8s.KubectlOptions) {
	k8s.KubectlDelete(t, options, AnvilDeploymentFile)
}

// StartAnvil
/**
 * StartAnvil starts a blockchain AnvilCfg node
 *
 * @param t *testing.T - The testing context.
 * @param options *helm.Options - The Helm options for the installation.
 * @param namespace string - The namespace for the AnvilCfg node.
 *
 */
func StartAnvil(t *testing.T, options *helm.Options, namespace string) (string, string, *AnvilCfg) {
	return StartAnvilTemplate(t, options, namespace, installAnvil, true)
}

type AnvilCfg struct {
	WssEndpoint string
	RPCEndpoint string
}

type AnvilInstallationStep func(t *testing.T, options *k8s.KubectlOptions)

func StartAnvilTemplate(
	t *testing.T,
	options *helm.Options,
	namespace string,
	installStep AnvilInstallationStep,
	awaitRunning bool,
) (string, string, *AnvilCfg) {
	var namespaceName string
	if namespace == "" {
		namespaceName = CreateRandomNamespace(t, 4)
	} else {
		namespaceName = namespace
	}

	kubectlOptions := k8s.NewKubectlOptions("", "", namespaceName)
	options.KubectlOptions = kubectlOptions
	options.KubectlOptions.Namespace = namespaceName

	installStep(t, kubectlOptions)

	t.Cleanup(func() {
		deleteAnvil(t, kubectlOptions)
	})

	anvil := &AnvilCfg{
		WssEndpoint: fmt.Sprintf(
			"ws://%s.%s.svc.cluster.local:8545",
			"anvil-service",
			namespaceName,
		),
		RPCEndpoint: fmt.Sprintf(
			"http://%s.%s.svc.cluster.local:8545",
			"anvil-service",
			namespaceName,
		),
	}

	if !awaitRunning {
		return namespaceName, AnvilDeploymentName, anvil
	}

	defer func() {
		// collect some useful diagnostics
		if t.Failed() {
			// ignore any errors. This is already failed
			_ = k8s.RunKubectlE(t, kubectlOptions, "describe", "deployment", AnvilDeploymentName)
		}
	}()

	AwaitNrReplicasScheduled(t, namespaceName, AnvilDeploymentName, 1)

	pods := FindPodsFromChart(t, namespaceName, AnvilDeploymentName)

	for _, pod := range pods {
		p := pod // capture range variable
		t.Cleanup(func() {
			if t.Failed() {
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

	AwaitNrReplicasReady(t, namespaceName, AnvilDeploymentName, 1)

	return namespaceName, AnvilDeploymentName, anvil
}
