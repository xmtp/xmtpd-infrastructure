package xmtp_helm

import (
	"github.com/gruntwork-io/terratest/modules/helm"
	"testing"

	"github.com/xmtp/xmtpd-infrastructure/v1/test/testlib"
)

func TestKubernetesBasicDependenciesInstall(t *testing.T) {
	options := helm.Options{
		SetValues: map[string]string{},
	}

	testlib.StartDB(t, &options, "")
}

func TestKubernetesAnvil(t *testing.T) {
	options := helm.Options{
		SetValues: map[string]string{},
	}
	testlib.StartAnvil(t, &options, "")
}

func TestKubernetesRedis(t *testing.T) {
	options := helm.Options{
		SetValues: map[string]string{
			"auth.enabled": "false",
		},
	}
	testlib.StartRedis(t, &options, "")
}
