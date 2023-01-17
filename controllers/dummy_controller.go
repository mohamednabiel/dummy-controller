/*
Copyright 2023.

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

	interviewcomv1alpha1 "github.com/anupamgogoi/anynines-homework/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// DummyReconciler reconciles a Dummy object
type DummyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=interview.com,resources=dummies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=interview.com,resources=dummies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=interview.com,resources=dummies/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=pods/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Dummy object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
const dummfinalizer = "dummy.finalizer.interview.com"

func (r *DummyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	log := log.Log.WithName("controllers").WithName("Dummy")
	log.Info("Stating to Reconciling Dummy custom resource")
	dummy := &interviewcomv1alpha1.Dummy{}
	if err := r.Get(ctx, req.NamespacedName, dummy); err != nil {
		if errors.IsNotFound(err) {
			log.Info("Dummy custom resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, nil
	}
	log.Info("Dummy custom resource fetched successfully", "Dummy.Name:", dummy.Name, "Dummy.Namespace:", dummy.Namespace, "Dummy.Spec.Message:", dummy.Spec.Message)
	//udate the SpecEcho status
	dummy.Status.SpecEcho = dummy.Spec.Message
	if err := r.Status().Update(ctx, dummy); err != nil {
		log.Info("unable to update Dummy custom resource SpecEcho status", "Dummy.Name:", dummy.Name, "Dummy.Namespace:", dummy.Namespace, "Dummy.Status.SpecEcho:", dummy.Status.SpecEcho)
		return ctrl.Result{}, nil
	}
	log.Info("Dummy custom resource status updated successfully", "Dummy.Name:", dummy.Name, "Dummy.Namespace:", dummy.Namespace, "Dummy.Status.SpecEcho:", dummy.Status.SpecEcho)
	//check if the dummy custom resource status is pending
	if dummy.Status.PodStatus == "Pending" {
		log.Info("Dummy Status:PodStatus is Pending", "Dummy.Name:", dummy.Name, "Dummy.Namespace:", dummy.Namespace)
	} else if dummy.Status.PodStatus == "Running" {
		log.Info("Dummy custom resource Status:PodStatus is Running and updated", "Dummy.Name:", dummy.Name, "Dummy.Namespace:", dummy.Namespace)
	} else {
		log.Info("Dummy custom resource Status:PodStatus is not updated yet", "Dummy.Name:", dummy.Name, "Dummy.Namespace:", dummy.Namespace)
		dummy.Status.PodStatus = "Pending"
		if err := r.Status().Update(ctx, dummy); err != nil {
			log.Info("unable to update Dummy custom resource Status:PodStatus", "Dummy.Name:", dummy.Name, "Dummy.Namespace:", dummy.Namespace)
			return ctrl.Result{}, nil
		}
		log.Info("Dummy custom resource Status:PodStatus is updated, Pending", "Dummy.Name:", dummy.Name, "Dummy.Namespace:", dummy.Namespace)
	}
	pod := newPodForCR(dummy)

	// Set Dummy instance as the owner and controller
	if err := ctrl.SetControllerReference(dummy, pod, r.Scheme); err != nil {
		log.Info("unable to set owner reference on new pod", "pod.Name:", pod.Name, "pod.Namespace:", pod.Namespace)
		return ctrl.Result{}, nil
	}
	//check exesting of the nginx  pod and create if not found
	found := &corev1.Pod{}
	err := r.Get(ctx, client.ObjectKey{Name: pod.Name, Namespace: pod.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating a new Nginx Pod", "pod.Name:", pod.Name, "pod.Namespace:", pod.Namespace)
		err = r.Create(ctx, pod)
		if err != nil {
			log.Info("unable to create Nginx pod", "pod.Name:", pod.Name, "pod.Namespace:", pod.Namespace)
			return ctrl.Result{}, nil
		}
		dummy.Status.PodStatus = "Running"
		if err := r.Status().Update(ctx, dummy); err != nil {
			log.Info("unable to update custom resource Dummy Status:PodStatus ", "Dummy.Name:", dummy.Name, "Dummy.Namespace:", dummy.Namespace)
			return ctrl.Result{}, nil
		}
		log.Info("Dummy custom resource Status:PodStatus  is Running", "Dummy.Name:", dummy.Name, "Dummy.Namespace:", dummy.Namespace)
		log.Info("Nginx Pod created successfully", "pod.Name:", pod.Name, "pod.Namespace:", pod.Namespace)

	} else if err != nil {
		log.Info("unable to get pods", "pod.Name:", pod.Name, "pod.Namespace:", pod.Namespace)
		return ctrl.Result{}, nil
	}
	log.Info("Pod already exists", "pod.Name:", pod.Name, "pod.Namespace:", pod.Namespace)

	//check if the dummy resource is marked to be deleted by checking the deletion timestamp is set
	if dummy.GetDeletionTimestamp() == nil {
		// The dummy is not being deleted,
		// we add our finalizer  here  if it's not updated and update the dummy object.
		if !containsString(dummy.GetFinalizers(), dummfinalizer) {
			dummy.SetFinalizers(append(dummy.GetFinalizers(), dummfinalizer))
			if err := r.Update(ctx, dummy); err != nil {
				log.Info("unable to update Dummy custom resource with finalizer", "Dummy.Name:", dummy.Name, "Dummy.Namespace:", dummy.Namespace)
				return ctrl.Result{}, nil
			}
			log.Info("Dummy custom resource updated successfully with finalizer", "Dummy.Name:", dummy.Name, "Dummy.Namespace:", dummy.Namespace)
		}
	} else {
		// The dummy is being deleted
		if containsString(dummy.GetFinalizers(), dummfinalizer) {
			// that means our finalizer is present,we only cleanthepod here
			if err := r.deleteAttachedPod(dummy); err != nil {
				log.Info("unable to delete attached Nginx pod", "pod.Name:", pod.Name, "pod.Namespace:", pod.Namespace)
				return ctrl.Result{}, nil
			}
			log.Info("Dummy custom resource is marked to be deleted", "Dummy.Name:", dummy.Name, "Dummy.Namespace:", dummy.Namespace)
			log.Info("Dummy custom resource deleted successfully", "Dummy.Name:", dummy.Name, "Dummy.Namespace:", dummy.Namespace)
			log.Info("Pod deleted successfully", "pod.Name:", pod.Name, "pod.Namespace:", pod.Namespace)
			// remove our finalizer from the list and update it.
			dummy.SetFinalizers(removeString(dummy.GetFinalizers(), dummfinalizer))
			if err := r.Update(ctx, dummy); err != nil {
				log.Info("unable to remove finalizer from Dummy custom resource ", "Dummy.Name:", dummy.Name, "Dummy.Namespace:", dummy.Namespace)
				return ctrl.Result{}, nil
			}
			log.Info("Dummy custom resource updated successfully with finalizer", "Dummy.Name:", dummy.Name, "Dummy.Namespace:", dummy.Namespace)
		}

	}
	return ctrl.Result{}, nil

}
func removeString(s []string, e string) []string {
	for i, v := range s {
		if v == e {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

func (r *DummyReconciler) deleteAttachedPod(cr *interviewcomv1alpha1.Dummy) error {
	pod := newPodForCR(cr)
	err := r.Delete(context.Background(), pod)
	if err != nil {
		return err
	}
	return nil
}

func containsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func newPodForCR(cr *interviewcomv1alpha1.Dummy) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "nginx-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "nginx",
					Image:   "nginx:latest",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *DummyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&interviewcomv1alpha1.Dummy{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
