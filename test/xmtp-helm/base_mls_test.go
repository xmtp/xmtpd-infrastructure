package xmtp_helm

import (
	"github.com/gruntwork-io/terratest/modules/helm"
	"testing"

	"github.com/xmtp/xmtpd-infrastructure/v1/test/testlib"
)

func TestKubernetesBasicMLSInstall(t *testing.T) {
	options := helm.Options{
		SetValues: map[string]string{},
	}

	testlib.StartMLS(t, &options, 1, "")
}
