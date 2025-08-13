package xmtp_helm

import (
	"github.com/gruntwork-io/terratest/modules/helm"
	"testing"

	"github.com/xmtp/xmtpd-infrastructure/v1/test/testlib"
)

func TestKubernetesBasicPayerInstall(t *testing.T) {
	t.Skip("Replace this with new gateway")
	namespace := testlib.CreateRandomNamespace(t, 2)

	options := helm.Options{
		SetValues: map[string]string{},
	}

	dbCh := testlib.RunAsync(func() testlib.DB {
		_, _, db := testlib.StartDB(t, &options, namespace)
		return db
	})
	anvilCh := testlib.RunAsync(func() testlib.AnvilCfg {
		_, _, anvil := testlib.StartAnvil(t, &options, namespace)
		return anvil
	})

	db := <-dbCh
	anvil := <-anvilCh

	secrets := testlib.GetDefaultSecrets(t)
	secrets["env.secret.XMTPD_DB_WRITER_CONNECTION_STRING"] = db.ConnString
	secrets["env.secret.XMTPD_SETTLEMENT_CHAIN_WSS_URL"] = anvil.WssEndpoint
	secrets["env.secret.XMTPD_APP_CHAIN_WSS_URL"] = anvil.WssEndpoint
	secrets["env.secret.XMTPD_SETTLEMENT_CHAIN_RPC_URL"] = anvil.RPCEndpoint
	secrets["env.secret.XMTPD_APP_CHAIN_RPC_URL"] = anvil.RPCEndpoint

	options = helm.Options{
		SetValues: secrets,
	}
	testlib.StartPayer(t, &options, 1, namespace)
}
