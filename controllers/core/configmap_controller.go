// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package controllers

import (
	"context"

	"github.com/aws/amazon-vpc-resource-controller-k8s/pkg/node/manager"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ConfigMapReconciler reconciles a ConfigMap object
type ConfigMapReconciler struct {
	client.Client
	Log         logr.Logger
	Scheme      *runtime.Scheme
	NodeManager manager.Manager
}

//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch
//+kubebuilder:rbac:groups=core,resources=configmaps/status,verbs=get;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ConfigMap object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.6.4/pkg/reconcile
func (r *ConfigMapReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Log.WithValues("configmap", req.NamespacedName)

	// We only want to update nodes on amazon-vpc-cni updates, return here for other updates
	if req.Name != "amazon-vpc-cni" {
		return ctrl.Result{}, nil
	}
	configmap := &corev1.ConfigMap{}
	if err := r.Client.Get(ctx, req.NamespacedName, configmap); err != nil {
		if errors.IsNotFound(err) {
			logger.Info(req.Name, " ConfigMap is deleted")
		} else {
			// Error reading the object
			logger.Error(err, "Failed to get ConfigMap ", req.Name)
			return ctrl.Result{}, err
		}
	}

	logger.Info("Enable-Windows-IPAM=", configmap.Data["enable-windows-ipam"])
	// Configmap is created/updated/deleted, update nodes
	err := r.NodeManager.UpdateNodesOnConfigMapChanges()
	if err != nil {
		// Error in updating nodes
		logger.Error(err, "Failed to update nodes on configmap changes")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConfigMapReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ConfigMap{}).
		Complete(r)
}
