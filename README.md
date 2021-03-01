# Database Provisioning Controller PoC
A controller built on [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) automatically creating a user and database on a database server.

# Index
- [About](#about)
  - [Custom Resources](#custom-resources)
    -  [DatabaseServer](#databaseserver)
    -  [Database](#database)
    -  [More Examples](#more-examples)
- [Getting Started](#getting-started)
  - [Install controllers and CRDs using Helm](#install-controllers-and-crds-using-helm)
- [Usage](#usage)
  - [Creating a DatabaseServer resource](#creating-a-databaseserver-resource)
  - [Creating a Database resource](#creating-a-database-resource)
    - [Manually creating a Database resource](#manually-creating-a-database-resource)
    - [Creating a Database resource using Helm](#creating-a-database-resource-using-helm)
# About
Creates a new user and database on a database server.
Generates a secret containing the username and password used to connect to the database created for that spesific user. 

*Currently supports Postgresql, Mysql and Mongodb.*
## Custom Resources
### DatabaseServer
DatabaseServer contains info about where the database server is located and how to connect to it.

A yaml example of how to define a DatabaseServer:
```YAML
apiVersion: database.stacc.com/v1alpha1
kind: DatabaseServer
metadata:
  name: postgres-server
spec:
  type: postgres
  secret:
    name: postgres-server-secret
    namespace: default
  postgres:
    host: localhost
    username: postgres
    port: 5432
    sslmode: disable

```
- type: The type of database server being used. [postgres, mysql or mongo]
- secret: Secret where the password used to login is stored. [Must contain a field called "password"]
    name: Name of the secret
    namespace: Namespace where the secret is located
- postgres: Object with the name of type of database. [Must be either postgres, mysql, or mongo]
  - host: Hostname to the server
  - username: Username used to login (Must be a user with permission to create users and databases)
  - port: Port of the server
  - sslmode: Which SSL mode to use. Differs between postgres, mysql and mongo
    - Postgres: "sslmode": [disable, allow, prefer, require, verify-ca, verify-full]
    - Mysql: "ssl": [true, false]
    - Mongo: "ssl": [true, false]

### Database
Provides the name of the database and user to be created.

Example:
```YAML
apiVersion: database.stacc.com/v1alpha1
kind: Database
metadata:
  name: postgres-db
spec:
  name: postgres-db
  username: db-user
  reclaimPolicy: delete
  server:
    name: postgres-server
    namespace: default
  secret:
    name: postgres-db-secret
    namespace: default

```
- name: The name of the database
- username(Optional): The username to be associated with the database. If omitted will default to the name of the database.
- reclaimPolicy: What will happen with the user and database when this resource is deleted. [delete, retain]
- server: The DatabaseServer resource this database will be created on.
  - name: Name of the DatabaseServer.
  - namespace: The namespace the resource is located.
- secret: A secret will be created with fields "username" and "password", used to login to the new database.
  - name: The name of the secret.
  - namespace: In which namespace the secret will be stored.
  
### More examples
Examples for both resources made for all types of databases can be found [here](https://github.com/AuStien/database-provisioning-controller-poc/tree/main/config/samples).
- DatabaseServer
  - [Postgres](https://github.com/AuStien/database-provisioning-controller-poc/blob/main/config/samples/database_v1alpha1_databaseserver_postgres.yaml)
  - [Mysql](https://github.com/AuStien/database-provisioning-controller-poc/blob/main/config/samples/database_v1alpha1_databaseserver_mysql.yaml)
  - [Mongo](https://github.com/AuStien/database-provisioning-controller-poc/blob/main/config/samples/database_v1alpha1_databaseserver_mongo.yaml)
- Database
  - [Postgres](https://github.com/AuStien/database-provisioning-controller-poc/blob/main/config/samples/database_v1alpha1_database_postgres.yaml)
  - [Mysql](https://github.com/AuStien/database-provisioning-controller-poc/blob/main/config/samples/database_v1alpha1_database_mysql.yaml)
  - [Mongo](https://github.com/AuStien/database-provisioning-controller-poc/blob/main/config/samples/database_v1alpha1_database_mongo.yaml)
  
# Getting started
  
## Install controllers and CRDs using Helm
The easiest way to install the controller and CRDs is with [helm](https://helm.sh/).
```SHELL
helm install database-controller helm/database-controller
```

# Usage
Make sure you have a database server running with access to a user able to create both users and databases.
## Creating a DatabaseServer resource
1. Create a Secret containing the password used to login to the admin user.
  The easiest way to create a secret is by using kubectl. Here you can either enter the password directly or pass a file.
  ```shell
  kubectl create secret generic postgres-server-secret --from-literal=password="password"
  ```
  or
  ```shell
  kubectl create secret generic postgres-server-secret --from-file=password=./password.txt
  ```
2. Create a DatabaseServer resource and apply it to the cluster.
  ```YAML
  apiVersion: database.stacc.com/v1alpha1
  kind: DatabaseServer
  metadata:
    name: postgres-server
  spec:
    type: postgres
    secret:
      name: postgres-server-secret
      namespace: default
    postgres:
      host: localhost
      username: postgres
      port: 5432
      sslmode: disable
  ```

## Creating a Database resource
### Manually creating a Database resource
1. Create a Database resource and apply it to the cluster.
  ```YAML
  apiVersion: database.stacc.com/v1alpha1
  kind: Database
  metadata:
    name: postgres-db
  spec:
    name: postgres-db
    username: db-user
    reclaimPolicy: delete
    server:
      name: postgres-server
      namespace: default
    secret:
      name: postgres-db-secret
      namespace: default

  ```
2. Now a user and database is created. The username and password required to login is stored on the Secret specified in the Database resource.

### Creating a Database resource using Helm
With a bit of custom setup it is possible to make the process of creating a Database much easier. There are many ways of doing this. The following is just one example.

**In this case it is assumed that the DatabaseServer is already up and running.**

Add this to the chart template, in *database.yaml*.
```YAML
{{- if .Values.database.enabled -}}
{{- $name := include "common.name" . -}}
apiVersion: database.stacc.com/v1alpha1
kind: Database
metadata:
  name: {{ $name }}-db
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "common.labels" . | nindent 4 }}
spec:
  name: {{ .Values.database.name }}
  username: {{ .Values.database.username }}
  reclaimPolicy: {{ .Values.database.reclaimPolicy }}
  server:
    name: {{ .Values.database.server.name }}
    namespace: {{ .Values.database.server.namespace }}
  secret:
    name: {{ .Values.database.secret.name }}
    namespace: {{ .Values.database.secret.namespace }}
  
{{- end }}

```
Now the HelmReleases can be edited to contain the following values.
```YAML
database: 
  enabled: true
  name: postgres-db
  username: db-user
  reclaimPolicy: delete
  server:
    name: postgres-server
    namespace: default
  secret:
    name: postgres-db-secret
    namespace: default
    
```
