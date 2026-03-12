package test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/entigolabs/entigo-infralib-common/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	terrak8s "github.com/gruntwork-io/terratest/modules/k8s"
	coreV1 "k8s.io/api/core/v1"
	storageV1 "k8s.io/api/storage/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	testTimeout       = 5 * time.Minute
	podReadyRetries   = 60
	podReadySleep     = 5 * time.Second
	podDeletedRetries = 30
	podDeletedSleep   = 5 * time.Second
)

func TestK8sAwsStorageclassBiz(t *testing.T) {
	testK8sAwsStorageclass(t, "aws", "biz")
}

func TestK8sAwsStorageclassPri(t *testing.T) {
	testK8sAwsStorageclass(t, "aws", "pri")
}

func testK8sAwsStorageclass(t *testing.T, cloudName string, envName string) {
	t.Parallel()
	kubectlOptions, _ := k8s.CheckKubectlConnection(t, cloudName, envName)

	// Use a unique suffix to avoid collisions
	uid := strings.ToLower(random.UniqueId())

	// ECR proxy image prefix per environment
	imagePrefix := fmt.Sprintf("877483565445.dkr.ecr.eu-north-1.amazonaws.com/%s-net-ecr-proxy-hub", envName)

	// Test that all expected storage classes exist with correct config
	t.Run("StorageClasses", func(t *testing.T) {
		t.Parallel()
		testStorageClassConfig(t, kubectlOptions, envName)
	})

	// Test GP3 (EBS) - RWO cross-node persistence
	t.Run("GP3", func(t *testing.T) {
		t.Parallel()
		testGP3CrossNode(t, kubectlOptions, imagePrefix, uid)
	})

	// Test EFS - RWX shared access between two pods
	// Biz has shared-a and shared-b, Pri has shared
	efsStorageClass := "shared"
	if envName == "biz" {
		efsStorageClass = "shared-a"
	}
	t.Run("EFS", func(t *testing.T) {
		t.Parallel()
		testEFSSharedAccess(t, kubectlOptions, efsStorageClass, imagePrefix, uid)
	})
}

// expectedSC defines the expected configuration of a storage class
type expectedSC struct {
	provisioner   string
	reclaimPolicy coreV1.PersistentVolumeReclaimPolicy
	bindingMode   storageV1.VolumeBindingMode
	allowExpand   bool
}

// testStorageClassConfig verifies all expected storage classes exist with correct configuration
func testStorageClassConfig(t *testing.T, opts *terrak8s.KubectlOptions, envName string) {
	clientSet, err := terrak8s.GetKubernetesClientFromOptionsE(t, opts)
	require.NoError(t, err)

	// Define expected storage classes per environment
	expected := map[string]expectedSC{
		"gp2":           {"ebs.csi.aws.com", coreV1.PersistentVolumeReclaimRetain, storageV1.VolumeBindingWaitForFirstConsumer, true},
		"gp2-no-retain": {"ebs.csi.aws.com", coreV1.PersistentVolumeReclaimDelete, storageV1.VolumeBindingWaitForFirstConsumer, true},
		"gp2-retain":    {"ebs.csi.aws.com", coreV1.PersistentVolumeReclaimRetain, storageV1.VolumeBindingWaitForFirstConsumer, true},
		"gp3":           {"ebs.csi.aws.com", coreV1.PersistentVolumeReclaimRetain, storageV1.VolumeBindingWaitForFirstConsumer, true},
		"gp3-no-retain": {"ebs.csi.aws.com", coreV1.PersistentVolumeReclaimDelete, storageV1.VolumeBindingWaitForFirstConsumer, true},
	}

	if envName == "biz" {
		expected["shared-a"] = expectedSC{"efs.csi.aws.com", coreV1.PersistentVolumeReclaimRetain, storageV1.VolumeBindingImmediate, true}
		expected["shared-a-no-retain"] = expectedSC{"efs.csi.aws.com", coreV1.PersistentVolumeReclaimDelete, storageV1.VolumeBindingImmediate, true}
		expected["shared-b"] = expectedSC{"efs.csi.aws.com", coreV1.PersistentVolumeReclaimRetain, storageV1.VolumeBindingImmediate, true}
		expected["shared-b-no-retain"] = expectedSC{"efs.csi.aws.com", coreV1.PersistentVolumeReclaimDelete, storageV1.VolumeBindingImmediate, true}
	} else {
		expected["shared"] = expectedSC{"efs.csi.aws.com", coreV1.PersistentVolumeReclaimRetain, storageV1.VolumeBindingImmediate, true}
		expected["shared-no-retain"] = expectedSC{"efs.csi.aws.com", coreV1.PersistentVolumeReclaimDelete, storageV1.VolumeBindingImmediate, true}
	}

	// Fetch all storage classes from the cluster
	scList, err := clientSet.StorageV1().StorageClasses().List(context.Background(), metaV1.ListOptions{})
	require.NoError(t, err, "Failed to list storage classes")

	// Build a map for quick lookup
	actual := make(map[string]storageV1.StorageClass)
	for _, sc := range scList.Items {
		actual[sc.Name] = sc
	}

	// Verify each expected storage class
	for name, exp := range expected {
		sc, exists := actual[name]
		assert.True(t, exists, "StorageClass %s not found", name)
		if !exists {
			continue
		}
		assert.Equal(t, exp.provisioner, sc.Provisioner, "StorageClass %s wrong provisioner", name)
		if sc.ReclaimPolicy != nil {
			assert.Equal(t, exp.reclaimPolicy, *sc.ReclaimPolicy, "StorageClass %s wrong reclaimPolicy", name)
		} else {
			t.Errorf("StorageClass %s reclaimPolicy is nil", name)
		}
		if sc.VolumeBindingMode != nil {
			assert.Equal(t, exp.bindingMode, *sc.VolumeBindingMode, "StorageClass %s wrong volumeBindingMode", name)
		} else {
			t.Errorf("StorageClass %s volumeBindingMode is nil", name)
		}
		if sc.AllowVolumeExpansion != nil {
			assert.Equal(t, exp.allowExpand, *sc.AllowVolumeExpansion, "StorageClass %s wrong allowVolumeExpansion", name)
		} else {
			t.Errorf("StorageClass %s allowVolumeExpansion is nil", name)
		}
	}
}

// testGP3CrossNode creates a PVC with gp3, writes data from a pod, deletes the pod,
// starts a new pod on a different node with anti-affinity, and reads the data back.
func testGP3CrossNode(t *testing.T, opts *terrak8s.KubectlOptions, imagePrefix string, uid string) {
	pvcName := fmt.Sprintf("gp3-test-%s", uid)
	writerPod := fmt.Sprintf("gp3-writer-%s", uid)
	readerPod := fmt.Sprintf("gp3-reader-%s", uid)
	testData := fmt.Sprintf("gp3-test-data-%s", uid)
	antiAffinityLabel := fmt.Sprintf("gp3-test-%s", uid)

	clientSet, err := terrak8s.GetKubernetesClientFromOptionsE(t, opts)
	require.NoError(t, err)
	ns := opts.Namespace

	// Create PVC
	pvc := newPVC(pvcName, "gp3", coreV1.ReadWriteOnce, "1Gi")
	_, err = clientSet.CoreV1().PersistentVolumeClaims(ns).Create(context.Background(), pvc, metaV1.CreateOptions{})
	require.NoError(t, err, "Failed to create GP3 PVC")

	// Create writer pod with a label for anti-affinity
	writer := newBusyboxPod(writerPod, pvcName, imagePrefix, map[string]string{"sc-test": antiAffinityLabel})
	writer.Spec.Containers[0].Command = []string{"sh", "-c", fmt.Sprintf("echo '%s' > /data/testfile && sleep 3600", testData)}
	_, err = clientSet.CoreV1().Pods(ns).Create(context.Background(), writer, metaV1.CreateOptions{})
	require.NoError(t, err, "Failed to create writer pod")

	// Wait for writer pod to be running
	waitForPodRunning(t, opts, writerPod)

	// Verify write succeeded
	output, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "exec", writerPod, "--", "cat", "/data/testfile")
	require.NoError(t, err, "Failed to read from writer pod")
	assert.Equal(t, testData, strings.TrimSpace(output), "Writer pod data mismatch")

	// Record which node the writer pod landed on
	writerNode := getPodNodeName(t, opts, writerPod)
	logger.Logf(t, "Writer pod scheduled on node: %s", writerNode)

	// Delete writer pod and wait for it to be gone
	err = clientSet.CoreV1().Pods(ns).Delete(context.Background(), writerPod, metaV1.DeleteOptions{})
	require.NoError(t, err, "Failed to delete writer pod")
	waitForPodDeleted(t, opts, writerPod)

	// Create reader pod with node affinity to prefer a different node
	reader := newBusyboxPod(readerPod, pvcName, imagePrefix, map[string]string{"sc-test": antiAffinityLabel})
	reader.Spec.Affinity = &coreV1.Affinity{
		NodeAffinity: &coreV1.NodeAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []coreV1.PreferredSchedulingTerm{
				{
					Weight: 100,
					Preference: coreV1.NodeSelectorTerm{
						MatchExpressions: []coreV1.NodeSelectorRequirement{
							{
								Key:      "kubernetes.io/hostname",
								Operator: coreV1.NodeSelectorOpNotIn,
								Values:   []string{writerNode},
							},
						},
					},
				},
			},
		},
	}
	reader.Spec.Containers[0].Command = []string{"sh", "-c", "sleep 3600"}
	_, err = clientSet.CoreV1().Pods(ns).Create(context.Background(), reader, metaV1.CreateOptions{})
	require.NoError(t, err, "Failed to create reader pod")

	// Wait for reader pod to be running
	waitForPodRunning(t, opts, readerPod)

	readerNode := getPodNodeName(t, opts, readerPod)
	logger.Logf(t, "Reader pod scheduled on node: %s", readerNode)

	// Verify data persisted across pod restart
	output, err = terrak8s.RunKubectlAndGetOutputE(t, opts, "exec", readerPod, "--", "cat", "/data/testfile")
	require.NoError(t, err, "Failed to read from reader pod")
	assert.Equal(t, testData, strings.TrimSpace(output), "GP3 data not persisted across pods")

	// Cleanup on success only
	_ = clientSet.CoreV1().Pods(ns).Delete(context.Background(), readerPod, metaV1.DeleteOptions{})
	waitForPodDeleted(t, opts, readerPod)
	_ = clientSet.CoreV1().PersistentVolumeClaims(ns).Delete(context.Background(), pvcName, metaV1.DeleteOptions{})
}

// testEFSSharedAccess creates a PVC with the given EFS storage class, starts two pods
// mounting the same PVC, writes from one and reads from the other.
func testEFSSharedAccess(t *testing.T, opts *terrak8s.KubectlOptions, storageClass string, imagePrefix string, uid string) {
	pvcName := fmt.Sprintf("efs-test-%s", uid)
	writerPod := fmt.Sprintf("efs-writer-%s", uid)
	readerPod := fmt.Sprintf("efs-reader-%s", uid)
	testData := fmt.Sprintf("efs-test-data-%s", uid)

	clientSet, err := terrak8s.GetKubernetesClientFromOptionsE(t, opts)
	require.NoError(t, err)
	ns := opts.Namespace

	// Create RWX PVC with EFS storage class
	pvc := newPVC(pvcName, storageClass, coreV1.ReadWriteMany, "1Gi")
	_, err = clientSet.CoreV1().PersistentVolumeClaims(ns).Create(context.Background(), pvc, metaV1.CreateOptions{})
	require.NoError(t, err, "Failed to create EFS PVC")

	// Create writer pod
	writer := newBusyboxPod(writerPod, pvcName, imagePrefix, nil)
	writer.Spec.Containers[0].Command = []string{"sh", "-c", fmt.Sprintf("echo '%s' > /data/testfile && sleep 3600", testData)}
	_, err = clientSet.CoreV1().Pods(ns).Create(context.Background(), writer, metaV1.CreateOptions{})
	require.NoError(t, err, "Failed to create EFS writer pod")

	// Create reader pod (concurrent with writer, both mount same PVC)
	reader := newBusyboxPod(readerPod, pvcName, imagePrefix, nil)
	reader.Spec.Containers[0].Command = []string{"sh", "-c", "sleep 3600"}
	_, err = clientSet.CoreV1().Pods(ns).Create(context.Background(), reader, metaV1.CreateOptions{})
	require.NoError(t, err, "Failed to create EFS reader pod")

	// Wait for both pods to be running
	waitForPodRunning(t, opts, writerPod)
	waitForPodRunning(t, opts, readerPod)

	// Give writer a moment to flush
	time.Sleep(2 * time.Second)

	// Read from the reader pod to verify shared access
	output, err := terrak8s.RunKubectlAndGetOutputE(t, opts, "exec", readerPod, "--", "cat", "/data/testfile")
	require.NoError(t, err, "Failed to read from EFS reader pod")
	assert.Equal(t, testData, strings.TrimSpace(output), "EFS shared data not readable from second pod")

	// Cleanup on success only
	_ = clientSet.CoreV1().Pods(ns).Delete(context.Background(), writerPod, metaV1.DeleteOptions{})
	_ = clientSet.CoreV1().Pods(ns).Delete(context.Background(), readerPod, metaV1.DeleteOptions{})
	waitForPodDeleted(t, opts, writerPod)
	waitForPodDeleted(t, opts, readerPod)
	_ = clientSet.CoreV1().PersistentVolumeClaims(ns).Delete(context.Background(), pvcName, metaV1.DeleteOptions{})
}

// newPVC creates a PersistentVolumeClaim spec
func newPVC(name, storageClass string, accessMode coreV1.PersistentVolumeAccessMode, size string) *coreV1.PersistentVolumeClaim {
	return &coreV1.PersistentVolumeClaim{
		ObjectMeta: metaV1.ObjectMeta{
			Name: name,
		},
		Spec: coreV1.PersistentVolumeClaimSpec{
			AccessModes:      []coreV1.PersistentVolumeAccessMode{accessMode},
			StorageClassName: &storageClass,
			Resources: coreV1.ResourceRequirements{
				Requests: coreV1.ResourceList{
					coreV1.ResourceStorage: resource.MustParse(size),
				},
			},
		},
	}
}

// newBusyboxPod creates a minimal pod spec that mounts a PVC at /data
// Compliant with PSA restricted profile
func newBusyboxPod(name, pvcName, imagePrefix string, labels map[string]string) *coreV1.Pod {
	nonRoot := true
	uid := int64(1000)
	gid := int64(1000)
	gracePeriod := int64(1)
	return &coreV1.Pod{
		ObjectMeta: metaV1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: coreV1.PodSpec{
			SecurityContext: &coreV1.PodSecurityContext{
				RunAsNonRoot: &nonRoot,
				RunAsUser:    &uid,
				RunAsGroup:   &gid,
				FSGroup:      &gid,
				SeccompProfile: &coreV1.SeccompProfile{
					Type: coreV1.SeccompProfileTypeRuntimeDefault,
				},
			},
			TerminationGracePeriodSeconds: &gracePeriod,
			Containers: []coreV1.Container{
				{
					Name:  "test",
					Image: fmt.Sprintf("%s/library/busybox:1.36", imagePrefix),
					SecurityContext: &coreV1.SecurityContext{
						AllowPrivilegeEscalation: func() *bool { b := false; return &b }(),
						Capabilities: &coreV1.Capabilities{
							Drop: []coreV1.Capability{"ALL"},
						},
					},
					VolumeMounts: []coreV1.VolumeMount{
						{
							Name:      "data",
							MountPath: "/data",
						},
					},
				},
			},
			Volumes: []coreV1.Volume{
				{
					Name: "data",
					VolumeSource: coreV1.VolumeSource{
						PersistentVolumeClaim: &coreV1.PersistentVolumeClaimVolumeSource{
							ClaimName: pvcName,
						},
					},
				},
			},
			RestartPolicy: coreV1.RestartPolicyNever,
		},
	}
}

// waitForPodRunning waits until a pod is in Running phase
func waitForPodRunning(t *testing.T, opts *terrak8s.KubectlOptions, podName string) {
	_, err := retry.DoWithRetryE(t,
		fmt.Sprintf("Wait for pod %s to be running", podName),
		podReadyRetries, podReadySleep,
		func() (string, error) {
			pod, err := terrak8s.GetPodE(t, opts, podName)
			if err != nil {
				return "", err
			}
			if pod.Status.Phase != coreV1.PodRunning {
				return "", fmt.Errorf("pod %s is in phase %s", podName, pod.Status.Phase)
			}
			return "Pod is running", nil
		},
	)
	require.NoError(t, err, "Pod %s did not reach Running phase", podName)
}

// waitForPodDeleted waits until a pod no longer exists
func waitForPodDeleted(t *testing.T, opts *terrak8s.KubectlOptions, podName string) {
	_, err := retry.DoWithRetryE(t,
		fmt.Sprintf("Wait for pod %s to be deleted", podName),
		podDeletedRetries, podDeletedSleep,
		func() (string, error) {
			_, err := terrak8s.GetPodE(t, opts, podName)
			if err != nil {
				return "Pod deleted", nil
			}
			return "", fmt.Errorf("pod %s still exists", podName)
		},
	)
	require.NoError(t, err, "Pod %s was not deleted", podName)
}

// getPodNodeName returns the node name where a pod is scheduled
func getPodNodeName(t *testing.T, opts *terrak8s.KubectlOptions, podName string) string {
	pod, err := terrak8s.GetPodE(t, opts, podName)
	require.NoError(t, err, "Failed to get pod %s", podName)
	return pod.Spec.NodeName
}
