package xmtp_helm

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/xmtp/xmtpd-infrastructure/v1/test/testlib"
	netv1 "k8s.io/api/networking/v1"
)

func TestGatewayEmpty(t *testing.T) {
	options := &helm.Options{
		SetValues: map[string]string{},
	}
	output := helm.RenderTemplate(
		t,
		options,
		testlib.XmtpGatewayHelmChartPath,
		"release-name",
		[]string{},
	)

	ingress := testlib.ExtractIngressE(t, output)
	assert.Nil(t, ingress)
}

func TestGatewayEnableIngress(t *testing.T) {
	options := &helm.Options{
		SetValues: map[string]string{
			"ingress.create": "true",
		},
	}

	output := helm.RenderTemplate(
		t,
		options,
		testlib.XmtpGatewayHelmChartPath,
		"release-name",
		[]string{},
	)

	ingress := testlib.ExtractIngress(t, output)
	assert.Contains(t, ingress.Annotations, "kubernetes.io/ingress.class")
	assert.Equal(t, "nginx", *ingress.Spec.IngressClassName)
}

func TestGatewayIngressTLSNoSecret(t *testing.T) {
	options := &helm.Options{
		SetValues: map[string]string{
			"ingress.create":         "true",
			"ingress.tls.certIssuer": "cert-manager",
		},
	}

	output := helm.RenderTemplate(
		t,
		options,
		testlib.XmtpGatewayHelmChartPath,
		"release-name",
		[]string{},
	)

	ingress := testlib.ExtractIngress(t, output)
	assert.Contains(t, ingress.Annotations, "cert-manager.io/cluster-issuer")
	assert.Equal(t, "cert-manager", ingress.Annotations["cert-manager.io/cluster-issuer"])
	assert.Empty(t, ingress.Spec.TLS)
}

func TestGatewayIngressTLSSecretNoCreate(t *testing.T) {
	options := &helm.Options{
		SetValues: map[string]string{
			"ingress.create":         "true",
			"ingress.tls.certIssuer": "cert-manager",
			"ingress.tls.secretName": "my-secret",
			"ingress.host":           "my-host",
		},
	}

	output := helm.RenderTemplate(
		t,
		options,
		testlib.XmtpGatewayHelmChartPath,
		"release-name",
		[]string{},
	)

	ingress := testlib.ExtractIngress(t, output)
	assert.Contains(t, ingress.Annotations, "cert-manager.io/cluster-issuer")
	assert.Equal(t, "cert-manager", ingress.Annotations["cert-manager.io/cluster-issuer"])

	expectedTLS := netv1.IngressTLS{
		Hosts:      []string{"my-host"},
		SecretName: "my-secret",
	}
	assert.Contains(t, ingress.Spec.TLS, expectedTLS)

	secret := testlib.ExtractNamedSecretE(t, output, "my-secret")
	assert.Nil(t, secret)
}
