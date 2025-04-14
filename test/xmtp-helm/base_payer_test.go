package xmtp_helm

import (
	"github.com/gruntwork-io/terratest/modules/helm"
	"testing"

	"github.com/xmtp/xmtpd-infrastructure/v1/test/testlib"
)

func TestKubernetesBasicPayerInstall(t *testing.T) {
	namespace := testlib.CreateRandomNamespace(t, 2)

	options := helm.Options{
		SetValues: map[string]string{},
	}
	_, _, db := testlib.StartDB(t, &options, namespace)
	_, _, anvil := testlib.StartAnvil(t, &options, namespace)

	secrets := testlib.GetDefaultSecrets(t)
	secrets["env.secret.XMTPD_DB_WRITER_CONNECTION_STRING"] = db.ConnString
	secrets["env.secret.XMTPD_CONTRACTS_RPC_URL"] = anvil.Endpoint

	options = helm.Options{
		SetValues: secrets,
	}
	testlib.StartPayer(t, &options, 1, namespace)
}
