/*
Copyright 2024 shiliohstuart6.

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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	myv1alpha1 "github.com/shilohstuart6/Custom-Controller.git/api/v1alpha1"
)

// MyAppResourceReconciler reconciles a MyAppResource object
type MyAppResourceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=my.api.group,resources=myappresources,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=my.api.group,resources=myappresources/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=my.api.group,resources=myappresources/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *MyAppResourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	// Get the custom resource
	mar := myv1alpha1.MyAppResource{}
	if err := r.Get(ctx, req.NamespacedName, &mar); err != nil {
		if client.IgnoreNotFound(err) != nil {
			l.Error(err, "Failed to fetch MyAppResource")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	l.Info("Reconciling", "Name", mar.Name, "Namespace", mar.Namespace)

	// Create a deployment spec based on custom resource
	l.Info("Making Deployment Spec")
	deploymentSpec, err := r.createSpec(mar)
	if err != nil {
		l.Error(err, "Failed to create Deployment Spec")
		return ctrl.Result{}, err
	}
	if err = ctrl.SetControllerReference(&mar, &deploymentSpec, r.Scheme); err != nil {
		l.Error(err, "Failed to set deployment controller reference")
		return ctrl.Result{}, err
	}

	deploymentName := types.NamespacedName{
		Namespace: deploymentSpec.Namespace,
		Name:      deploymentSpec.Name,
	}
	// Check if deployment already exists
	if err := r.Get(ctx, deploymentName, &appsv1.Deployment{}); err != nil {
		if client.IgnoreNotFound(err) != nil {
			l.Error(err, "Failed to check for existing deployment")
			return ctrl.Result{}, err
		}

		// Deployment not found - create it
		l.Info("Creating Deployment")
		if err = r.Create(ctx, &deploymentSpec); err != nil {
			l.Error(err, "Failed to create Deployment")
			return ctrl.Result{}, client.IgnoreAlreadyExists(err)
		}

		l.Info("Deployment created", "Name", mar.Name, "Namespace", mar.Namespace)
		return ctrl.Result{}, nil
	}

	// Update existing deployment
	l.Info("Updating Deployment")
	err = r.Update(ctx, &deploymentSpec)
	if err != nil {
		l.Error(err, "Failed to update Deployment")
	}

	l.Info("Deployment updated", "Name", mar.Name, "Namespace", mar.Namespace)
	return ctrl.Result{}, nil
}

func (r *MyAppResourceReconciler) createSpec(mar myv1alpha1.MyAppResource) (appsv1.Deployment, error) {
	if mar.Spec.Redis.Enabled {
		return r.createSpecWithRedis(mar)
	}
	return r.createSpecNoRedis(mar)
}

func (r *MyAppResourceReconciler) createSpecNoRedis(mar myv1alpha1.MyAppResource) (appsv1.Deployment, error) {
	memR, err := resource.ParseQuantity(mar.Spec.Resources.MemoryRequest)
	if err != nil {
		return appsv1.Deployment{}, err
	}
	memL, err := resource.ParseQuantity(mar.Spec.Resources.MemoryLimit)
	if err != nil {
		return appsv1.Deployment{}, err
	}
	cpuR, err := resource.ParseQuantity(mar.Spec.Resources.CpuRequest)
	if err != nil {
		return appsv1.Deployment{}, err
	}
	cpuL, err := resource.ParseQuantity(mar.Spec.Resources.CpuLimit)
	if err != nil {
		return appsv1.Deployment{}, err
	}

	d := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{Kind: "Deployment"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      mar.Name,
			Namespace: mar.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &mar.Spec.ReplicaCount,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "myappresource",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "myappresource",
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyAlways,
					Containers: []corev1.Container{
						{
							Name:    "podinfo",
							Image:   mar.Spec.Image.Repository + ":" + mar.Spec.Image.Tag,
							Command: []string{"./podinfo", "--port=9898"},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    memR,
									corev1.ResourceMemory: memL,
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    cpuR,
									corev1.ResourceMemory: cpuL,
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 9898,
									Protocol:      "TCP",
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "PODINFO_UI_COLOR",
									Value: mar.Spec.UI.Color,
								},
								{
									Name:  "PODINFO_UI_MESSAGE",
									Value: mar.Spec.UI.Message,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "conf",
						},
						{
							Name: "data",
						},
					},
				},
			},
		},
	}

	return d, nil
}

func (r *MyAppResourceReconciler) createSpecWithRedis(mar myv1alpha1.MyAppResource) (appsv1.Deployment, error) {
	memR, err := resource.ParseQuantity(mar.Spec.Resources.MemoryRequest)
	if err != nil {
		return appsv1.Deployment{}, err
	}
	memL, err := resource.ParseQuantity(mar.Spec.Resources.MemoryLimit)
	if err != nil {
		return appsv1.Deployment{}, err
	}
	cpuR, err := resource.ParseQuantity(mar.Spec.Resources.CpuRequest)
	if err != nil {
		return appsv1.Deployment{}, err
	}
	cpuL, err := resource.ParseQuantity(mar.Spec.Resources.CpuLimit)
	if err != nil {
		return appsv1.Deployment{}, err
	}

	d := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{Kind: "Deployment"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      mar.Name,
			Namespace: mar.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &mar.Spec.ReplicaCount,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "myappresource",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "myappresource",
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyAlways,
					Containers: []corev1.Container{
						{
							Name:    "redis",
							Image:   "redis:latest",
							Command: []string{"redis-server"},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    memR,
									corev1.ResourceMemory: memL,
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    cpuR,
									corev1.ResourceMemory: cpuL,
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "client",
									ContainerPort: 6379,
									Protocol:      "TCP",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "conf",
									MountPath: "/conf",
								},
								{
									Name:      "data",
									MountPath: "/data",
								},
							},
						},
						{
							Name:    "podinfo",
							Image:   mar.Spec.Image.Repository + ":" + mar.Spec.Image.Tag,
							Command: []string{"./podinfo", "--port=9898"},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    memR,
									corev1.ResourceMemory: memL,
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    cpuR,
									corev1.ResourceMemory: cpuL,
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 9898,
									Protocol:      "TCP",
								},
							},
							Env: []corev1.EnvVar{
								{
									Name: "POD_IP",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "status.podIP",
										},
									},
								},
								{
									Name:  "PODINFO_UI_COLOR",
									Value: mar.Spec.UI.Color,
								},
								{
									Name:  "PODINFO_UI_MESSAGE",
									Value: mar.Spec.UI.Message,
								},
								{
									Name:  "PODINFO_CACHE_SERVER",
									Value: "tcp://$(POD_IP):6379",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "conf",
						},
						{
							Name: "data",
						},
					},
				},
			},
		},
	}

	return d, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MyAppResourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&myv1alpha1.MyAppResource{}).
		Complete(r)
}
