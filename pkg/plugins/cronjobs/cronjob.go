package cronjobs

import (
	"path/filepath"

	"github.com/getupio-undistro/inspect/apis/inspect/v1alpha1"
	"github.com/getupio-undistro/inspect/pkg/kubeconfig"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	workerContainerName  = "worker"
	kubeconfigVolumeName = "kubeconfig"
	kubeconfigMountPath  = "/etc/undistro-inspect"
	kubeconfigFile       = "kubeconfig.yml"
	resultsVolumeName    = "results"
	resultsDir           = "/tmp/undistro-inspect/results"
)

var (
	// commonEnv environment variables to be used in worker and plugin containers
	commonEnv = []corev1.EnvVar{
		{
			Name:  "DONE_DIR",
			Value: resultsDir,
		},
	}
	// commonVolumeMounts volume mounts to be used in worker and plugin containers
	commonVolumeMounts = []corev1.VolumeMount{
		{
			Name:      resultsVolumeName,
			MountPath: resultsDir,
		},
	}
	// pluginVolumeMounts volume mounts to be used in plugin container
	pluginVolumeMounts = append(commonVolumeMounts, corev1.VolumeMount{
		Name:      kubeconfigVolumeName,
		ReadOnly:  true,
		MountPath: kubeconfigMountPath,
	})
)

func New(name, namespace string) *batchv1.CronJob {
	return &batchv1.CronJob{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}
}

type Mutator struct {
	Scheme             *runtime.Scheme
	Existing           *batchv1.CronJob
	Plugin             *v1alpha1.Plugin
	PluginRef          v1alpha1.PluginReference
	Clusterscan        *v1alpha1.ClusterScan
	KubeconfigSecret   *corev1.Secret
	WorkerImage        string
	ServiceAccountName string
}

// Mutate returns a function which mutates the existing CronJob into it's desired state.
func (r *Mutator) Mutate() controllerutil.MutateFn {
	return func() error {
		r.Existing.Spec.Schedule = firstNonEmptyString(r.PluginRef.Schedule, r.Clusterscan.Spec.Schedule)
		r.Existing.Spec.ConcurrencyPolicy = batchv1.ForbidConcurrent
		r.Existing.Spec.Suspend = firstNonNilBoolPointer(r.PluginRef.Suspend, r.Clusterscan.Spec.Suspend)
		r.Existing.Spec.JobTemplate.Spec.Template.Spec.RestartPolicy = corev1.RestartPolicyOnFailure
		r.Existing.Spec.JobTemplate.Spec.Template.Spec.ServiceAccountName = r.ServiceAccountName
		r.Existing.Spec.JobTemplate.Spec.Template.Spec.Volumes = []corev1.Volume{
			{
				Name: kubeconfigVolumeName,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName:  r.KubeconfigSecret.Name,
						DefaultMode: pointer.Int32(0644),
						Items:       []corev1.KeyToPath{{Key: kubeconfig.SecretField, Path: kubeconfigFile}},
					},
				},
			},
			{
				Name:         resultsVolumeName,
				VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
			},
		}

		desiredContainers := map[string]corev1.Container{
			workerContainerName: r.workerContainer(),
			r.Plugin.Name:       r.pluginContainer(),
		}

		foundContainers := 0
		for i, container := range r.Existing.Spec.JobTemplate.Spec.Template.Spec.Containers {
			desired, found := desiredContainers[container.Name]
			if found {
				foundContainers++
				r.Existing.Spec.JobTemplate.Spec.Template.Spec.Containers[i].Name = desired.Name
				r.Existing.Spec.JobTemplate.Spec.Template.Spec.Containers[i].Image = desired.Image
				r.Existing.Spec.JobTemplate.Spec.Template.Spec.Containers[i].Command = desired.Command
				r.Existing.Spec.JobTemplate.Spec.Template.Spec.Containers[i].Args = desired.Args
				r.Existing.Spec.JobTemplate.Spec.Template.Spec.Containers[i].EnvFrom = desired.EnvFrom
				r.Existing.Spec.JobTemplate.Spec.Template.Spec.Containers[i].Env = desired.Env
				r.Existing.Spec.JobTemplate.Spec.Template.Spec.Containers[i].Resources = desired.Resources
				r.Existing.Spec.JobTemplate.Spec.Template.Spec.Containers[i].ImagePullPolicy = desired.ImagePullPolicy
				r.Existing.Spec.JobTemplate.Spec.Template.Spec.Containers[i].SecurityContext = desired.SecurityContext
				r.Existing.Spec.JobTemplate.Spec.Template.Spec.Containers[i].VolumeMounts = desired.VolumeMounts
			}
		}
		if foundContainers != len(desiredContainers) {
			containers := make([]corev1.Container, 0, len(desiredContainers))
			for _, c := range desiredContainers {
				containers = append(containers, c)
			}
			r.Existing.Spec.JobTemplate.Spec.Template.Spec.Containers = containers
		}

		return ctrl.SetControllerReference(r.Clusterscan, r.Existing, r.Scheme)
	}
}

// workerContainer returns a Container for Worker
func (r *Mutator) workerContainer() corev1.Container {
	return corev1.Container{
		Name:            workerContainerName,
		Image:           r.WorkerImage,
		Command:         []string{"printenv"}, //TODO remove printenv command
		Env:             r.workerEnv(),
		ImagePullPolicy: corev1.PullIfNotPresent,
		VolumeMounts:    commonVolumeMounts,
	}
}

// pluginContainer returns a Container for Plugin
func (r *Mutator) pluginContainer() corev1.Container {
	return corev1.Container{
		Name:            r.Plugin.Name,
		Image:           r.Plugin.Spec.Image,
		Command:         r.Plugin.Spec.Command,
		Args:            r.Plugin.Spec.Args,
		EnvFrom:         r.Plugin.Spec.EnvFrom,
		Env:             r.pluginEnv(),
		Resources:       r.Plugin.Spec.Resources,
		ImagePullPolicy: r.Plugin.Spec.GetImagePullPolicy(),
		SecurityContext: r.Plugin.Spec.SecurityContext,
		VolumeMounts:    pluginVolumeMounts,
	}
}

// pluginEnv returns a list of environment variables to set in the Plugin container
func (r *Mutator) pluginEnv() []corev1.EnvVar {
	p := append(r.Plugin.Spec.Env, r.PluginRef.Env...)
	p = append(p, commonEnv...)
	p = append(p, corev1.EnvVar{
		Name:  "KUBECONFIG",
		Value: filepath.Join(kubeconfigMountPath, kubeconfigFile),
	})
	return p
}

// workerEnv returns a list of environment variables to set in the Worker container
func (r *Mutator) workerEnv() []corev1.EnvVar {
	return append(commonEnv,
		corev1.EnvVar{
			Name:  "CLUSTER_NAME",
			Value: r.Clusterscan.Name,
		},
		corev1.EnvVar{
			Name:  "CLUSTER_ISSUES_NAMESPACE",
			Value: r.Clusterscan.Namespace,
		},
		corev1.EnvVar{
			Name: "JOB_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.labels['job-name']", APIVersion: "v1"},
			},
		},
		corev1.EnvVar{
			Name: "JOB_NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.namespace", APIVersion: "v1"},
			},
		},
		corev1.EnvVar{
			Name: "JOB_UID",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.labels['controller-uid']", APIVersion: "v1"},
			},
		},
	)
}

func firstNonEmptyString(strings ...string) string {
	for _, s := range strings {
		if s != "" {
			return s
		}
	}
	return ""
}

func firstNonNilBoolPointer(pointers ...*bool) *bool {
	for _, b := range pointers {
		if b != nil {
			return b
		}
	}
	return pointer.Bool(false)
}