/*
Copyright 2022 fmeng.

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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	etcdv1alpha1 "github.com/am2901/etcd-operator/api/v1alpha1"
)

// EtcdClusterReconciler reconciles a EtcdCluster object
type EtcdClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=etcd.fmeng.io,resources=etcdclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=etcd.fmeng.io,resources=etcdclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=etcd.fmeng.io,resources=etcdclusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the EtcdCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *EtcdClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	var etcdCluster etcdv1alpha1.EtcdCluster
	if err := r.Get(ctx, req.NamespacedName, &etcdCluster); err != nil {
		// EtcdCluster was gone，skip
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// CreateOrUpdate Service
	var svc corev1.Service
	svc.Name = etcdCluster.Name
	svc.Namespace = etcdCluster.Namespace
	or, err := ctrl.CreateOrUpdate(ctx, r.Client, &svc, func() error {
		MutateHeadlessSvc(&etcdCluster, &svc)
		return controllerutil.SetControllerReference(&etcdCluster, &svc, r.Scheme)
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	log.Log.Info("Create Or Update fmeng etcd", "Service", or)

	// CreateOrUpdate StatefulSet
	var sts appsv1.StatefulSet
	sts.Name = etcdCluster.Name
	sts.Namespace = etcdCluster.Namespace
	or, err = ctrl.CreateOrUpdate(ctx, r.Client, &sts, func() error {
		MutateStatefulSet(&etcdCluster, &sts)
		return controllerutil.SetControllerReference(&etcdCluster, &sts, r.Scheme)
	})
	if err != nil {
		return ctrl.Result{}, err
	}
	log.Log.Info("Create Or Update fmeng etcd", "statefulset", or)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EtcdClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&etcdv1alpha1.EtcdCluster{}).
		Complete(r)
}
