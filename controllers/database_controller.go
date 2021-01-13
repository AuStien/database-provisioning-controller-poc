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

	var database databasev1alpha1.Database
	err := r.Get(ctx, req.NamespacedName, &database)
	if err != nil {
		log.Error(err, "uanble to get database resource")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var databaseServer databasev1alpha1.DatabaseServer
	err = r.Get(ctx, client.ObjectKey{Namespace: database.Spec.DatabaseServerNamespace, Name: database.Spec.DatabaseServerName}, &databaseServer)
	if err != nil {
		log.Error(err, "uanble to get databaseServer resource")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if databaseServer.Status.Connected == false {
		log.Info("Database server not connected")
		return ctrl.Result{}, nil
	}

	clientset, err := kubernetes.NewClientset(r.KubernetesClient)
	if err != nil {
		log.Error(err, "Error obtaining clientset")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	serverSecret, err := clientset.CoreV1().Secrets(databaseServer.GetNamespace()).Get(databaseServer.Spec.SecretName, metav1.GetOptions{})
	if err != nil {
		log.Error(err, "Error obtaining secret")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	url := fmt.Sprintf("postgresql://%s:%s@%s:%d/postgres", databaseServer.Spec.Postgres.Username, serverSecret.Data["password"], databaseServer.Spec.Postgres.Host, databaseServer.Spec.Postgres.Port)

	dbpool, err := pgxpool.Connect(ctx, url)
	if err != nil {
		log.Info("Unable to connect to database")
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}
	defer dbpool.Close()

	username := database.Spec.Username
	if username == "" {
		username = database.Spec.Name
	}

	// Generate a password with length 48, 10 digits, allow uppercase, allow repeated chars
	password, err := password.Generate(48, 10, 0, false, true)
	if err != nil {
		log.Error(err, "unable to generate password")
		return ctrl.Result{}, err
	}

	dbSecret, err := clientset.CoreV1().Secrets(database.Spec.SecretNamespace).Get(database.Spec.SecretName, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			log.Error(err, "unable to create secret")
			return ctrl.Result{}, err
		}
		dbSecret = &coreV1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      database.Spec.SecretName,
				Namespace: database.Spec.SecretNamespace,
			},
			Data: map[string][]byte{
				"username": []byte(username),
				"password": []byte(password),
			},
		}
		_, err := clientset.CoreV1().Secrets(database.Spec.SecretNamespace).Create(dbSecret)
		if err != nil {
			log.Error(err, "unable to create secret")
			return ctrl.Result{}, err
		}
	} else {
		dbSecret.Data["username"] = []byte(username)
		dbSecret.Data["password"] = []byte(password)
		_, err := clientset.CoreV1().Secrets(database.Spec.SecretNamespace).Update(dbSecret)
		if err != nil {
			log.Error(err, "unable to update secret")
			return ctrl.Result{}, err
		}
	}

	_, err = dbpool.Exec(ctx, fmt.Sprintf("CREATE USER \"%s\" WITH PASSWORD '%s'", username, password))
	if err != nil {
		log.Error(err, "unable to create role in database")
		return ctrl.Result{}, err
	}

	_, err = dbpool.Exec(ctx, fmt.Sprintf("CREATE DATABASE \"%s\" TEMPLATE \"template0\"", database.Spec.Name))
	if err != nil {
		log.Error(err, "unable to create database in database server")
		return ctrl.Result{}, err
	}

	_, err = dbpool.Exec(ctx, fmt.Sprintf("GRANT ALL ON DATABASE \"%s\" TO \"%s\"", database.Spec.Name, username))
	if err != nil {
		log.Error(err, "unable to grant permissions in database")
		return ctrl.Result{}, err
	}

	log.Info("Database setup successfull")

	return ctrl.Result{}, nil
}

func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&databasev1alpha1.Database{}).
		Complete(r)
}
