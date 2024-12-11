package xmtp_helm

import (
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/xmtp/xmtpd-infrastructure/v1/test/testlib"
	v1 "k8s.io/api/networking/v1"
	"testing"
)

func TestPayerEmpty(t *testing.T) {
	options := &helm.Options{
		SetValues: map[string]string{},
	}
	output := helm.RenderTemplate(t, options, testlib.XMTP_PAYER_HELM_CHART_PATH, "release-name", []string{})

	ingress := extractIngressE(t, output)
	assert.Nil(t, ingress)

}

func TestPayerEnableIngress(t *testing.T) {

	options := &helm.Options{
		SetValues: map[string]string{
			"ingress.create": "true",
		},
	}

	output := helm.RenderTemplate(t, options, testlib.XMTP_PAYER_HELM_CHART_PATH, "release-name", []string{})

	ingress := extractIngress(t, output)
	assert.Contains(t, ingress.Annotations, "kubernetes.io/ingress.class")
	assert.Equal(t, "nginx", *ingress.Spec.IngressClassName)
}

func TestPayerIngressTLSNoSecret(t *testing.T) {

	options := &helm.Options{
		SetValues: map[string]string{
			"ingress.create":         "true",
			"ingress.tls.certIssuer": "cert-manager",
		},
	}

	output := helm.RenderTemplate(t, options, testlib.XMTP_PAYER_HELM_CHART_PATH, "release-name", []string{})

	ingress := extractIngress(t, output)
	assert.Contains(t, ingress.Annotations, "cert-manager.io/cluster-issuer")
	assert.Equal(t, "cert-manager", ingress.Annotations["cert-manager.io/cluster-issuer"])
	assert.Empty(t, ingress.Spec.TLS)
}

func TestPayerIngressTLSSecretNoCreate(t *testing.T) {

	options := &helm.Options{
		SetValues: map[string]string{
			"ingress.create":         "true",
			"ingress.tls.certIssuer": "cert-manager",
			"ingress.tls.secretName": "my-secret",
			"ingress.host":           "my-host",
		},
	}

	output := helm.RenderTemplate(t, options, testlib.XMTP_PAYER_HELM_CHART_PATH, "release-name", []string{})

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
