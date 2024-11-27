package testlib

import (
	"fmt"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
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

func CreateRandomNamespace(t *testing.T, depth int) string {
	randomSuffix := strings.ToLower(random.UniqueId())
	callerName := getFunctionCallerName(depth)
	namespaceName := fmt.Sprintf("%s-%s", strings.ToLower(callerName), randomSuffix)
	CreateNamespace(t, namespaceName)

	return namespaceName
}

func CreateNamespace(t *testing.T, namespaceName string) {
	kubectlOptions := k8s.NewKubectlOptions("", "", namespaceName)

	k8s.CreateNamespace(t, kubectlOptions, namespaceName)

	// this method is async
	go GetK8sEventLog(t, namespaceName)

	AddTeardown(TEARDOWN_GLOBAL, func() {
		k8s.DeleteNamespace(t, kubectlOptions, namespaceName)
	})
}
