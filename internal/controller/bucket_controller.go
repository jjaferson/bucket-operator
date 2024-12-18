/*
Copyright 2024.

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

package controller

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	objectstoragev1alpha1 "mystorage.sh/bucket/api/v1alpha1"
	obj "mystorage.sh/bucket/pkg/objectstorage"
)

// BucketReconciler reconciles a Bucket object
type BucketReconciler struct {
	client.Client
	obj.ObjectStorageClient
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=objectstorage.mystorage.sh,resources=buckets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=objectstorage.mystorage.sh,resources=buckets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=objectstorage.mystorage.sh,resources=buckets/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;create;list;update;delete;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Bucket object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *BucketReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	bucket := objectstoragev1alpha1.Bucket{}
	if err := r.Client.Get(ctx, req.NamespacedName, &bucket); err != nil {
		if apierrors.IsNotFound(err) {
			// If the bucket resource is not found, it doesn't exists on the cluster
			// So we can stop the reconciliation
			return ctrl.Result{}, nil
		}

		// Returns the error and re-queu the controller
		log.Error(err, "failed to retrieve bucket object")
		return ctrl.Result{}, err
	}

	if !bucket.GetDeletionTimestamp().IsZero() {
		log.Info("deleting bucket")
		if err := r.ObjectStorageClient.DeleteBucket(ctx, &bucket); err != nil {
			log.Error(err, "failed to delete bucket")
			return ctrl.Result{}, err
		}

		// Remove finalizer
		controllerutil.RemoveFinalizer(&bucket, objectstoragev1alpha1.Finalizer)
		if err := r.Update(ctx, &bucket); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to remove finalizer of the bucket resource: %w", err)
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer to prevent resource from being deleted before clean up
	if !controllerutil.ContainsFinalizer(&bucket, objectstoragev1alpha1.Finalizer) {
		controllerutil.AddFinalizer(&bucket, objectstoragev1alpha1.Finalizer)
		if err := r.Update(ctx, &bucket); err != nil {
			log.Error(err, "failed to add finalizer to bucket resource")
			return ctrl.Result{}, err
		}
	}

	if err := r.ObjectStorageClient.CreateBucket(ctx, &bucket); err != nil {
		log.Error(err, "failed to create bucket")
		return ctrl.Result{}, err
	}

	//TODO: update status with address to access the bucket via API

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *BucketReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&objectstoragev1alpha1.Bucket{}).
		Complete(r)
}
