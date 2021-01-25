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
	"fmt"
	"strings"
	"time"

	databasev1alpha1 "flow.stacc.dev/database-provisioning-poc/api/v1alpha1"
	kubernetes "flow.stacc.dev/database-provisioning-poc/pkg/kubernetes"
	"github.com/go-logr/logr"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sethvargo/go-password/password"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DatabaseReconciler reconciles a Database object
type DatabaseReconciler struct {
	client.Client
	Log              logr.Logger
	Scheme           *runtime.Scheme
	KubernetesClient *kubernetes.Client
}

// +kubebuilder:rbac:groups=database.stacc.com,resources=databases,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=database.stacc.com,resources=databases/status,verbs=get;update;patch

func (r *DatabaseReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("database", req.NamespacedName)

	finalizer := "database.stacc.com/finalizer"

	// Get database resource
	var database databasev1alpha1.Database
	err := r.Get(ctx, req.NamespacedName, &database)
	if err != nil {
		log.Error(err, "uanble to get database resource")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Get database Server resource
	var databaseServer databasev1alpha1.DatabaseServer
	err = r.Get(ctx, client.ObjectKey{Namespace: database.Spec.Server.Namespace, Name: database.Spec.Server.Name}, &databaseServer)
	if err != nil {
		log.Error(err, "uanble to get databaseServer resource")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Stop reconsiling if database server is not connected
	if databaseServer.Status.Connected == false {
		log.Info("Database server not connected")
		return ctrl.Result{}, nil
	}

	// Get k8s clientset
	clientset, err := kubernetes.NewClientset(r.KubernetesClient)
	if err != nil {
		log.Error(err, "Error obtaining clientset")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Get secret with database server password
	serverSecret, err := clientset.CoreV1().Secrets(databaseServer.GetNamespace()).Get(databaseServer.Spec.SecretName, metav1.GetOptions{})
	if err != nil {
		log.Error(err, "Error obtaining secret")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Create url for inital connect to server
	url := fmt.Sprintf("postgresql://%s:%s@%s:%d/postgres", databaseServer.Spec.Postgres.Username, serverSecret.Data["password"], databaseServer.Spec.Postgres.Host, databaseServer.Spec.Postgres.Port)

	log.Info("Inital connect to database", "username", databaseServer.Spec.Postgres.Username, "database", "postgres", "host", databaseServer.Spec.Postgres.Host, "port", databaseServer.Spec.Postgres.Port)

	// Create connection to server
	dbpool, err := pgxpool.Connect(ctx, url)
	if err != nil {
		log.Info("Unable to connect to database")
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}
	defer dbpool.Close()

	// Get username, set to database name if not present
	username := database.Spec.Username
	if username == "" {
		username = database.Spec.Name
	}

	// Finalize handler
	if !database.ObjectMeta.DeletionTimestamp.IsZero() && database.Spec.Deletable {
		log.Info("Database being finalized")

		_, err = dbpool.Exec(ctx, fmt.Sprintf("DROP DATABASE \"%s\"", database.Spec.Name))
		if err != nil {
			log.Error(err, "unable to drop database in database server")
			return ctrl.Result{}, err
		}

		_, err = dbpool.Exec(ctx, fmt.Sprintf("DROP USER \"%s\"", username))
		if err != nil {
			log.Error(err, "unable to drop user in database server")
			return ctrl.Result{}, err
		}

		database.ObjectMeta.Finalizers = removeString(database.ObjectMeta.Finalizers, finalizer)
		if err := r.Update(ctx, &database); err != nil {
			log.Error(err, "unable to update database resource")
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	// Generate a password with length 48, 10 digits, allow uppercase, allow repeated chars
	password, err := password.Generate(48, 10, 0, false, true)
	if err != nil {
		log.Error(err, "unable to generate password")
		return ctrl.Result{}, err
	}

	// Check if database secret exists
	dbSecret, err := clientset.CoreV1().Secrets(database.Spec.Secret.Namespace).Get(database.Spec.Secret.Name, metav1.GetOptions{})
	if err != nil {
		// If error is other than "Not found" stop reconsiling
		if !errors.IsNotFound(err) {
			log.Error(err, "unable to create secret")
			return ctrl.Result{}, err
		}
		// Create database secret
		dbSecret = &coreV1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      database.Spec.Secret.Name,
				Namespace: database.Spec.Secret.Namespace,
			},
			Data: map[string][]byte{
				"username": []byte(username),
				"password": []byte(password),
			},
		}
		_, err := clientset.CoreV1().Secrets(database.Spec.Secret.Namespace).Create(dbSecret)
		if err != nil {
			log.Error(err, "unable to create secret")
			return ctrl.Result{}, err
		}
		// Secret already exists and needs to be updated
	} else {
		dbSecret.Data["username"] = []byte(username)
		dbSecret.Data["password"] = []byte(password)
		_, err := clientset.CoreV1().Secrets(database.Spec.Secret.Namespace).Update(dbSecret)
		if err != nil {
			log.Error(err, "unable to update secret")
			return ctrl.Result{}, err
		}
	}
	database.Status.Secret = "done"
	if err := r.Status().Update(ctx, &database); err != nil {
		log.Error(err, "unable to update database status")
		return ctrl.Result{}, err
	}

	// Check if user exists on server
	commandTag, err := dbpool.Exec(ctx, fmt.Sprintf("SELECT usename FROM pg_user WHERE usename='%s'", username))
	// If user doesn't exist create new
	if err != nil || commandTag.RowsAffected() == 0 {
		_, err = dbpool.Exec(ctx, fmt.Sprintf("CREATE USER \"%s\" WITH PASSWORD '%s'", username, password))
		if err != nil {
			log.Error(err, "unable to create role in database")
			return ctrl.Result{}, err
		}
		// If user exists update password
	} else {
		_, err = dbpool.Exec(ctx, fmt.Sprintf("ALTER USER \"%s\" WITH PASSWORD '%s'", username, password))
		if err != nil {
			log.Error(err, "unable to alter user in database")
			return ctrl.Result{}, err
		}
	}
	database.Status.User = "done"
	if err := r.Status().Update(ctx, &database); err != nil {
		log.Error(err, "unable to update database status")
		return ctrl.Result{}, err
	}

	// Try to create database
	_, err = dbpool.Exec(ctx, fmt.Sprintf("CREATE DATABASE \"%s\" TEMPLATE \"template0\"", database.Spec.Name))
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			log.Info("Database already exisis")
		} else {
			log.Error(err, "unable to create database in database server")
			return ctrl.Result{}, err
		}
	}
	database.Status.DB = "done"
	if err := r.Status().Update(ctx, &database); err != nil {
		log.Error(err, "unable to update database status")
		return ctrl.Result{}, err
	}

	// Grant permissions to user
	_, err = dbpool.Exec(ctx, fmt.Sprintf("GRANT ALL ON DATABASE \"%s\" TO \"%s\"", database.Spec.Name, username))
	if err != nil {
		log.Error(err, "unable to grant permissions in database")
		return ctrl.Result{}, err
	}
	database.Status.Permissions = "done"
	if err := r.Status().Update(ctx, &database); err != nil {
		log.Error(err, "unable to update database status")
		return ctrl.Result{}, err
	}

	// Test if connection to new database works
	newURL := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", username, password, databaseServer.Spec.Postgres.Host, databaseServer.Spec.Postgres.Port, database.Spec.Name)

	log.Info("Connecting to new database", "username", username, "database", database.Spec.Name, "host", databaseServer.Spec.Postgres.Host, "port", databaseServer.Spec.Postgres.Port)

	newDBpool, err := pgxpool.Connect(ctx, newURL)
	if err != nil {
		log.Info("Unable to connect to new database", "err", err)
		database.Status.Connection = false
		if err := r.Status().Update(ctx, &database); err != nil {
			log.Error(err, "unable to update database status")
			return ctrl.Result{}, err
		}
	} else {
		database.Status.Connection = true
		if err := r.Status().Update(ctx, &database); err != nil {
			log.Error(err, "unable to update database status")
			return ctrl.Result{}, err
		}
	}
	defer newDBpool.Close()

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
