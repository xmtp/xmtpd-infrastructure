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

	// technically the payer does NOT require a DB
	// but XMTPD 0.1.1 has some incorrect defaults
	// which will prevent the start of the service without DB connectivity
	defer testlib.Teardown(testlib.TEARDOWN_DATABASE)
	_, _, db := testlib.StartDB(t, &options, namespace)

	secrets := testlib.GetDefaultSecrets(t)
	secrets["env.secret.XMTPD_DB_WRITER_CONNECTION_STRING"] = db.ConnString

	options = helm.Options{
		SetValues: secrets,
	}

	defer testlib.Teardown(testlib.TEARDOWN_PAYER)
	testlib.StartPayer(t, &options, 2, namespace)
}
