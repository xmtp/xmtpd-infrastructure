package testlib

import (
	"context"
	"fmt"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const NAMESPACE_NAME_PREFIX = "test"

/** Used to wait for all async calls of GetAppLog() routine to finish before the test finishes */
var appLogCollectorsWg sync.WaitGroup

/** Lists of the teardown and diagnostic teardown funcs */
var teardownLists = make(map[string][]func())

/**
 * add a teardown function to the named list - for deferred execution.
 *
 * The teardown functions are called in reverse order of insertion, by a call to Teardown(name).
 *
 * The typical idiom is:
 * <pre>
 *   testlib.AddTeardown("DATABASE", func() { ...})
 *   // possibly more testlib.AddTeardown("DATABASE", func() { ... })
 *   defer testlib.Teardown("DATABASE")
 * <pre>
 */
func AddTeardown(name string, teardownFunc func()) {
	teardownLists[name] = append(teardownLists[name], teardownFunc)
}

/**
 * Call the stored teardown functions in the named list, in the correct order (last-in-first-out)
 *
 * NOTE: Any DIAGNOSTIC teardowns - those added with AddDiagnosticTeardown() for this name - are called BEFORE any other teardowns for this name.
 *
 * The typical use of Teardown is with a deferred call:
 * defer testlib.Teardown("SOME NAME")
 * See: testlib.AddTeardown(); testlib.AddDiagnosticTeardown()
 */
func Teardown(name string) {
	// ensure both list and diagnostic list are removed.
	defer func() { delete(teardownLists, name) }()

	list := teardownLists[name]

	for x := len(list) - 1; x >= 0; x-- {
		list[x]()
	}
}

/**
 * Adds a teardown function to all named teardown lists - for deferred execution.
 *
 * The teardown functions are called in reverse order of insertion, by a call to Teardown(name).
 *
 */
func AddGlobalTeardown(teardownFunc func()) {
	for name := range teardownLists {
		AddTeardown(name, teardownFunc)
	}
}

/**
* Verify all teardownLists have been executed already; and throw an require if not.
* Can be used to verify correct coding of a test that uses teardown - and to ensure eventual release of resources.
*
* NOTE: while the funcs are called in the correct order for each list,
* there can be NO guarantee that the lists are iterated in the correct order.
*
* This function MUST NOT be used as a replacement for calling teardown() at the correct point in the code.
 */

func VerifyTeardown(t *testing.T) {
	// ensure all funcs in all lists are released
	defer func() { teardownLists = make(map[string][]func()) }()

	// release all remaining resources - this is a "best effort" as the order of iterating the map is arbitrary
	uncleared := make([]string, 0)

	// make a "best-effort" at releasing all remaining resources
	for name, list := range teardownLists {
		uncleared = append(uncleared, name)

		for x := len(list) - 1; x >= 0; x-- {
			list[x]()
		}
	}

	require.Equal(t, 0, len(uncleared), "Error - %d teardownLists were left uncleared: %s", len(uncleared), uncleared)
	t.Log("Waiting for all logging collectors to finish")
	appLogCollectorsWg.Wait()
}

func GetAppLog(t *testing.T, namespace string, podName string, fileNameSuffix string, podLogOptions *corev1.PodLogOptions) string {
	defer appLogCollectorsWg.Done()
	appLogCollectorsWg.Add(1)
	dirPath := filepath.Join(RESULT_DIR, namespace)
	filePath := filepath.Join(dirPath, podName+fileNameSuffix+".log")

	_ = os.MkdirAll(dirPath, 0700)

	f, err := os.Create(filePath)
	require.NoError(t, err)
	defer f.Close()

	writer := io.Writer(f)

	reader, err := getAppLogStreamE(t, namespace, podName, podLogOptions)
	// avoid generating test failure just because container logs are not available
	if _, ok := err.(*ContainersNotStarted); ok {
		t.Logf("Skipping log collection for pod %s because no container has been started", podName)
		return ""
	}
	require.NoError(t, err)
	require.NotNil(t, reader)
	_, err = io.Copy(writer, reader)
	require.NoError(t, err)

	t.Logf("Finished reading log file %s", filePath)

	return filePath
}

func getAppLogStreamE(t *testing.T, namespace string, podName string, podLogOptions *corev1.PodLogOptions) (reader io.ReadCloser, err error) {
	options := k8s.NewKubectlOptions("", "", namespace)

	client, err := k8s.GetKubernetesClientFromOptionsE(t, options)

	if podLogOptions.Container == "" {
		// Select first container if not specified; otherwise the GetLogs method will fail if there are sidecars
		pod, e := client.CoreV1().Pods(options.Namespace).Get(context.TODO(), podName, metav1.GetOptions{})
		if e != nil {
			err = e
			return
		}
		container := findXmtpContainer(pod)
		if container == nil {
			err = &ContainersNotStarted{}
			return
		}
		podLogOptions.Container = container.Name
		for _, containerStatus := range pod.Status.ContainerStatuses {
			// if the container is in Waiting state (e.g. because
			// the pod is in CrashLoopBackOff state), then get the
			// logs from the previous invocation of the container
			if containerStatus.Name == container.Name && containerStatus.State.Waiting != nil {
				podLogOptions.Previous = true
			}
		}
		if podLogOptions.Previous {
			t.Logf("Multiple containers found in pod %s. Getting logs from previous container %s.", podName, podLogOptions.Container)
		} else {
			t.Logf("Multiple containers found in pod %s. Getting logs from container %s.", podName, podLogOptions.Container)
		}
	}

	reader, err = client.CoreV1().Pods(options.Namespace).GetLogs(podName, podLogOptions).Stream(context.TODO())
	return
}

func doesContainerHaveLogs(container *corev1.Container, containerStatuses []corev1.ContainerStatus) bool {
	for _, status := range containerStatuses {
		// check the status of the container; if it is in Waiting state,
		// then check that it has a non-0 restart count; otherwise the
		// container has no logs to retrieve
		if status.Name == container.Name && (status.State.Waiting == nil || status.RestartCount > 0) {
			return true
		}
	}
	return false
}

func findXmtpContainer(pod *corev1.Pod) *corev1.Container {
	// look for any container named "xmtp" that has logs
	for _, container := range pod.Spec.Containers {
		if (container.Name == "xmtpd") && doesContainerHaveLogs(&container, pod.Status.ContainerStatuses) {
			return &container
		}
	}
	// look for any container that has logs
	for _, container := range pod.Spec.Containers {
		if doesContainerHaveLogs(&container, pod.Status.ContainerStatuses) {
			return &container
		}
	}
	return nil
}

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

func Await(t *testing.T, lmbd func() bool, timeout time.Duration) {
	now := time.Now()
	for timeExpired := time.After(timeout); ; {
		select {
		case <-timeExpired:
			t.Logf("Function %s timed out", runtime.FuncForPC(reflect.ValueOf(lmbd).Pointer()).Name())
			t.Logf("Full stack trace of caller:\n%s", string(debug.Stack()))
			t.Fatalf("function call timed out after %f seconds. Start of await was '%s'", timeout.Seconds(), now)
		default:
			if lmbd() {
				return
			}

			time.Sleep(1 * time.Second)
		}
	}
}

func AwaitNrReplicasScheduled(t *testing.T, namespace string, expectedName string, nrReplicas int) {
	// in multi-cluster tests the Pods won't be scheduled until disks are
	// provisioned which takes longer than in minikube; adjust the timeout if
	// needed
	timeout := 5 * time.Second

	Await(t, func() bool {
		var pods []corev1.Pod
		var podNames string
		for _, pod := range FindAllPodsInSchema(t, namespace) {
			if strings.Contains(pod.Name, expectedName) {
				//ignore all completed pods
				if pod.Status.Phase == corev1.PodSucceeded {
					continue
				}

				if arePodConditionsMet(&pod, corev1.PodScheduled, corev1.ConditionTrue) {
					// build array of scheduled pods
					pods = append(pods, pod)

					// build formatted list of pod names
					if podNames != "" {
						podNames += ", "
					}
					podNames += pod.Name

					// log any pods not in Pending or Running phase
					if pod.Status.Phase != corev1.PodPending && pod.Status.Phase != corev1.PodRunning {
						t.Logf("Unexpected phase for pod %s: %s", pod.Name, pod.Status.Phase)
					}
				}
			}
		}

		t.Logf("%d pods SCHEDULED for name '%s': expected=%d, pods=[%s]\n", len(pods), expectedName, nrReplicas, podNames)

		return len(pods) == nrReplicas
	}, timeout)
}

func AwaitNrReplicasReady(t *testing.T, namespace string, expectedName string, nrReplicas int) {
	Await(t, func() bool {
		var cnt int
		for _, pod := range FindAllPodsInSchema(t, namespace) {
			if strings.Contains(pod.Name, expectedName) {
				if arePodConditionsMet(&pod, corev1.PodReady, corev1.ConditionTrue) {
					cnt++
				}
			}
		}

		t.Logf("%d pods READY for name '%s'\n", cnt, expectedName)

		return cnt == nrReplicas
	}, 30*time.Second)
}

func FindAllPodsInSchema(t *testing.T, namespace string) []corev1.Pod {
	options := k8s.NewKubectlOptions("", "", namespace)
	filter := metav1.ListOptions{}
	pods := k8s.ListPods(t, options, filter)
	sort.SliceStable(pods, func(i, j int) bool {
		return pods[j].CreationTimestamp.Before(&pods[i].CreationTimestamp)
	})
	return pods
}

func arePodConditionsMet(pod *corev1.Pod, condition corev1.PodConditionType,
	status corev1.ConditionStatus) bool {
	for _, cnd := range pod.Status.Conditions {
		if cnd.Type == condition && cnd.Status == status {
			return true
		}
	}

	return false
}

func FindPodsFromChart(t *testing.T, namespace string, expectedName string) []corev1.Pod {

	var pods []corev1.Pod

	for _, pod := range FindAllPodsInSchema(t, namespace) {
		if strings.Contains(pod.Name, expectedName) {
			pods = append(pods, pod)
		}
	}
	return pods
}
