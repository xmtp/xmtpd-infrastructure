package xmtp_helm

import (
	"github.com/gruntwork-io/terratest/modules/helm"
	"testing"

	"github.com/xmtp/xmtpd-infrastructure/v1/test/testlib"
)

func TestKubernetesBasicMLSInstall(t *testing.T) {
	defer testlib.VerifyTeardown(t)
	defer testlib.Teardown(testlib.TEARDOWN_GLOBAL)

	options := helm.Options{
		SetValues: map[string]string{},
	}

	defer testlib.Teardown(testlib.TEARDOWN_MLS)

	testlib.StartMLS(t, &options, 1, "")
}
