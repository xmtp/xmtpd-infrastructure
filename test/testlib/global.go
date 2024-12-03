package testlib

import (
	"context"
	"fmt"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func getFunctionCallerName(depth int) string {
	pc, _, _, _ := runtime.Caller(depth)
	nameFull := runtime.FuncForPC(pc).Name() // main.foo
	nameEnd := filepath.Ext(nameFull)        // .foo
	name := strings.TrimPrefix(nameEnd, ".") // foo

	return name
}

// CreateRandomNamespace
/**
 * CreateRandomNamespace creates a Kubernetes namespace with a random suffix derived from the caller function name.
 *
 * @param t *testing.T - The testing context.
 * @param depth int - The depth in the call stack to identify the caller function name.
 *
 * @return string - Returns the created namespace name.
 *
 * The function generates a namespace name based on the caller function name and a random suffix.
 * It creates the namespace in the Kubernetes cluster and registers a teardown function to delete it.
 * The depth is intended to point to the name of the test so the namespace has the test name in it
 */
func CreateRandomNamespace(t *testing.T, depth int) string {
	randomSuffix := strings.ToLower(random.UniqueId())
	callerName := getFunctionCallerName(depth)
	namespaceName := fmt.Sprintf("%s-%s", strings.ToLower(callerName), randomSuffix)
	CreateNamespace(t, namespaceName)

	return namespaceName
}

// CreateNamespace
/**
 * CreateNamespace creates a specified namespace in the Kubernetes cluster.
 *
 * @param t *testing.T - The testing context.
 * @param namespaceName string - The name of the namespace to create.
 *
 * The function creates the specified namespace and registers a teardown function to delete the namespace.
 * It also starts logging events for the namespace in the background.
 */
func CreateNamespace(t *testing.T, namespaceName string) {
	kubectlOptions := k8s.NewKubectlOptions("", "", namespaceName)

	k8s.CreateNamespace(t, kubectlOptions, namespaceName)

	go GetK8sEventLog(t, namespaceName)

	AddTeardown(TEARDOWN_GLOBAL, func() {
		k8s.DeleteNamespace(t, kubectlOptions, namespaceName)
	})
}

// GetK8sEventLog
/**
 * GetK8sEventLog collects Kubernetes event logs for the specified namespace and writes them to a file.
 *
 * @param t *testing.T - The testing context.
 * @param namespace string - The namespace of the Kubernetes cluster.
 *
 */
func GetK8sEventLog(t *testing.T, namespace string) {
	dirPath := filepath.Join(RESULT_DIR, namespace)
	filePath := filepath.Join(dirPath, K8S_EVENT_LOG_FILE)

	_ = os.MkdirAll(dirPath, 0700)

	f, err := os.Create(filePath)
	require.NoError(t, err)
	defer f.Close()

	options := k8s.NewKubectlOptions("", "", namespace)

	client, err := k8s.GetKubernetesClientFromOptionsE(t, options)
	require.NoError(t, err)

	var opts metav1.ListOptions

	events, err := client.CoreV1().Events(namespace).Watch(context.TODO(), opts)
	require.NoError(t, err)

	writer := io.Writer(f)

	for event := range events.ResultChan() {
		_, err = fmt.Fprintln(writer, event)
		require.NoError(t, err)
	}
}
