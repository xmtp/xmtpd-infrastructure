package xmtp_helm

import (
	"github.com/gruntwork-io/terratest/modules/helm"
	"testing"

	"github.com/xmtp/xmtpd-infrastructure/v1/test/testlib"
)

func TestKubernetesBasicPayerInstall(t *testing.T) {
	defer testlib.VerifyTeardown(t)
	defer testlib.Teardown(testlib.TEARDOWN_GLOBAL)

	options := helm.Options{
		SetValues: testlib.GetDefaultSecrets(t),
	}

	defer testlib.Teardown(testlib.TEARDOWN_PAYER)
	testlib.StartPayer(t, &options, 2, "")
}
