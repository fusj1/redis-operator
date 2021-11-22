/*
Copyright 2021 fusj1.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"encoding/json"
	"reflect"

	rdv1beta1 "github.com/fusj1/redis-operator/api/v1beta1"
	resources "github.com/fusj1/redis-operator/resources"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// RedisClusterReconciler reconciles a RedisCluster object
type RedisClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps,resources=statefulset,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=service,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rd.kanzhun.com,resources=redisclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=rd.kanzhun.com,resources=redisclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=rd.kanzhun.com,resources=redisclusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the RedisCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *RedisClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	var redisCluster rdv1beta1.RedisCluster
	err := r.Get(ctx, req.NamespacedName, &redisCluster)
	if err != nil {
		// RedisCluster 被删除的时候 忽略
		if client.IgnoreNotFound(err) != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
	log.Info("fetch rediscluster", "rediscluster================", redisCluster)
	//   如果不存在，则创建关联资源
	//   如果存在，判断是否需要更新
	//   如果需要更新，则直接更新
	//   如果不需要更新，则正常返回
	//   关联sts
	sts := resources.NewStateFulSet(&redisCluster)
	if err := r.Get(ctx, req.NamespacedName, sts); err != nil && errors.IsNotFound(err) {
		// 1、关联Annotation
		data, _ := json.Marshal(sts.Spec)
		if redisCluster.Annotations != nil {
			redisCluster.Annotations["oldSpecAnnotation"] = string(data)
		} else {
			redisCluster.Annotations = map[string]string{"oldSpecAnnotation": string(data)}
		}
		if err := r.Client.Update(ctx, &redisCluster); err != nil {
			return ctrl.Result{}, err
		}
		// 创建关联资源
		// 1、创建statefulset
		sts := resources.NewStateFulSet(&redisCluster)
		if err := r.Client.Create(ctx, sts); err != nil {
			return ctrl.Result{}, err
		}
		// 2、创建service资源
		svc := resources.NewService(&redisCluster)
		if err := r.Client.Create(ctx, svc); err != nil {
			return ctrl.Result{}, err
		}
	}
	oldSpec := rdv1beta1.RedisClusterSpec{}
	if err := json.Unmarshal([]byte(redisCluster.Annotations["oldSpecAnnotation"]), &oldSpec); err != nil {
		return ctrl.Result{}, err
	}
	// 当前规范与旧的不一致的时候需要更新
	if !reflect.DeepEqual(redisCluster.Spec, oldSpec) {
		newSts := resources.NewStateFulSet(&redisCluster)
		oldSts := &appsv1.StatefulSet{}
		if err := r.Get(ctx, req.NamespacedName, oldSts); err != nil {
			return ctrl.Result{}, err
		}
		oldSts.Spec = newSts.Spec
		if err := r.Client.Update(ctx, oldSts); err != nil {
			return ctrl.Result{}, err
		}
		newSvc := resources.NewService(&redisCluster)
		oldSvc := &corev1.Service{}
		if err := r.Client.Get(ctx, req.NamespacedName, oldSvc); err != nil {
			return ctrl.Result{}, err
		}
		newSvc.Spec.ClusterIP = oldSvc.Spec.ClusterIP
		oldSvc.Spec = newSvc.Spec
		if err := r.Client.Update(ctx, oldSvc); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RedisClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&rdv1beta1.RedisCluster{}).
		Complete(r)
}
