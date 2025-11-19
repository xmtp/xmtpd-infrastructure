package xmtp_helm

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"

	"github.com/xmtp/xmtpd-infrastructure/v1/test/testlib"
)

func TestKubernetesBasicXMTPDInstall(t *testing.T) {
	namespace := testlib.CreateRandomNamespace(t, 2)

	options := helm.Options{}

	dbCh := testlib.RunAsync(func() *testlib.DB {
		_, _, db := testlib.StartDB(t, &options, namespace)
		return db
	})

	mlsCh := testlib.RunAsync(func() *testlib.MLS {
		_, _, mls := testlib.StartMLS(t, &options, 1, namespace)
		return mls
	})

	anvilCh := testlib.RunAsync(func() *testlib.AnvilCfg {
		_, _, anvil := testlib.StartAnvil(t, &options, namespace)
		return anvil
	})

	db := <-dbCh
	mls := <-mlsCh
	anvil := <-anvilCh

	secrets := testlib.GetDefaultSecrets(t)
	secrets["env.secret.XMTPD_DB_WRITER_CONNECTION_STRING"] = db.ConnString
	secrets["env.secret.XMTPD_MLS_VALIDATION_GRPC_ADDRESS"] = mls.Endpoint
	secrets["env.secret.XMTPD_SETTLEMENT_CHAIN_WSS_URL"] = anvil.WssEndpoint
	secrets["env.secret.XMTPD_APP_CHAIN_WSS_URL"] = anvil.WssEndpoint
	secrets["env.secret.XMTPD_SETTLEMENT_CHAIN_RPC_URL"] = anvil.RPCEndpoint
	secrets["env.secret.XMTPD_APP_CHAIN_RPC_URL"] = anvil.RPCEndpoint

	// TODO(mkysel) remove after 1.0.0 tag exists
	secrets["image.tag"] = "sha-16c3dd2"

	options = helm.Options{
		SetValues: secrets,
	}

	testlib.StartXMTPD(t, &options, 1, namespace)
}

func TestKubernetesXMTPDCronJob(t *testing.T) {
	namespace := testlib.CreateRandomNamespace(t, 2)

	options := helm.Options{}

	dbCh := testlib.RunAsync(func() *testlib.DB {
		_, _, db := testlib.StartDB(t, &options, namespace)
		return db
	})

	mlsCh := testlib.RunAsync(func() *testlib.MLS {
		_, _, mls := testlib.StartMLS(t, &options, 1, namespace)
		return mls
	})

	anvilCh := testlib.RunAsync(func() *testlib.AnvilCfg {
		_, _, anvil := testlib.StartAnvil(t, &options, namespace)
		return anvil
	})

	db := <-dbCh
	mls := <-mlsCh
	anvil := <-anvilCh

	secrets := testlib.GetDefaultSecrets(t)
	secrets["env.secret.XMTPD_DB_WRITER_CONNECTION_STRING"] = db.ConnString
	secrets["env.secret.XMTPD_MLS_VALIDATION_GRPC_ADDRESS"] = mls.Endpoint
	secrets["env.secret.XMTPD_SETTLEMENT_CHAIN_WSS_URL"] = anvil.WssEndpoint
	secrets["env.secret.XMTPD_APP_CHAIN_WSS_URL"] = anvil.WssEndpoint
	secrets["env.secret.XMTPD_SETTLEMENT_CHAIN_RPC_URL"] = anvil.RPCEndpoint
	secrets["env.secret.XMTPD_APP_CHAIN_RPC_URL"] = anvil.RPCEndpoint
	secrets["prune.create"] = "true"

	// TODO(mkysel) remove after 1.0.0 tag exists
	secrets["image.tag"] = "sha-16c3dd2"
	secrets["prune.image.tag"] = "sha-16c3dd2"

	options = helm.Options{
		SetValues: secrets,
	}

	releaseName, _, _ := testlib.StartXMTPD(t, &options, 1, namespace)
	cronJobExpectedName := releaseName + "-prune"

	jobs := testlib.FindCronJobsFromChart(t, namespace, cronJobExpectedName)
	require.Len(t, jobs, 1)

	manualJobName := "manualjob"

	// the schedule can be anything, we need to trigger it manually at least once
	testlib.CreateJobFromCronJob(t, namespace, &jobs[0], manualJobName)
	testlib.AwaitNrReplicasCreated(t, namespace, manualJobName, 1)
	jobPods := testlib.FindPodsFromChart(t, namespace, manualJobName)
	require.Len(t, jobPods, 1)

	jobPod := jobPods[0]

	testlib.AwaitPodTerminated(t, namespace, jobPod.Name)

	testlib.GetTerminatedPodLog(t, namespace, &jobPod, "", &corev1.PodLogOptions{})

	// we do not explicitly validate that the job has done its job
}
