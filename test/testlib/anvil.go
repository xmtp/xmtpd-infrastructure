package testlib

import (
	"fmt"
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

func installAnvil(t *testing.T, options *k8s.KubectlOptions) {
	k8s.KubectlApply(t, options, ANVIL_DEPLOYMENT_FILE)
}

func deleteAnvil(t *testing.T, options *k8s.KubectlOptions) {
	k8s.KubectlDelete(t, options, ANVIL_DEPLOYMENT_FILE)
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
func StartAnvil(t *testing.T, options *helm.Options, namespace string) (string, string, AnvilCfg) {
	return StartAnvilTemplate(t, options, namespace, installAnvil, true)
}

type AnvilCfg struct {
	Endpoint string
}

type AnvilInstallationStep func(t *testing.T, options *k8s.KubectlOptions)

func StartAnvilTemplate(t *testing.T, options *helm.Options, namespace string, installStep AnvilInstallationStep, awaitRunning bool) (string, string, AnvilCfg) {

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

	anvil := AnvilCfg{
		Endpoint: fmt.Sprintf("ws://%s.%s.svc.cluster.local:8545", "anvil-service", namespaceName),
	}

	if !awaitRunning {
		return namespaceName, ANVIL_DEPLOYMENT_NAME, anvil
	}

	defer func() {
		// collect some useful diagnostics
		if t.Failed() {
			// ignore any errors. This is already failed
			_ = k8s.RunKubectlE(t, kubectlOptions, "describe", "deployment", ANVIL_DEPLOYMENT_NAME)
		}
	}()

	AwaitNrReplicasScheduled(t, namespaceName, ANVIL_DEPLOYMENT_NAME, 1)

	pods := FindPodsFromChart(t, namespaceName, ANVIL_DEPLOYMENT_NAME)

	for _, pod := range pods {
		t.Cleanup(func() {
			if t.Failed() {
				// dump diagnostic info to test logs
				_ = k8s.RunKubectlE(t, kubectlOptions, "describe", "pod", pod.Name)
			}
			// collect logs
			go GetAppLog(t, namespaceName, pod.Name, "", &corev1.PodLogOptions{Follow: true})

		})
	}

	AwaitNrReplicasReady(t, namespaceName, ANVIL_DEPLOYMENT_NAME, 1)

	return namespaceName, ANVIL_DEPLOYMENT_NAME, anvil
}
