package testlib

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
)

func ExtractIngressE(t *testing.T, output string) *netv1.Ingress {
	parts := strings.Split(output, "---")
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}

		if !strings.Contains(part, "kind: Ingress") {
			continue
		}

		var object netv1.Ingress
		helm.UnmarshalK8SYaml(t, part, &object)

		return &object
	}

	return nil
}

func ExtractIngress(t *testing.T, output string) *netv1.Ingress {
	ingress := ExtractIngressE(t, output)

	if ingress == nil {
		t.Fatalf("Could not extract ingress from template")
	}

	return ingress
}

func ExtractNamedSecretE(t *testing.T, output string, secretName string) *corev1.Secret {
	parts := strings.Split(output, "---")
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}

		if !strings.Contains(part, "kind: Secret") {
			continue
		}

		if !strings.Contains(part, fmt.Sprintf("name: %s", secretName)) {
			continue
		}

		var object corev1.Secret
		helm.UnmarshalK8SYaml(t, part, &object)

		return &object
	}

	return nil
}

func ExtractDeploymentE(t *testing.T, output string, deploymentName string) *appsv1.Deployment {
	parts := strings.Split(output, "---")
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}

		if !strings.Contains(part, "kind: Deployment") {
			continue
		}

		if !strings.Contains(part, fmt.Sprintf("name: %s", deploymentName)) {
			continue
		}

		var object appsv1.Deployment
		helm.UnmarshalK8SYaml(t, part, &object)

		return &object
	}

	return nil
}

func ExtractDeployment(t *testing.T, output string, deploymentName string) *appsv1.Deployment {
	deployment := ExtractDeploymentE(t, output, deploymentName)

	if deployment == nil {
		t.Fatalf("Could not extract deployment from template")
	}

	return deployment
}

func ExtractCronJobE(t *testing.T, output string, cronjobName string) *v1.CronJob {
	parts := strings.Split(output, "---")
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}

		if !strings.Contains(part, "kind: CronJob") {
			continue
		}

		if !strings.Contains(part, fmt.Sprintf("name: %s", cronjobName)) {
			continue
		}

		var object v1.CronJob
		helm.UnmarshalK8SYaml(t, part, &object)

		return &object
	}

	return nil
}

func ExtractCronjob(t *testing.T, output string, cronjobName string) *v1.CronJob {
	deployment := ExtractCronJobE(t, output, cronjobName)

	if deployment == nil {
		t.Fatalf("Could not extract cronjob from template")
	}

	return deployment
}
