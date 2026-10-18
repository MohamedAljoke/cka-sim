package kube

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/domain/labs"
	"github.com/MohamedAljoke/cka-sim/apps/platform/internal/infra/adapters/labagent"
)

const namespace = "cka-sim-labs"

var settings = PodSettings{Image: "ghcr.io/mohamedaljoke/cka-sim-lab:main", TokenSecret: "cka-sim-agent"}

func TestUpCreatesThePodTheManifestsExpect(t *testing.T) {
	client := fake.NewClientset()
	pod := NewPod(client.CoreV1().Pods(namespace), settings)

	if err := pod.Up(context.Background()); err != nil {
		t.Fatal(err)
	}
	got, err := client.CoreV1().Pods(namespace).Get(context.Background(), "lab", metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}

	// infra/gke/k8s/lab.yaml's NetworkPolicy and the labs node pool's label and taint depend on these.
	if got.Labels["app"] != "cka-sim-lab" {
		t.Errorf("labels %v, want app=cka-sim-lab", got.Labels)
	}
	if got.Spec.NodeSelector["pool"] != "labs" {
		t.Errorf("node selector %v, want pool=labs", got.Spec.NodeSelector)
	}
	if !tolerates(got.Spec.Tolerations, "lab", "true", corev1.TaintEffectNoSchedule) {
		t.Errorf("tolerations %+v miss lab=true:NoSchedule", got.Spec.Tolerations)
	}

	lab := got.Spec.Containers[0]
	if lab.Image != settings.Image || lab.SecurityContext == nil || !*lab.SecurityContext.Privileged {
		t.Errorf("container %s, privileged %v", lab.Image, lab.SecurityContext)
	}
	if ref := tokenRef(lab.Env); ref == nil || ref.Name != "cka-sim-agent" || ref.Key != "token" {
		t.Errorf("the token comes from %+v, want Secret cka-sim-agent key token", ref)
	}
	if got.Spec.RestartPolicy != corev1.RestartPolicyNever || *got.Spec.AutomountServiceAccountToken {
		t.Errorf("restart policy %s, service account token mounted %v", got.Spec.RestartPolicy,
			*got.Spec.AutomountServiceAccountToken)
	}
}

func TestUpTwiceKeepsTheOnePod(t *testing.T) {
	pod := NewPod(fake.NewClientset().CoreV1().Pods(namespace), settings)

	for range 2 {
		if err := pod.Up(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
}

func TestDownDeletesThePodAndSaysWhetherThereWasOne(t *testing.T) {
	client := fake.NewClientset(labPod(corev1.PodStatus{}))
	pod := NewPod(client.CoreV1().Pods(namespace), settings)

	for _, want := range []bool{true, false} {
		removed, err := pod.Down(context.Background())
		if err != nil || removed != want {
			t.Errorf("got %v, %v; want %v", removed, err, want)
		}
	}
}

func TestWhereFollowsThePodUntilItsAgentAnswers(t *testing.T) {
	deleted := metav1.NewTime(time.Now())
	for name, tc := range map[string]struct {
		pod  *corev1.Pod
		want labagent.Place
	}{
		"no pod": {want: labagent.Place{Lab: labs.Lab{State: labs.None}}},
		"not scheduled": {
			pod:  labPod(corev1.PodStatus{Phase: corev1.PodPending}),
			want: starting("Waiting for a node"),
		},
		"scheduled": {
			pod:  labPod(corev1.PodStatus{Phase: corev1.PodPending, Conditions: conditions(corev1.PodScheduled)}),
			want: starting("Pulling the lab image"),
		},
		"running": {
			pod:  labPod(corev1.PodStatus{Phase: corev1.PodRunning, Conditions: conditions(corev1.PodScheduled)}),
			want: starting("Starting Docker and the agent"),
		},
		"ready": {
			pod: labPod(corev1.PodStatus{Phase: corev1.PodRunning, PodIP: "10.8.0.5",
				Conditions: conditions(corev1.PodScheduled, corev1.PodReady)}),
			want: labagent.Place{URL: "http://10.8.0.5:7171", OnDemand: true},
		},
		"being deleted": {
			pod:  withDeletion(labPod(corev1.PodStatus{Phase: corev1.PodRunning}), &deleted),
			want: labagent.Place{Lab: labs.Lab{State: labs.Stopping}},
		},
	} {
		got, err := NewPod(clientWith(tc.pod).CoreV1().Pods(namespace), settings).Where(context.Background())
		got.Lab.StartedAt = time.Time{}
		if err != nil || got != tc.want {
			t.Errorf("%s: got %+v, %v; want %+v", name, got, err, tc.want)
		}
	}
}

func TestALabPodThatCantRunIsAFailedLab(t *testing.T) {
	pulling := corev1.ContainerStatus{Name: "lab", State: corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{
		Reason: "ImagePullBackOff", Message: "manifest unknown",
	}}}
	for name, tc := range map[string]struct {
		status corev1.PodStatus
		want   string
	}{
		"bad image": {
			status: corev1.PodStatus{Phase: corev1.PodPending, ContainerStatuses: []corev1.ContainerStatus{pulling}},
			want:   "the lab can't start: ImagePullBackOff: manifest unknown",
		},
		"evicted": {
			status: corev1.PodStatus{Phase: corev1.PodFailed, Reason: "Evicted"},
			want:   "the lab pod stopped: Evicted",
		},
		"exited": {
			status: corev1.PodStatus{Phase: corev1.PodSucceeded},
			want:   "the lab pod stopped",
		},
	} {
		got, err := NewPod(clientWith(labPod(tc.status)).CoreV1().Pods(namespace), settings).Where(context.Background())
		if err != nil || got.URL != "" || got.Lab.State != labs.Failed || got.Lab.Error != tc.want {
			t.Errorf("%s: got %+v, %v; want failed with %q", name, got, err, tc.want)
		}
	}
}

func TestAStartingLabStartedWhenItsPodWasMade(t *testing.T) {
	made := time.Date(2026, 10, 18, 9, 0, 0, 0, time.UTC)
	p := labPod(corev1.PodStatus{Phase: corev1.PodPending})
	p.CreationTimestamp = metav1.NewTime(made)

	got, err := NewPod(clientWith(p).CoreV1().Pods(namespace), settings).Where(context.Background())
	if err != nil || !got.Lab.StartedAt.Equal(made) {
		t.Errorf("got %+v, %v; want started at %s", got.Lab, err, made)
	}
}

func labPod(status corev1.PodStatus) *corev1.Pod {
	return &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "lab", Namespace: namespace}, Status: status}
}

func withDeletion(pod *corev1.Pod, at *metav1.Time) *corev1.Pod {
	pod.DeletionTimestamp = at
	pod.Finalizers = []string{"test/keep"}
	return pod
}

func clientWith(pod *corev1.Pod) *fake.Clientset {
	if pod == nil {
		return fake.NewClientset()
	}
	return fake.NewClientset([]runtime.Object{pod}...)
}

func conditions(kinds ...corev1.PodConditionType) []corev1.PodCondition {
	var all []corev1.PodCondition
	for _, kind := range kinds {
		all = append(all, corev1.PodCondition{Type: kind, Status: corev1.ConditionTrue})
	}
	return all
}

func starting(step string) labagent.Place {
	return labagent.Place{Lab: labs.Lab{State: labs.Starting, Step: step}}
}

func tolerates(all []corev1.Toleration, key, value string, effect corev1.TaintEffect) bool {
	for _, t := range all {
		if t.Key == key && t.Value == value && t.Effect == effect {
			return true
		}
	}
	return false
}

func tokenRef(env []corev1.EnvVar) *corev1.SecretKeySelector {
	for _, e := range env {
		if e.Name == "CKA_SIM_AGENT_TOKEN" && e.ValueFrom != nil {
			return e.ValueFrom.SecretKeyRef
		}
	}
	return nil
}
