package kube

import (
	"context"
	"fmt"
	"net"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/infra/adapters/labagent"
)

const (
	podName   = "lab"
	agentPort = "7171"
)

type PodSettings struct {
	Image string
	// The Secret, in the lab's namespace, whose "token" the agent accepts.
	TokenSecret string
}

// Pod is a lab machine made on demand: Up creates the lab pod, and the labs node pool grows a node for it; Down
// deletes it, and the node goes away.
type Pod struct {
	pods     corev1client.PodInterface
	settings PodSettings
}

// NewInClusterPod manages the lab pod in namespace with the identity of the pod the platform runs in.
func NewInClusterPod(namespace string, settings PodSettings) (*Pod, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("reach the Kubernetes API from inside the cluster: %w", err)
	}

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	return NewPod(client.CoreV1().Pods(namespace), settings), nil
}

func NewPod(pods corev1client.PodInterface, settings PodSettings) *Pod {
	return &Pod{pods: pods, settings: settings}
}

func (p *Pod) Up(ctx context.Context) error {
	_, err := p.pods.Create(ctx, p.settings.spec(), metav1.CreateOptions{})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("create the lab pod: %w", err)
	}

	return nil
}

func (p *Pod) Down(ctx context.Context) (bool, error) {
	err := p.pods.Delete(ctx, podName, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("delete the lab pod: %w", err)
	}

	return true, nil
}

func (p *Pod) Where(ctx context.Context) (labagent.Place, error) {
	pod, err := p.pods.Get(ctx, podName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return labagent.Place{Lab: labs.Lab{State: labs.None}}, nil
	}

	if err != nil {
		return labagent.Place{}, fmt.Errorf("read the lab pod: %w", err)
	}

	if lab, waiting := stateOf(pod); waiting {
		return labagent.Place{Lab: lab}, nil
	}

	return labagent.Place{URL: "http://" + net.JoinHostPort(pod.Status.PodIP, agentPort), OnDemand: true}, nil
}

// Pull and config errors don't fix themselves, so the lab fails instead of starting forever.
var brokenContainer = map[string]bool{
	"ErrImagePull": true, "ImagePullBackOff": true, "InvalidImageName": true,
	"CreateContainerConfigError": true, "CreateContainerError": true,
}

// The same container as the compose file's lab: Docker inside runs the kind cluster, so it is privileged. It gets a
// labs node to itself, and is never restarted: a crashed lab shows as failed, and Start makes a new one.
func (s PodSettings) spec() *corev1.Pod {
	privileged, noToken, noLinks, grace := true, false, false, int64(30)

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: podName, Labels: map[string]string{"app": "cka-sim-lab"}},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			NodeSelector:  map[string]string{"pool": "labs"},
			Tolerations: []corev1.Toleration{{
				Key: "lab", Operator: corev1.TolerationOpEqual, Value: "true", Effect: corev1.TaintEffectNoSchedule,
			}},
			AutomountServiceAccountToken:  &noToken,
			EnableServiceLinks:            &noLinks,
			TerminationGracePeriodSeconds: &grace,
			Containers: []corev1.Container{{
				Name:            "lab",
				Image:           s.Image,
				SecurityContext: &corev1.SecurityContext{Privileged: &privileged},
				Env: []corev1.EnvVar{
					{Name: "CKA_SIM_AGENT_ADDR", Value: ":" + agentPort},
					{Name: "CKA_SIM_CLUSTER", Value: "cka-sim"},
					{Name: "CKA_SIM_AGENT_TOKEN", ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: s.TokenSecret}, Key: "token",
					}}},
				},
				Ports: []corev1.ContainerPort{{Name: "agent", ContainerPort: 7171}},
				ReadinessProbe: &corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{Path: "/healthz", Port: intstr.FromString("agent")},
					},
					PeriodSeconds: 2,
				},
				Resources: corev1.ResourceRequirements{Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("2"),
					corev1.ResourceMemory: resource.MustParse("8Gi"),
				}},
				VolumeMounts: []corev1.VolumeMount{
					{Name: "docker", MountPath: "/var/lib/docker"},
					{Name: "modules", MountPath: "/lib/modules", ReadOnly: true},
				},
			}},
			Volumes: []corev1.Volume{
				{Name: "docker", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
				{Name: "modules", VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: "/lib/modules"}}},
			},
		},
	}
}

// stateOf says what the lab is while its pod is not yet, or no longer, one an agent answers in. It answers false
// once the agent is up and can say itself.
func stateOf(pod *corev1.Pod) (labs.Lab, bool) {
	starting := labs.Lab{State: labs.Starting, StartedAt: pod.CreationTimestamp.Time}

	switch {
	case pod.DeletionTimestamp != nil:
		return labs.Lab{State: labs.Stopping}, true
	case pod.Status.Phase == corev1.PodFailed || pod.Status.Phase == corev1.PodSucceeded:
		return labs.Lab{State: labs.Failed, Error: why("the lab pod stopped", pod.Status.Reason, pod.Status.Message)}, true
	}

	for _, c := range pod.Status.ContainerStatuses {
		if w := c.State.Waiting; w != nil && brokenContainer[w.Reason] {
			return labs.Lab{State: labs.Failed, Error: why("the lab can't start", w.Reason, w.Message)}, true
		}
	}

	switch {
	case !hasCondition(pod, corev1.PodScheduled):
		starting.Step = "Waiting for a node"
	case pod.Status.Phase == corev1.PodPending:
		starting.Step = "Pulling the lab image"
	case !hasCondition(pod, corev1.PodReady):
		starting.Step = "Starting Docker and the agent"
	default:
		return labs.Lab{}, false
	}

	return starting, true
}

func hasCondition(pod *corev1.Pod, kind corev1.PodConditionType) bool {
	for _, c := range pod.Status.Conditions {
		if c.Type == kind {
			return c.Status == corev1.ConditionTrue
		}
	}

	return false
}

func why(what string, details ...string) string {
	var parts []string

	for _, d := range details {
		if d != "" {
			parts = append(parts, d)
		}
	}

	if len(parts) == 0 {
		return what
	}

	return what + ": " + strings.Join(parts, ": ")
}
