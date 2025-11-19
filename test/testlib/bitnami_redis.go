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

func installRedis(t *testing.T, options *helm.Options, helmChartReleaseName string) {
	helm.Install(t, options, "oci://registry-1.docker.io/bitnamicharts/redis", helmChartReleaseName)

}

// StartRedis
/**
 * StartRedis starts a Redis cluster using the specified Helm options and namespace.
 *
 * @param t *testing.T - The testing context.
 * @param options *helm.Options - The Helm options for the installation.
 * @param namespace string - The namespace for the Redis pod.
 *
 * @return (string, string, Redis) - Returns the Helm chart release name, namespace, and Redis connection information.
 */
func StartRedis(t *testing.T, options *helm.Options, namespace string) (string, string, Redis) {
	return startRedisTemplate(t, options, 1, namespace, "redis", installRedis, true)
}

type RedisInstallationStep func(t *testing.T, options *helm.Options, helmChartReleaseName string)

type Redis struct {
	ConnString string
}

func startRedisTemplate(t *testing.T, options *helm.Options, replicaCount int, namespace string, releaseName string, installStep RedisInstallationStep, awaitRunning bool) (string, string, Redis) {
	randomSuffix := strings.ToLower(random.UniqueId())

	helmChartReleaseName := releaseName
	if helmChartReleaseName == "" {
		helmChartReleaseName = fmt.Sprintf("redis-%s", randomSuffix)
	}

	var namespaceName string
	if namespace == "" {
		namespaceName = CreateRandomNamespace(t, 4)
	} else {
		namespaceName = namespace
	}

	kubectlOptions := k8s.NewKubectlOptions("", "", namespaceName)
	options.KubectlOptions = kubectlOptions
	options.KubectlOptions.Namespace = namespaceName

	installStep(t, options, helmChartReleaseName)

	t.Cleanup(func() {
		helm.Delete(t, options, helmChartReleaseName, true)
	})

	if !awaitRunning {
		return helmChartReleaseName, namespaceName, Redis{}
	}

	// bitnami redis cluster does not prepend the release name to the resource names
	redisStatefulSet := fmt.Sprintf("%s", "redis-master")

	defer func() {
		// collect some useful diagnostics
		if t.Failed() {
			// ignore any errors. This is already failed
			_ = k8s.RunKubectlE(t, kubectlOptions, "describe", "statefulset", redisStatefulSet)
		}
	}()

	AwaitNrReplicasScheduled(t, namespaceName, redisStatefulSet, replicaCount)

	pods := FindPodsFromChart(t, namespaceName, redisStatefulSet)

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

	AwaitNrReplicasReady(t, namespaceName, redisStatefulSet, replicaCount)

	redis := Redis{
		ConnString: fmt.Sprintf("redis://%s.%s.svc.cluster.local:6379", redisStatefulSet, namespaceName),
	}

	return helmChartReleaseName, namespaceName, redis
}
