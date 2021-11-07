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
	"context"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	webservicev1 "github.com/Rasek91/webservice-operator/api/v1"
	cmv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
)

// WebAppReconciler reconciles a WebApp object
type WebAppReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// WebApp permissions
//+kubebuilder:rbac:groups=webservice.my.domain,resources=webapps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=webservice.my.domain,resources=webapps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=webservice.my.domain,resources=webapps/finalizers,verbs=update

// Certificate permissions
//+kubebuilder:rbac:groups=cert-manager.io,resources=certificates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cert-manager.io,resources=certificates/status,verbs=get

// Deployment permissions
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get

// Service permissions
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete

// Ingress permissions
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete

func (r *WebAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("WebApp", req.NamespacedName)
	log.Info("reconciling webapp")
	applyOpts := []client.PatchOption{client.ForceOwnership, client.FieldOwner("webapp-controller")}
	var webapp webservicev1.WebApp

	if error := r.Get(ctx, req.NamespacedName, &webapp); error != nil {
		return ctrl.Result{}, client.IgnoreNotFound(error)
	}

	certificate, error := r.certificate(webapp)
	log.Info("reconciling certificate")

	if error != nil {
		return ctrl.Result{}, error
	}

	error = r.Patch(ctx, &certificate, client.Apply, applyOpts...)
	log.Info("force ownership certificate")

	if error != nil {
		return ctrl.Result{}, error
	}

	certificate_status := "Unknown"

	if len(certificate.Status.Conditions) != 0 {
		certificate_status = string(certificate.Status.Conditions[0].Status)
	}

	deployment, error := r.deployment(webapp)
	log.Info("reconciling deployment")

	if error != nil {
		return ctrl.Result{}, error
	}

	error = r.Patch(ctx, &deployment, client.Apply, applyOpts...)
	log.Info("force ownership deployment")

	if error != nil {
		return ctrl.Result{}, error
	}

	service, error := r.service(webapp)
	log.Info("reconciling service")

	if error != nil {
		return ctrl.Result{}, error
	}

	error = r.Patch(ctx, &service, client.Apply, applyOpts...)
	log.Info("force ownership service")

	if error != nil {
		return ctrl.Result{}, error
	}

	var ingress networkv1.Ingress

	if certificate_status == "True" {
		log.Info("certificate is ready")
		ingress, error = r.ingress(webapp)
		log.Info("reconciling ingress")

		if error != nil {
			return ctrl.Result{}, error
		}

		error = r.Patch(ctx, &ingress, client.Apply, applyOpts...)
		log.Info("force ownership ingress")

		if error != nil {
			return ctrl.Result{}, error
		}
	}

	log.Info("setup status")

	if certificate_status == "True" {
		webapp.Status.Host = ingress.Spec.Rules[0].Host
	} else {
		webapp.Status.Host = ""
	}

	webapp.Status.Replicas = deployment.Status.Replicas

	if len(certificate.Status.Conditions) != 0 {
		webapp.Status.CertificateStatus = certificate.Status.Conditions[0].Message
	} else {
		webapp.Status.CertificateStatus = ""
	}

	error = r.Status().Update(ctx, &webapp)

	if error != nil {
		return ctrl.Result{}, error
	}

	if certificate_status != "True" {
		log.Info("certificate is not ready")
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	log.Info("reconciled webapp")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *WebAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webservicev1.WebApp{}).
		Owns(&cmv1.Certificate{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&networkv1.Ingress{}).
		Complete(r)
}
