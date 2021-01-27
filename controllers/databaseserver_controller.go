/*


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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	k8s "k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	databasev1alpha1 "flow.stacc.dev/database-provisioning-poc/api/v1alpha1"
	postgres "flow.stacc.dev/database-provisioning-poc/pkg/db"
	kubernetes "flow.stacc.dev/database-provisioning-poc/pkg/kubernetes"
)

// DatabaseServerReconciler reconciles a DatabaseServer object
type DatabaseServerReconciler struct {
	client.Client
	Log                 logr.Logger
	Scheme              *runtime.Scheme
	KubernetesClient    *kubernetes.Client
	KubernetesClientset *k8s.Clientset
}

// +kubebuilder:rbac:groups=database.stacc.com,resources=databaseservers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=database.stacc.com,resources=databaseservers/status,verbs=get;update;patch

// Reconcile DatabaseServer
func (r *DatabaseServerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("databaseserver", req.NamespacedName)

	var databaseServer databasev1alpha1.DatabaseServer
	err := r.Get(ctx, req.NamespacedName, &databaseServer)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	secret, err := r.KubernetesClientset.CoreV1().Secrets(databaseServer.GetNamespace()).Get(databaseServer.Spec.SecretName, metav1.GetOptions{})
	if err != nil {
		log.Error(err, "Error obtaining secret")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	server := postgres.PostgresServer{
		Host: databaseServer.Spec.Postgres.Host,
		Port: databaseServer.Spec.Postgres.Port,
	}

	if msg, err := server.Connect(databaseServer.Spec.Postgres.Username, string(secret.Data["password"])); err != nil {
		log.Error(err, msg)
		databaseServer.Status.Connected = false
		if err := r.Status().Update(ctx, &databaseServer); err != nil {
			log.Error(err, "unable to update databaseServer status")
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	defer server.Disconnect()

	log.Info("Successfully connected to database")
	databaseServer.Status.Connected = true
	if err := r.Status().Update(ctx, &databaseServer); err != nil {
		log.Error(err, "unable to update databaseServer status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: time.Minute}, nil
}

// SetupWithManager for DatabaseServer
func (r *DatabaseServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&databasev1alpha1.DatabaseServer{}).
		Complete(r)
}
