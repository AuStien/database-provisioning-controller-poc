# Database Provisioning Controller PoC
A controller built on [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) automatically creating a user and database on a database server.

# About
Creates a new user and database on a database server.
Generates a secret containing the username and password used to connect to the database created for that spesific user. 

Currently supports Postgresql, Mysql and Mongodb.
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
  - sslmode: Which SSL mode to use. Differs between postgres, mysql and mongo. See below

Postgres: "sslmode": [disable, allow, prefer, require, verify-ca, verify-full]
Mysql: "ssl": [true, false]
Mongo: "ssl": [true, false]

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
- server: Points to the DatabaseServer resource this database shall be hosted on.
  - name: Name of the resource.
  - namespace: The namespace the resource currently is.
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
  
## Helm
The easiest way to install the controller and CRDs is with [helm](https://helm.sh/).
```SHELL
helm install database-controller helm/database-controller
```

