package xmtp_helm

import (
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/assert"
	"github.com/xmtp/xmtpd-infrastructure/v1/test/testlib"
	v1 "k8s.io/api/batch/v1"
	netv1 "k8s.io/api/networking/v1"
	"testing"
)

func TestXmtpdEmpty(t *testing.T) {
	options := &helm.Options{
		SetValues: map[string]string{},
	}
	output := helm.RenderTemplate(t, options, testlib.XMTPD_HELM_CHART_PATH, "release-name", []string{})

	ingress := testlib.ExtractIngressE(t, output)
	assert.Nil(t, ingress)

}

func TestXmtpdEnableIngress(t *testing.T) {

	options := &helm.Options{
		SetValues: map[string]string{
			"ingress.create": "true",
		},
	}

	output := helm.RenderTemplate(t, options, testlib.XMTPD_HELM_CHART_PATH, "release-name", []string{})

	ingress := testlib.ExtractIngress(t, output)
	assert.Contains(t, ingress.Annotations, "kubernetes.io/ingress.class")
	assert.Equal(t, "nginx", *ingress.Spec.IngressClassName)
}

func TestXmtpdIngressTLSNoSecret(t *testing.T) {

	options := &helm.Options{
		SetValues: map[string]string{
			"ingress.create":         "true",
			"ingress.tls.certIssuer": "cert-manager",
		},
	}

	output := helm.RenderTemplate(t, options, testlib.XMTPD_HELM_CHART_PATH, "release-name", []string{})

	ingress := testlib.ExtractIngress(t, output)
	assert.Contains(t, ingress.Annotations, "cert-manager.io/cluster-issuer")
	assert.Equal(t, "cert-manager", ingress.Annotations["cert-manager.io/cluster-issuer"])
	assert.Empty(t, ingress.Spec.TLS)
}

func TestXmtpdIngressTLSSecretNoCreate(t *testing.T) {

	options := &helm.Options{
		SetValues: map[string]string{
			"ingress.create":         "true",
			"ingress.tls.certIssuer": "cert-manager",
			"ingress.tls.secretName": "my-secret",
			"ingress.host":           "my-host",
		},
	}

	output := helm.RenderTemplate(t, options, testlib.XMTPD_HELM_CHART_PATH, "release-name", []string{})

	ingress := testlib.ExtractIngress(t, output)
	assert.Contains(t, ingress.Annotations, "cert-manager.io/cluster-issuer")
	assert.Equal(t, "cert-manager", ingress.Annotations["cert-manager.io/cluster-issuer"])

	expectedTLS := netv1.IngressTLS{
		Hosts:      []string{"my-host"},
		SecretName: "my-secret",
	}
	assert.Contains(t, ingress.Spec.TLS, expectedTLS)

	secret := testlib.ExtractNamedSecretE(t, output, "my-secret")
	assert.Nil(t, secret)
}

func TestXmtpdNoEnvWorks(t *testing.T) {

	options := &helm.Options{}

	output := helm.RenderTemplate(t, options, testlib.XMTPD_HELM_CHART_PATH, "release-name", []string{})
	deployment := testlib.ExtractDeployment(t, output, "release-name-xmtpd")

	assert.NotNil(t, deployment)
}

func TestXmtpdEnvWorks(t *testing.T) {

	options := &helm.Options{
		SetValues: map[string]string{
			"env.secret.XMTPD_LOG_LEVEL": "debug",
		},
	}

	output := helm.RenderTemplate(t, options, testlib.XMTPD_HELM_CHART_PATH, "release-name", []string{})
	deployment := testlib.ExtractDeployment(t, output, "release-name-xmtpd")

	assert.NotNil(t, deployment)
	assert.NotEmpty(t, deployment.Spec.Template.Spec.Containers[0].Env)
}

func TestXmtpdPruneCreate(t *testing.T) {
	assertScheduleEquals := func(cronjob *v1.CronJob, schedule string) {
		assert.Equal(t, cronjob.Spec.Schedule, schedule)
	}
	assertHistoriesEquals := func(cronjob *v1.CronJob, successfulJobsHistoryLimit int, failedJobsHistoryLimit int) {
		assert.EqualValues(t, successfulJobsHistoryLimit, *cronjob.Spec.SuccessfulJobsHistoryLimit)
		assert.EqualValues(t, failedJobsHistoryLimit, *cronjob.Spec.FailedJobsHistoryLimit)
	}

	tests := []struct {
		name       string
		options    *helm.Options
		exists     bool
		assertions func(*v1.CronJob)
	}{
		{"disable", &helm.Options{}, false, func(cronjob *v1.CronJob) {}},
		{"enable default", &helm.Options{
			SetValues: map[string]string{
				"prune.create": "true",
			},
		}, true, func(cronjob *v1.CronJob) {
			assertScheduleEquals(cronjob, "0 * * * *")
			assertHistoriesEquals(cronjob, 3, 1)
		}},
		{"set schedule", &helm.Options{
			SetValues: map[string]string{
				"prune.create":   "true",
				"prune.schedule": "* * * * *",
			},
		}, true, func(cronjob *v1.CronJob) {
			assertScheduleEquals(cronjob, "* * * * *")
		}},
		{"set limits", &helm.Options{
			SetValues: map[string]string{
				"prune.create":                     "true",
				"prune.successfulJobsHistoryLimit": "100",
				"prune.failedJobsHistoryLimit":     "100",
			},
		}, true, func(cronjob *v1.CronJob) {
			assertHistoriesEquals(cronjob, 100, 100)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := helm.RenderTemplate(t, tt.options, testlib.XMTPD_HELM_CHART_PATH, "release-name", []string{})
			cronjob := testlib.ExtractCronJobE(t, output, "release-name-xmtpd-prune")
			if !tt.exists {
				assert.Nil(t, cronjob)
				return
			}
			assert.NotNil(t, cronjob)
			tt.assertions(cronjob)
		})
	}
}
