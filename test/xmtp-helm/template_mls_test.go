package xmtp_helm

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/xmtp/xmtpd-infrastructure/v1/test/testlib"
)

func TestMLSNoEnvWorks(t *testing.T) {
	options := &helm.Options{}

	output := helm.RenderTemplate(
		t,
		options,
		testlib.MlsHelmChartPath,
		"release-name",
		[]string{},
	)
	deployment := testlib.ExtractDeployment(t, output, "release-name-mls-validation-service")

	assert.NotNil(t, deployment)
	assert.Empty(t, deployment.Spec.Template.Spec.Containers[0].Env)
}

func TestMLSEnvWorks(t *testing.T) {
	options := &helm.Options{
		SetValues: map[string]string{
			"env.CHAIN_RPC_10": "https://optimism.madeuprpc.com",
		},
	}

	output := helm.RenderTemplate(
		t,
		options,
		testlib.MlsHelmChartPath,
		"release-name",
		[]string{},
	)
	deployment := testlib.ExtractDeployment(t, output, "release-name-mls-validation-service")

	assert.NotNil(t, deployment)
	envVars := deployment.Spec.Template.Spec.Containers[0].Env
	assert.NotEmpty(t, envVars)
	assert.Equal(t, 1, len(envVars))
	assert.Equal(t, "CHAIN_RPC_10", envVars[0].Name)
	assert.Equal(t, "https://optimism.madeuprpc.com", envVars[0].Value)
}
