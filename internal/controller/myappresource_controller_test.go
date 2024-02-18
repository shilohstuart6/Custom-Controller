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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	myv1alpha1 "github.com/shilohstuart6/Custom-Controller.git/api/v1alpha1"
)

var _ = Describe("MyAppResource Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // Modify as needed
		}
		myappresource := &myv1alpha1.MyAppResource{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind MyAppResource")
			err := k8sClient.Get(ctx, typeNamespacedName, myappresource)
			if err != nil && errors.IsNotFound(err) {
				myappresource = &myv1alpha1.MyAppResource{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: myv1alpha1.MyAppResourceSpec{
						ReplicaCount: 1,
						Resources: myv1alpha1.RequestsAndLimits{
							MemoryRequest: "32Mi",
							MemoryLimit:   "64Mi",
							CpuRequest:    "100m",
							CpuLimit:      "200m",
						},
						Image: myv1alpha1.Image{
							Repository: "ghcr.io/stefanprodan/podinfo",
							Tag:        "latest",
						},
						UI: myv1alpha1.UI{
							Color:   "#111111",
							Message: "hello",
						},
						Redis: myv1alpha1.Redis{Enabled: true},
					},
				}
				Expect(k8sClient.Create(ctx, myappresource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &myv1alpha1.MyAppResource{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance MyAppResource")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &MyAppResourceReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			deployment := appsv1.Deployment{}
			err = k8sClient.Get(ctx, typeNamespacedName, &deployment)
			Expect(err).NotTo(HaveOccurred())

			// Check deployment spec is correct
			Expect(*deployment.Spec.Replicas).To(BeEquivalentTo(myappresource.Spec.ReplicaCount))

			i := 0
			if myappresource.Spec.Redis.Enabled {
				// Check redis container
				Expect(deployment.Spec.Template.Spec.Containers[i].Name).To(BeEquivalentTo("redis"))
				Expect(deployment.Spec.Template.Spec.Containers[i].Image).To(BeEquivalentTo("redis:latest"))
				Expect(deployment.Spec.Template.Spec.Containers[i].Ports[0].ContainerPort).To(BeEquivalentTo(6379))
				Expect(deployment.Spec.Template.Spec.Containers[i].Ports[0].Protocol).To(BeEquivalentTo("TCP"))
				Expect(deployment.Spec.Template.Spec.Containers[i].Resources.Requests.Memory().String()).To(
					BeEquivalentTo(myappresource.Spec.Resources.MemoryRequest))
				Expect(deployment.Spec.Template.Spec.Containers[i].Resources.Limits.Memory().String()).To(
					BeEquivalentTo(myappresource.Spec.Resources.MemoryLimit))
				Expect(deployment.Spec.Template.Spec.Containers[i].Resources.Requests.Cpu().String()).To(
					BeEquivalentTo(myappresource.Spec.Resources.CpuRequest))
				Expect(deployment.Spec.Template.Spec.Containers[i].Resources.Limits.Cpu().String()).To(
					BeEquivalentTo(myappresource.Spec.Resources.CpuLimit))
				i += 1
			}
			// Check podinfo container
			image := myappresource.Spec.Image.Repository + ":" + myappresource.Spec.Image.Tag
			Expect(deployment.Spec.Template.Spec.Containers[i].Name).To(BeEquivalentTo("podinfo"))
			Expect(deployment.Spec.Template.Spec.Containers[i].Image).To(BeEquivalentTo(image))
			Expect(deployment.Spec.Template.Spec.Containers[i].Ports[0].ContainerPort).To(BeEquivalentTo(9898))
			Expect(deployment.Spec.Template.Spec.Containers[i].Ports[0].Protocol).To(BeEquivalentTo("TCP"))
			Expect(deployment.Spec.Template.Spec.Containers[i].Resources.Requests.Memory().String()).To(
				BeEquivalentTo(myappresource.Spec.Resources.MemoryRequest))
			Expect(deployment.Spec.Template.Spec.Containers[i].Resources.Limits.Memory().String()).To(
				BeEquivalentTo(myappresource.Spec.Resources.MemoryLimit))
			Expect(deployment.Spec.Template.Spec.Containers[i].Resources.Requests.Cpu().String()).To(
				BeEquivalentTo(myappresource.Spec.Resources.CpuRequest))
			Expect(deployment.Spec.Template.Spec.Containers[i].Resources.Limits.Cpu().String()).To(
				BeEquivalentTo(myappresource.Spec.Resources.CpuLimit))
			for _, envvar := range deployment.Spec.Template.Spec.Containers[i].Env {
				if envvar.Name == "PODINFO_UI_COLOR" {
					Expect(envvar.Value).To(BeEquivalentTo(myappresource.Spec.UI.Color))
				} else if envvar.Name == "PODINFO_UI_MESSAGE" {
					Expect(envvar.Value).To(BeEquivalentTo(myappresource.Spec.UI.Message))
				}
			}
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling an updated resource")
			controllerReconciler := &MyAppResourceReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			myappresource = &myv1alpha1.MyAppResource{
				ObjectMeta: metav1.ObjectMeta{
					Name:            resourceName,
					Namespace:       "default",
					ResourceVersion: myappresource.ObjectMeta.ResourceVersion,
				},
				Spec: myv1alpha1.MyAppResourceSpec{
					ReplicaCount: 2,
					Resources: myv1alpha1.RequestsAndLimits{
						MemoryRequest: "16Mi",
						MemoryLimit:   "32Mi",
						CpuRequest:    "50m",
						CpuLimit:      "100m",
					},
					Image: myv1alpha1.Image{
						Repository: "ghcr.io/stefanprodan/podinfo",
						Tag:        "latest",
					},
					UI: myv1alpha1.UI{
						Color:   "#333333",
						Message: "blegh",
					},
					Redis: myv1alpha1.Redis{Enabled: false},
				},
			}
			Expect(k8sClient.Update(ctx, myappresource)).To(Succeed())
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			deployment := appsv1.Deployment{}
			err = k8sClient.Get(ctx, typeNamespacedName, &deployment)
			Expect(err).NotTo(HaveOccurred())

			// Check deployment spec is correct
			Expect(*deployment.Spec.Replicas).To(BeEquivalentTo(myappresource.Spec.ReplicaCount))

			i := 0
			if myappresource.Spec.Redis.Enabled {
				// Check redis container
				Expect(deployment.Spec.Template.Spec.Containers[i].Name).To(BeEquivalentTo("redis"))
				Expect(deployment.Spec.Template.Spec.Containers[i].Image).To(BeEquivalentTo("redis:latest"))
				Expect(deployment.Spec.Template.Spec.Containers[i].Ports[0].ContainerPort).To(BeEquivalentTo(6379))
				Expect(deployment.Spec.Template.Spec.Containers[i].Ports[0].Protocol).To(BeEquivalentTo("TCP"))
				Expect(deployment.Spec.Template.Spec.Containers[i].Resources.Requests.Memory().String()).To(
					BeEquivalentTo(myappresource.Spec.Resources.MemoryRequest))
				Expect(deployment.Spec.Template.Spec.Containers[i].Resources.Limits.Memory().String()).To(
					BeEquivalentTo(myappresource.Spec.Resources.MemoryLimit))
				Expect(deployment.Spec.Template.Spec.Containers[i].Resources.Requests.Cpu().String()).To(
					BeEquivalentTo(myappresource.Spec.Resources.CpuRequest))
				Expect(deployment.Spec.Template.Spec.Containers[i].Resources.Limits.Cpu().String()).To(
					BeEquivalentTo(myappresource.Spec.Resources.CpuLimit))
				i += 1
			}
			// Check podinfo container
			image := myappresource.Spec.Image.Repository + ":" + myappresource.Spec.Image.Tag
			Expect(deployment.Spec.Template.Spec.Containers[i].Name).To(BeEquivalentTo("podinfo"))
			Expect(deployment.Spec.Template.Spec.Containers[i].Image).To(BeEquivalentTo(image))
			Expect(deployment.Spec.Template.Spec.Containers[i].Ports[0].ContainerPort).To(BeEquivalentTo(9898))
			Expect(deployment.Spec.Template.Spec.Containers[i].Ports[0].Protocol).To(BeEquivalentTo("TCP"))
			Expect(deployment.Spec.Template.Spec.Containers[i].Resources.Requests.Memory().String()).To(
				BeEquivalentTo(myappresource.Spec.Resources.MemoryRequest))
			Expect(deployment.Spec.Template.Spec.Containers[i].Resources.Limits.Memory().String()).To(
				BeEquivalentTo(myappresource.Spec.Resources.MemoryLimit))
			Expect(deployment.Spec.Template.Spec.Containers[i].Resources.Requests.Cpu().String()).To(
				BeEquivalentTo(myappresource.Spec.Resources.CpuRequest))
			Expect(deployment.Spec.Template.Spec.Containers[i].Resources.Limits.Cpu().String()).To(
				BeEquivalentTo(myappresource.Spec.Resources.CpuLimit))
			for _, envvar := range deployment.Spec.Template.Spec.Containers[i].Env {
				if envvar.Name == "PODINFO_UI_COLOR" {
					Expect(envvar.Value).To(BeEquivalentTo(myappresource.Spec.UI.Color))
				} else if envvar.Name == "PODINFO_UI_MESSAGE" {
					Expect(envvar.Value).To(BeEquivalentTo(myappresource.Spec.UI.Message))
				}
			}
		})
	})
})
