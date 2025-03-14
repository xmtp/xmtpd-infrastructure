package xmtp_helm

import (
	"github.com/gruntwork-io/terratest/modules/helm"
	"testing"

	"github.com/xmtp/xmtpd-infrastructure/v1/test/testlib"
)

func TestKubernetesBasicPayerInstall(t *testing.T) {
	defer testlib.VerifyTeardown(t)
	defer testlib.Teardown(testlib.TEARDOWN_GLOBAL)

	namespace := testlib.CreateRandomNamespace(t, 2)

	options := helm.Options{
		SetValues: map[string]string{},
	}

	secrets := testlib.GetDefaultSecrets(t)

	options = helm.Options{
		SetValues: secrets,
	}

	defer testlib.Teardown(testlib.TEARDOWN_PAYER)
	testlib.StartPayer(t, &options, 1, namespace)
}
