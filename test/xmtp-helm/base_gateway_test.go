package xmtp_helm

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"

	"github.com/xmtp/xmtpd-infrastructure/v1/test/testlib"
)

func TestKubernetesBasicGatewayInstall(t *testing.T) {
	namespace := testlib.CreateRandomNamespace(t, 2)

	options := helm.Options{
		SetValues: map[string]string{
			"auth.enabled": "false",
		},
	}

	dbCh := testlib.RunAsync(func() *testlib.DB {
		_, _, db := testlib.StartDB(t, &options, namespace)
		return db
	})
	anvilCh := testlib.RunAsync(func() *testlib.AnvilCfg {
		_, _, anvil := testlib.StartAnvil(t, &options, namespace)
		return anvil
	})

	redisCh := testlib.RunAsync(func() *testlib.Redis {
		_, _, redis := testlib.StartRedis(t, &options, namespace)
		return redis
	})

	db := <-dbCh
	anvil := <-anvilCh
	redis := <-redisCh

	secrets := testlib.GetDefaultSecrets(t)
	secrets["env.secret.XMTPD_DB_WRITER_CONNECTION_STRING"] = db.ConnString
	secrets["env.secret.XMTPD_REDIS_URL"] = redis.ConnString
	secrets["env.secret.XMTPD_SETTLEMENT_CHAIN_WSS_URL"] = anvil.WssEndpoint
	secrets["env.secret.XMTPD_APP_CHAIN_WSS_URL"] = anvil.WssEndpoint
	secrets["env.secret.XMTPD_SETTLEMENT_CHAIN_RPC_URL"] = anvil.RPCEndpoint
	secrets["env.secret.XMTPD_APP_CHAIN_RPC_URL"] = anvil.RPCEndpoint

	options = helm.Options{
		SetValues: secrets,
	}
	testlib.StartGateway(t, &options, 1, namespace)
}
