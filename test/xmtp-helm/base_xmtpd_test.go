package xmtp_helm

import (
	"github.com/gruntwork-io/terratest/modules/helm"
	"testing"

	"github.com/xmtp/xmtpd-infrastructure/v1/test/testlib"
)

func TestKubernetesBasicXMTPDInstall(t *testing.T) {
	defer testlib.VerifyTeardown(t)
	defer testlib.Teardown(testlib.TEARDOWN_GLOBAL)

	namespace := testlib.CreateRandomNamespace(t, 2)

	options := helm.Options{
		SetValues: map[string]string{},
	}

	defer testlib.Teardown(testlib.TEARDOWN_DATABASE)
	_, _, db := testlib.StartDB(t, &options, namespace)

	defer testlib.Teardown(testlib.TEARDOWN_MLS)
	_, _, mls := testlib.StartMLS(t, &options, 1, namespace)

	secrets := testlib.GetDefaultSecrets(t)
	secrets["env.secret.XMTPD_DB_WRITER_CONNECTION_STRING"] = db.ConnString
	secrets["env.secret.XMTPD_MLS_VALIDATION_GRPC_ADDRESS"] = mls.Endpoint

	options = helm.Options{
		SetValues: secrets,
	}

	defer testlib.Teardown(testlib.TEARDOWN_XMTPD)
	testlib.StartXMTPD(t, &options, 1, namespace)
}
