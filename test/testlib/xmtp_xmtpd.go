package testlib

import (
	"fmt"
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/random"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	"os"
	"strings"
	"testing"
)

func installXMTPD(t *testing.T, options *helm.Options, helmChartReleaseName string) {
	if options.Version == "" {
		helm.Install(t, options, XMTPD_HELM_CHART_PATH, helmChartReleaseName)
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

	AddTeardown(TEARDOWN_XMTPD, func() {
		helm.Delete(t, options, helmChartReleaseName, true)
	})

	if !awaitRunning {
		return
	}

	xmtpdDeploymentSync := fmt.Sprintf("%s-sync", helmChartReleaseName)
	xmtpdDeploymentApi := fmt.Sprintf("%s-api", helmChartReleaseName)

	defer func() {
		// collect some useful diagnostics
		if t.Failed() {
			// ignore any errors. This is already failed
			_ = k8s.RunKubectlE(t, kubectlOptions, "describe", "deployment", xmtpdDeploymentSync)
			_ = k8s.RunKubectlE(t, kubectlOptions, "describe", "deployment", xmtpdDeploymentApi)
		}
	}()

	AwaitNrReplicasScheduled(t, namespaceName, xmtpdDeploymentSync, replicaCount)
	AwaitNrReplicasScheduled(t, namespaceName, xmtpdDeploymentApi, replicaCount)

	podsSync := FindPodsFromChart(t, namespaceName, xmtpdDeploymentSync)
	podsApi := FindPodsFromChart(t, namespaceName, xmtpdDeploymentApi)

	allPods := append(podsSync, podsApi...)
	for _, pod := range allPods {
		AddTeardown(TEARDOWN_XMTPD, func() {
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

	return
}

func getLastSection(envKey string) string {
	// Split the string by dots
	parts := strings.Split(envKey, ".")
	// Return the last part
	return parts[len(parts)-1]
}

// GetDefaultSecrets loads secrets in the following priority order:
// 1. Environmental variables.
// 2. Well-known file on disk (e.g., LOCAL_SECRETS_FILE).
// 3. Default values.
func GetDefaultSecrets(t *testing.T) map[string]string {
	defaultSecrets := map[string]string{
		"env.secret.XMTPD_SIGNER_PRIVATE_KEY": "<replace-me>",
		"env.secret.XMTPD_PAYER_PRIVATE_KEY":  "<replace-me>",
		"env.secret.XMTPD_CONTRACTS_RPC_URL":  "<replace-me>",
		"env.secret.XMTPD_LOG_LEVEL":          "debug",
	}

	// Load secrets from a well-known file
	secretsFromFile := loadSecretsFromYAMLFile(t, LOCAL_SECRETS_FILE)

	// Merge secrets with priority: environment variable > file > default
	mergedSecrets := make(map[string]string)
	for key, defaultValue := range defaultSecrets {
		if value, found := os.LookupEnv(getLastSection(key)); found {
			mergedSecrets[key] = value
		} else if value, found := secretsFromFile[key]; found {
			mergedSecrets[key] = value
		} else {
			mergedSecrets[key] = defaultValue
		}
	}

	return mergedSecrets
}

// loadSecretsFromYAMLFile loads secrets from a given YAML file.
// The YAML file should have a flat key-value structure.
// For Example:
// env.secret.XMTPD_SIGNER_PRIVATE_KEY: "custom_private_key"
// env.secret.XMTPD_LOG_LEVEL: "info"
// env.secret.XMTPD_CONTRACTS_RPC_URL: "https://xmtp-testnet.g.alchemy.com/..."
func loadSecretsFromYAMLFile(t *testing.T, filePath string) map[string]string {
	secrets := make(map[string]string)

	file, err := os.Open(filePath)
	if err != nil {
		t.Logf("Could not open file %s. Using default values. Error: %s", filePath, err.Error())
		return secrets
	}
	defer func() {
		_ = file.Close()
	}()

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&secrets)
	if err != nil {
		t.Logf("Could not decode file %s. Using default values. Error: %s", filePath, err.Error())
		return secrets
	}

	return secrets
}
