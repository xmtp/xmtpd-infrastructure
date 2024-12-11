package xmtp_helm

import (
	"fmt"
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/xmtp/xmtpd-infrastructure/v1/test/testlib"
	v2 "k8s.io/api/core/v1"
	v1 "k8s.io/api/networking/v1"
	"strings"
	"testing"
)

func extractIngressE(t *testing.T, output string) *v1.Ingress {
	parts := strings.Split(output, "---")
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}

		if !strings.Contains(part, "kind: Ingress") {
			continue
		}

		var object v1.Ingress
		helm.UnmarshalK8SYaml(t, part, &object)

		return &object
	}

	return nil
}

func extractIngress(t *testing.T, output string) *v1.Ingress {
	ingress := extractIngressE(t, output)

	if ingress == nil {
		t.Fatalf("Could not extract ingress from template")
	}

	return ingress
}

func extractNamedSecretE(t *testing.T, output string, secretName string) *v2.Secret {
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

		var object v2.Secret
		helm.UnmarshalK8SYaml(t, part, &object)

		return &object
	}

	return nil
}

func TestXmtpdEmpty(t *testing.T) {
	options := &helm.Options{
		SetValues: map[string]string{},
	}
	output := helm.RenderTemplate(t, options, testlib.XMTPD_HELM_CHART_PATH, "release-name", []string{})

	ingress := extractIngressE(t, output)
	assert.Nil(t, ingress)

}

func TestXmtpdEnableIngress(t *testing.T) {

	options := &helm.Options{
		SetValues: map[string]string{
			"ingress.create": "true",
		},
	}

	output := helm.RenderTemplate(t, options, testlib.XMTPD_HELM_CHART_PATH, "release-name", []string{})

	ingress := extractIngress(t, output)
	assert.Contains(t, ingress.Annotations, "kubernetes.io/ingress.class")
	assert.Equal(t, "nginx", *ingress.Spec.IngressClassName)
}

func TestXmtpdIngressTLSNoSecret(t *testing.T) {

	options := &helm.Options{
		SetValues: map[string]string{
			"ingress.create":         "true",
			"ingress.tls.certIssuer": "cert-manager",
		},
	}

	output := helm.RenderTemplate(t, options, testlib.XMTPD_HELM_CHART_PATH, "release-name", []string{})

	ingress := extractIngress(t, output)
	assert.Contains(t, ingress.Annotations, "cert-manager.io/cluster-issuer")
	assert.Equal(t, "cert-manager", ingress.Annotations["cert-manager.io/cluster-issuer"])
	assert.Empty(t, ingress.Spec.TLS)
}

func TestXmtpdIngressTLSSecretNoCreate(t *testing.T) {

	options := &helm.Options{
		SetValues: map[string]string{
			"ingress.create":         "true",
			"ingress.tls.certIssuer": "cert-manager",
			"ingress.tls.secretName": "my-secret",
			"ingress.host":           "my-host",
		},
	}

	output := helm.RenderTemplate(t, options, testlib.XMTPD_HELM_CHART_PATH, "release-name", []string{})

	ingress := extractIngress(t, output)
	assert.Contains(t, ingress.Annotations, "cert-manager.io/cluster-issuer")
	assert.Equal(t, "cert-manager", ingress.Annotations["cert-manager.io/cluster-issuer"])

	expectedTLS := v1.IngressTLS{
		Hosts:      []string{"my-host"},
		SecretName: "my-secret",
	}
	assert.Contains(t, ingress.Spec.TLS, expectedTLS)

	secret := extractNamedSecretE(t, output, "my-secret")
	assert.Nil(t, secret)
}
