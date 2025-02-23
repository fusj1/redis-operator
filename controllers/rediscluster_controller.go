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

	rdv1beta1 "github.com/fusj1/redis-operator/api/v1beta1"
	"github.com/fusj1/redis-operator/resources"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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
	if err := r.Get(ctx, req.NamespacedName, &redisCluster); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// reconcile service
	var svc corev1.Service
	svc.Name = redisCluster.Name
	svc.Namespace = redisCluster.Namespace
	or, err := ctrl.CreateOrUpdate(ctx, r.Client, &svc, func() error {
		resources.MutateSvc(&redisCluster, &svc)
		return controllerutil.SetControllerReference(&redisCluster, &svc, r.Scheme)
	})
	if err != nil {
		return ctrl.Result{}, err
	}
	log.Info("fetch rediscluster", "service", or)

	// reconcile sts
	var sts appsv1.StatefulSet
	sts.Name = redisCluster.Name
	sts.Namespace = redisCluster.Namespace
	or, err = ctrl.CreateOrUpdate(ctx, r.Client, &sts, func() error {
		resources.Mutatests(&redisCluster, &sts)
		return controllerutil.SetControllerReference(&redisCluster, &svc, r.Scheme)
	})
	if err != nil {
		return ctrl.Result{}, err
	}
	log.Info("fetch rediscluster", "statefulset", or)

	return ctrl.Result{}, nil

}

// SetupWithManager sets up the controller with the Manager.
func (r *RedisClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&rdv1beta1.RedisCluster{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
