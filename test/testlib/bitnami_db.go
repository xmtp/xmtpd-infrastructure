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

func installDB(t *testing.T, options *helm.Options, helmChartReleaseName string) {
	helm.Install(t, options, "oci://registry-1.docker.io/bitnamicharts/postgresql", helmChartReleaseName)

}

// StartDB
/**
 * StartDB starts a PostgreSQL database using the specified Helm options and namespace.
 *
 * @param t *testing.T - The testing context.
 * @param options *helm.Options - The Helm options for the installation.
 * @param namespace string - The namespace for the PostgreSQL pod.
 *
 * @return (string, string, DB) - Returns the Helm chart release name, namespace, and DB connection information.
 */
func StartDB(t *testing.T, options *helm.Options, namespace string) (string, string, DB) {
	return startDBTemplate(t, options, 1, namespace, "pg", installDB, true)
}

type DBInstallationStep func(t *testing.T, options *helm.Options, helmChartReleaseName string)

type DB struct {
	Password   string
	ConnString string
}

func startDBTemplate(t *testing.T, options *helm.Options, replicaCount int, namespace string, releaseName string, installStep DBInstallationStep, awaitRunning bool) (helmChartReleaseName string, namespaceName string, db DB) {
	randomSuffix := strings.ToLower(random.UniqueId())

	helmChartReleaseName = releaseName
	if helmChartReleaseName == "" {
		helmChartReleaseName = fmt.Sprintf("pg-%s", randomSuffix)
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

	AddTeardown(TEARDOWN_DATABASE, func() {
		helm.Delete(t, options, helmChartReleaseName, true)
	})

	if !awaitRunning {
		return
	}

	dbStatefulSet := fmt.Sprintf("%s-%s", helmChartReleaseName, "postgresql")

	defer func() {
		// collect some useful diagnostics
		if t.Failed() {
			// ignore any errors. This is already failed
			_ = k8s.RunKubectlE(t, kubectlOptions, "describe", "statefulset", dbStatefulSet)
		}
	}()

	AwaitNrReplicasScheduled(t, namespaceName, dbStatefulSet, replicaCount)

	pods := FindPodsFromChart(t, namespaceName, dbStatefulSet)

	for _, pod := range pods {
		AddTeardown(TEARDOWN_DATABASE, func() {
			if t.Failed() {
				// dump diagnostic info to test logs
				_ = k8s.RunKubectlE(t, kubectlOptions, "describe", "pod", pod.Name)
			}
			// collect logs
			go GetAppLog(t, namespaceName, pod.Name, "", &corev1.PodLogOptions{Follow: true})

		})
	}

	pwd := getPGPwd(t, namespaceName, fmt.Sprintf("%s-postgresql", helmChartReleaseName))

	AwaitNrReplicasReady(t, namespaceName, dbStatefulSet, replicaCount)

	db = DB{
		Password:   pwd,
		ConnString: fmt.Sprintf("postgres://postgres:%s@%s.%s.svc.cluster.local:5432/postgres?sslmode=disable", pwd, dbStatefulSet, namespaceName),
	}

	return
}

func getPGPwd(t *testing.T, namespace string, secretName string) string {
	kubectlOptions := k8s.NewKubectlOptions("", "", namespace)

	secret := k8s.GetSecret(t, kubectlOptions, secretName)
	if secret == nil {
		t.Fatalf("secret %s not found", secretName)
	}
	// Extract the password from the secret's data
	password, ok := secret.Data["postgres-password"]
	if !ok {
		t.Fatalf("password key not found in secret %s", secretName)
	}

	return string(password)
}
