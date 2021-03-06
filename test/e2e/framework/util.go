package framework

import (
	"fmt"
	"path/filepath"

	"github.com/appscode/go/runtime"
	"github.com/appscode/go/types"
	exec_util "github.com/appscode/kutil/tools/exec"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

var (
	KubedTestConfigFileDir = filepath.Join(runtime.GOPath(), "src", "github.com", "appscode", "kubed", "test", "e2e", "config.yaml")
)

func deleteInBackground() *metav1.DeleteOptions {
	policy := metav1.DeletePropagationBackground
	return &metav1.DeleteOptions{PropagationPolicy: &policy}
}

func deleteInForeground() *metav1.DeleteOptions {
	policy := metav1.DeletePropagationForeground
	return &metav1.DeleteOptions{PropagationPolicy: &policy}
}

func (fi *Invocation) WaitUntilDeploymentReady(meta metav1.ObjectMeta) error {
	return wait.PollImmediate(interval, timeout, func() (done bool, err error) {
		if obj, err := fi.KubeClient.AppsV1beta1().Deployments(meta.Namespace).Get(meta.Name, metav1.GetOptions{}); err == nil {
			return types.Int32(obj.Spec.Replicas) == obj.Status.ReadyReplicas, nil
		}
		return false, nil
	})
}

func (fi *Invocation) WaitUntilDeploymentTerminated(meta metav1.ObjectMeta) error {
	return wait.PollImmediate(interval, timeout, func() (done bool, err error) {
		if pods, err := fi.KubeClient.CoreV1().Pods(meta.Namespace).List(metav1.ListOptions{}); err == nil {
			return len(pods.Items) == 0, nil
		}
		return false, nil
	})
}

func (fi *Invocation) RemoveFromOperatorPod(dir string) error {
	pod, err := fi.OperatorPod()
	if err != nil {
		return err
	}

	_, err = exec_util.ExecIntoPod(fi.ClientConfig, pod, "rm", "-rf", dir)
	if err != nil {
		return err
	}

	return nil
}

func (fi *Invocation) OperatorPod() (*core.Pod, error) {
	pods, err := fi.KubeClient.CoreV1().Pods(OperatorNamespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, pod := range pods.Items {
		for _, c := range pod.Spec.Containers {
			if c.Name == ContainerOperator {
				return &pod, nil
			}
		}
	}

	return nil, fmt.Errorf("pod not found")
}

func (fi *Invocation) DeleteService(meta metav1.ObjectMeta) error {
	return fi.KubeClient.CoreV1().Services(meta.Namespace).Delete(meta.Name, deleteInBackground())
}

func (fi *Invocation) DeleteEndpoints(meta metav1.ObjectMeta) error {
	return fi.KubeClient.CoreV1().Endpoints(meta.Namespace).Delete(meta.Name, deleteInBackground())
}
