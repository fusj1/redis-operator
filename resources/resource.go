package resources

import (
	rdv1beta1 "github.com/fusj1/redis-operator/api/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Mutatests(rediscluster *rdv1beta1.RedisCluster, sts *appsv1.StatefulSet) {
	labels := map[string]string{"app": rediscluster.Name}
	selector := &metav1.LabelSelector{MatchLabels: labels}
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
		VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
			corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name: "datadir",
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("1Gi"),
						},
					},
				},
			},
		},
		Selector: selector,
	}

}
func MutateSvc(rediscluster *rdv1beta1.RedisCluster, svc *corev1.Service) {
	svc.Spec = corev1.ServiceSpec{
		ClusterIP: corev1.ClusterIPNone,
		Ports: []corev1.ServicePort{
			{
				Name: "redis",
				Port: 6379,
			},
		},
		Selector: map[string]string{
			"app": rediscluster.Name,
		},
	}
}

func newContainers(rediscluster *rdv1beta1.RedisCluster) []corev1.Container {
	containerPorts := []corev1.ContainerPort{
		{
			Name:          "redis",
			ContainerPort: 6379,
		},
		{
			Name:          "client",
			ContainerPort: 6380,
		},
	}
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
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "datadir",
					MountPath: "/data",
				},
			},
		},
	}
}

func pullpolicy(specPolicy corev1.PullPolicy) corev1.PullPolicy {
	if specPolicy == "" {
		return corev1.PullAlways
	}
	return specPolicy
}
