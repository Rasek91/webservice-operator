/*
Copyright 2021.

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
	webservicev1 "github.com/Rasek91/webservice-operator/api/v1"
	cmv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *WebAppReconciler) certificate(webapp webservicev1.WebApp) (cmv1.Certificate, error) {
	certificate := cmv1.Certificate{
		TypeMeta: metav1.TypeMeta{APIVersion: cmv1.SchemeGroupVersion.String(), Kind: "Certificate"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      webapp.Name,
			Namespace: webapp.Namespace,
		},
		Spec: cmv1.CertificateSpec{
			SecretName: webapp.Name,
			IssuerRef:  cmmeta.ObjectReference{Name: webapp.Spec.Issuer},
			CommonName: webapp.Spec.Host,
			DNSNames:   []string{webapp.Spec.Host},
		},
	}

	if err := ctrl.SetControllerReference(&webapp, &certificate, r.Scheme); err != nil {
		return certificate, err
	}

	return certificate, nil
}

func (r *WebAppReconciler) deployment(webapp webservicev1.WebApp) (appsv1.Deployment, error) {
	deployment := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: appsv1.SchemeGroupVersion.String(), Kind: "Deployment"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      webapp.Name,
			Namespace: webapp.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: webapp.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"webapp": webapp.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"webapp": webapp.Name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "webapp",
							Image: webapp.Spec.Image,
							Ports: []corev1.ContainerPort{
								{ContainerPort: webapp.Spec.ContainerPort, Protocol: "TCP"},
							},
							Resources: *webapp.Spec.Resources.DeepCopy(),
						},
					},
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(&webapp, &deployment, r.Scheme); err != nil {
		return deployment, err
	}

	return deployment, nil
}

func (r *WebAppReconciler) service(webapp webservicev1.WebApp) (corev1.Service, error) {
	service := corev1.Service{
		TypeMeta: metav1.TypeMeta{APIVersion: corev1.SchemeGroupVersion.String(), Kind: "Service"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      webapp.Name,
			Namespace: webapp.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{Port: 80, Protocol: "TCP"},
			},
			Selector: map[string]string{"webapp": webapp.Name},
		},
	}

	if err := ctrl.SetControllerReference(&webapp, &service, r.Scheme); err != nil {
		return service, err
	}

	return service, nil
}

func (r *WebAppReconciler) ingress(webapp webservicev1.WebApp) (networkv1.Ingress, error) {
	pathtype := networkv1.PathTypePrefix
	ingress := networkv1.Ingress{
		TypeMeta: metav1.TypeMeta{APIVersion: networkv1.SchemeGroupVersion.String(), Kind: "Ingress"},
		ObjectMeta: metav1.ObjectMeta{
			Name:        webapp.Name,
			Namespace:   webapp.Namespace,
			Annotations: map[string]string{"cert-manager.io/issuer": webapp.Spec.Issuer},
		},
		Spec: networkv1.IngressSpec{
			Rules: []networkv1.IngressRule{
				{
					Host: webapp.Spec.Host,
					IngressRuleValue: networkv1.IngressRuleValue{
						HTTP: &networkv1.HTTPIngressRuleValue{
							Paths: []networkv1.HTTPIngressPath{
								{
									PathType: &pathtype,
									Path:     "/",
									Backend: networkv1.IngressBackend{
										Service: &networkv1.IngressServiceBackend{
											Name: webapp.Name,
											Port: networkv1.ServiceBackendPort{Number: 80},
										},
									},
								},
							},
						},
					},
				},
			},
			TLS: []networkv1.IngressTLS{
				{
					Hosts:      []string{webapp.Spec.Host},
					SecretName: webapp.Name,
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(&webapp, &ingress, r.Scheme); err != nil {
		return ingress, err
	}

	return ingress, nil
}
