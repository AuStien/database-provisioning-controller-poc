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
	"strings"
	"time"

	databasev1alpha1 "flow.stacc.dev/database-provisioning-poc/api/v1alpha1"
	db "flow.stacc.dev/database-provisioning-poc/pkg/db"
	kubernetes "flow.stacc.dev/database-provisioning-poc/pkg/kubernetes"
	"github.com/go-logr/logr"

	"github.com/sethvargo/go-password/password"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8s "k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DatabaseReconciler reconciles a Database object
type DatabaseReconciler struct {
	client.Client
	Log                 logr.Logger
	Scheme              *runtime.Scheme
	KubernetesClient    *kubernetes.Client
	KubernetesClientset *k8s.Clientset
}

// +kubebuilder:rbac:groups=database.stacc.com,resources=databases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=database.stacc.com,resources=databases/status,verbs=get;update;patch

func (r *DatabaseReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("database", req.NamespacedName)

	finalizer := "database.stacc.com/finalizer"

	// Get database resource
	var database databasev1alpha1.Database
	err := r.Get(ctx, req.NamespacedName, &database)
	if err != nil {
		log.Info("Uanble to get database resource")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Get database Server resource
	var databaseServer databasev1alpha1.DatabaseServer
	err = r.Get(ctx, client.ObjectKey{Namespace: database.Spec.Server.Namespace, Name: database.Spec.Server.Name}, &databaseServer)
	if err != nil {
		log.Error(err, "uanble to get databaseServer resource. Retrying in 10 seconds.")
		return ctrl.Result{RequeueAfter: time.Second * 10}, client.IgnoreNotFound(err)
	}

	// Stop reconsiling if database server is not connected
	if databaseServer.Status.Connected == false {
		log.Info("Database server not connected. Retrying in 10 seconds.")
		return ctrl.Result{RequeueAfter: time.Second * 10}, nil
	}

	// Get secret with database server password
	serverSecret, err := r.KubernetesClientset.CoreV1().Secrets(databaseServer.Spec.Secret.Namespace).Get(databaseServer.Spec.Secret.Name, metav1.GetOptions{})
	if err != nil {
		log.Error(err, "Error obtaining secret. Retrying in 1 minute.")
		return ctrl.Result{RequeueAfter: time.Minute}, client.IgnoreNotFound(err)
	}

	// Get username, set to database name if not present
	username := database.Spec.Username
	if username == "" {
		username = database.Spec.Name
	}

	var pass string

	// Check if database secret exists
	dbSecret, err := r.KubernetesClientset.CoreV1().Secrets(database.Spec.Secret.Namespace).Get(database.Spec.Secret.Name, metav1.GetOptions{})
	if err != nil {
		// If error is other than "Not found" stop reconsiling
		if !errors.IsNotFound(err) {
			log.Error(err, "unable to create secret")
			return ctrl.Result{}, err
		}
		// Generate a password with length 48, 10 digits, allow uppercase, allow repeated chars
		genPass, err := password.Generate(48, 10, 0, false, true)
		if err != nil {
			log.Error(err, "unable to generate password")
			return ctrl.Result{}, err
		}
		pass = genPass
		// Create database secret
		dbSecret = &coreV1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      database.Spec.Secret.Name,
				Namespace: database.Spec.Secret.Namespace,
			},
			Data: map[string][]byte{
				"username": []byte(username),
				"password": []byte(pass),
			},
		}
		_, err = r.KubernetesClientset.CoreV1().Secrets(database.Spec.Secret.Namespace).Create(dbSecret)
		if err != nil {
			log.Error(err, "unable to create secret")
			return ctrl.Result{}, err
		}
		database.Status.CreatedSecret = true
		if err := r.Status().Update(ctx, &database); err != nil {
			log.Error(err, "unable to update database status")
			return ctrl.Result{}, err
		}
	} else {
		pass = string(dbSecret.Data["password"])
	}

	var sqlServer db.SQLServer

	if databaseServer.Spec.Type == "postgresql" || databaseServer.Spec.Type == "postgres" {
		sqlServer = &db.PostgresServer{
			Username: databaseServer.Spec.Postgres.Username,
			Password: string(serverSecret.Data["password"]),
			Host:     databaseServer.Spec.Postgres.Host,
			Port:     databaseServer.Spec.Postgres.Port,
			SslMode:  databaseServer.Spec.Postgres.SslMode,
			Postgres: db.Postgres{
				Name:     database.Spec.Name,
				Username: username,
				Password: pass,
			},
		}
	} else if databaseServer.Spec.Type == "mysql" {
		sqlServer = &db.MysqlServer{
			Username: databaseServer.Spec.Mysql.Username,
			Password: string(serverSecret.Data["password"]),
			Host:     databaseServer.Spec.Mysql.Host,
			Port:     databaseServer.Spec.Mysql.Port,
			Ssl:      databaseServer.Spec.Mysql.Ssl,

			Mysql: db.Mysql{
				Name:     database.Spec.Name,
				Username: username,
				Password: pass,
			},
		}
	} else if databaseServer.Spec.Type == "mongo" || databaseServer.Spec.Type == "mongodb" {
		sqlServer = &db.MongoServer{
			Username: databaseServer.Spec.Mongo.Username,
			Password: string(serverSecret.Data["password"]),
			Host:     databaseServer.Spec.Mongo.Host,
			Port:     databaseServer.Spec.Mongo.Port,
			Ssl:      databaseServer.Spec.Mongo.Ssl,
			Mongo: db.Mongo{
				Name:     database.Spec.Name,
				Username: username,
				Password: pass,
			},
		}
	}

	if msg, err := sqlServer.Connect(); err != nil {
		log.Error(err, msg)
		return ctrl.Result{}, err
	}
	database.Status.Connection = true
	if err := r.Status().Update(ctx, &database); err != nil {
		log.Error(err, "unable to update database status")
		return ctrl.Result{}, err
	}
	defer sqlServer.Disconnect()

	// If database shall be deleted with CR, add finalizer
	if database.Spec.ReclaimPolicy == "delete" && !containsString(database.ObjectMeta.Finalizers, finalizer) {
		database.ObjectMeta.Finalizers = append(database.ObjectMeta.Finalizers, finalizer)
		if err := r.Update(ctx, &database); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Finalize handler
	if !database.ObjectMeta.DeletionTimestamp.IsZero() && database.Spec.ReclaimPolicy == "delete" {
		log.Info("Database being finalized")

		if msg, err := sqlServer.DeleteDatabase(); err != nil {
			log.Info(msg, "err", err)
		}

		if msg, err := sqlServer.DeleteUser(); err != nil {
			log.Info(msg, "err", err)
		}

		if err := r.KubernetesClientset.CoreV1().Secrets(database.Spec.Secret.Namespace).Delete(database.Spec.Secret.Name, &metav1.DeleteOptions{}); err != nil {
			log.Info("unable to delete secret", "err", err)
		}

		// Remove finalizer to complete finazling
		database.ObjectMeta.Finalizers = removeString(database.ObjectMeta.Finalizers, finalizer)
		if err := r.Update(ctx, &database); err != nil {
			log.Error(err, "unable to update database resource")
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	if msg, err := sqlServer.CreateDatabase(); err != nil {
		log.Error(err, msg)
		return ctrl.Result{}, err
	}
	database.Status.CreatedDatabase = true
	if err := r.Status().Update(ctx, &database); err != nil {
		log.Error(err, "unable to update database status")
		return ctrl.Result{}, err
	}

	if msg, err := sqlServer.CreateUser(); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			log.Info("User already exists", "user", username)
		} else {
			log.Error(err, msg)
			return ctrl.Result{}, err
		}
	}
	database.Status.CreatedUser = true
	if err := r.Status().Update(ctx, &database); err != nil {
		log.Error(err, "unable to update database status")
		return ctrl.Result{}, err
	}

	if msg, err := sqlServer.GrantPermissions(); err != nil {
		log.Error(err, msg)
		return ctrl.Result{}, err
	}
	database.Status.GrantedPermissions = true
	if err := r.Status().Update(ctx, &database); err != nil {
		log.Error(err, "unable to update database status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&databasev1alpha1.Database{}).
		Complete(r)
}
