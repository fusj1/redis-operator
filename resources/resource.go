package resources

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	rdv1beta1 "github.com/fusj1/redis-operator/api/v1beta1"
)

func NewStateFulSet(rediscluster *rdv1beta1.RedisCluster) *appsv1.StatefulSet {
	sts := &appsv1.StatefulSet{}
	labels := map[string]string{"app": rediscluster.Name}
	selector := &metav1.LabelSelector{MatchLabels: labels}
	sts.TypeMeta = metav1.TypeMeta{
		APIVersion: "rd.kanzhun.com",
		Kind:       "Statefulset",
	}
	sts.ObjectMeta = metav1.ObjectMeta{
		Name:      rediscluster.Name,
		Namespace: rediscluster.Namespace,
		OwnerReferences: []metav1.OwnerReference{
			*metav1.NewControllerRef(rediscluster, schema.GroupVersionKind{
				Group:   rdv1beta1.GroupVersion.Group,
				Version: rdv1beta1.GroupVersion.Version,
				Kind:    rediscluster.Kind,
			}),
		},
	}
	sts.Spec = appsv1.StatefulSetSpec{
		Replicas: rediscluster.Spec.Size,
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labels,
			},
			Spec: corev1.PodSpec{
				Containers: newContainers(rediscluster),
			},
		},
		Selector: selector,
	}

	return sts

}

func newContainers(rediscluster *rdv1beta1.RedisCluster) []corev1.Container {
	containerPorts := []corev1.ContainerPort{}
	for _, svcPort := range rediscluster.Spec.Ports {
		cport := corev1.ContainerPort{}
		cport.ContainerPort = svcPort.TargetPort.IntVal
		containerPorts = append(containerPorts, cport)
	}
	return []corev1.Container{
		{
			Name:            rediscluster.Name,
			Image:           rediscluster.Spec.Image,
			Ports:           containerPorts,
			ImagePullPolicy: pullpolicy(rediscluster.Spec.ImagePullPolicy),
			Env:             rediscluster.Spec.Envs,
			Resources:       rediscluster.Spec.Resources,
		},
	}
}

func pullpolicy(specPolicy corev1.PullPolicy) corev1.PullPolicy {
	if specPolicy == "" {
		return corev1.PullAlways
	}
	return specPolicy
}

func NewService(rediscluster *rdv1beta1.RedisCluster) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      rediscluster.Name,
			Namespace: rediscluster.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(rediscluster, schema.GroupVersionKind{
					Group:   rdv1beta1.GroupVersion.Group,
					Version: rdv1beta1.GroupVersion.Version,
					Kind:    rediscluster.Kind,
				}),
			},
		},
		Spec: corev1.ServiceSpec{
			Type:  corev1.ServiceTypeNodePort,
			Ports: rediscluster.Spec.Ports,
			Selector: map[string]string{
				"app": rediscluster.Name,
			},
		},
	}
}
