package payloads

import (
	"github.com/getupio-undistro/inspect/apis/inspect/v1alpha1"
	"github.com/getupio-undistro/inspect/pkg/formats"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewCluster(c v1alpha1.Cluster) Cluster {
	res := &Resources{}
	if cpu, ok := c.Status.Resources[corev1.ResourceCPU]; ok {
		res.CPU = &Resource{
			Available:       formats.CPU(cpu.Available),
			Usage:           formats.CPU(cpu.Usage),
			UsagePercentage: cpu.UsagePercentage,
		}
	}
	if mem, ok := c.Status.Resources[corev1.ResourceMemory]; ok {
		res.Memory = &Resource{
			Available:       formats.Memory(mem.Available),
			Usage:           formats.Memory(mem.Usage),
			UsagePercentage: mem.UsagePercentage,
		}
	}
	return Cluster{
		Name:              c.Name,
		Namespace:         c.Namespace,
		Environment:       c.Labels[v1alpha1.LabelEnvironment],
		Provider:          c.Status.Provider,
		Region:            c.Status.Region,
		TotalNodes:        c.Status.TotalNodes,
		Ready:             true,
		Version:           c.Status.KubernetesVersion,
		TotalIssues:       0,
		Resources:         res,
		CreationTimestamp: c.Status.CreationTimestamp,
	}
}

type Cluster struct {
	Name              string      `json:"name,omitempty"`
	Namespace         string      `json:"namespace,omitempty"`
	Environment       string      `json:"environment,omitempty"`
	Provider          string      `json:"provider,omitempty"`
	Region            string      `json:"region,omitempty"`
	TotalNodes        int         `json:"totalNodes,omitempty"`
	Ready             bool        `json:"ready,omitempty"`
	Version           string      `json:"version,omitempty"`
	TotalIssues       int         `json:"totalIssues,omitempty"`
	Resources         *Resources  `json:"resources,omitempty"`
	CreationTimestamp metav1.Time `json:"creationTimestamp,omitempty"`
}

type Resources struct {
	Memory *Resource `json:"memory,omitempty"`
	CPU    *Resource `json:"cpu,omitempty"`
}

type Resource struct {
	Available       string `json:"available,omitempty"`
	Usage           string `json:"usage,omitempty"`
	UsagePercentage int32  `json:"usagePercentage,omitempty"`
}