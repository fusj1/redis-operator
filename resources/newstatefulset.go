package resources

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/fusj1/redis-operator/api/v1beta1"
)

func NewStateFulSet(rediscluster *v1beta1.RedisCluster) *appsv1.StatefulSet {
	sts := &appsv1.StatefulSet{}
	return sts

}

func NewService(rediscluster *v1beta1.RedisCluster) *corev1.Service {
	svc := &corev1.Service{}
	return svc

}
